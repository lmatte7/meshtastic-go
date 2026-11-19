// Command weatherbot answers "!weather" on the mesh with the current weather
// for where the radio is, looked up from wttr.in.
//
// The radio's own GPS position is used when it has one. Otherwise the
// position is estimated as the mean of the positions of nodes the radio has
// heard over the air recently (no MQTT, at most a few hops away). The reply
// goes back the way the request came: on the channel, or directly for a
// direct message. "!weather <place>" looks up a named place instead.
//
// The bot answers at most one request per interval (default five minutes)
// across the whole mesh, so it makes at most one wttr.in request and sends at
// most one reply per interval however many nodes ask.
//
//	go run ./examples/weatherbot -port /dev/cu.usbmodem143101
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

// node is the part of a stream node record this bot uses.
type node struct {
	NodeID              string  `json:"node_id"`
	LongName            string  `json:"longname"`
	Latitude            float64 `json:"latitude_degrees"`
	Longitude           float64 `json:"longitude_degrees"`
	HopsAway            int     `json:"hops_away"`
	ViaMQTT             bool    `json:"via_mqtt"`
	LastHeardSecondsAgo int     `json:"last_heard_seconds_ago"`
}

func (n node) hasPosition() bool { return n.Latitude != 0 || n.Longitude != 0 }

// event holds the parts of a stream line this bot looks at.
type event struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Text     string `json:"text"`
	Direct   bool   `json:"direct"`
	Outgoing bool   `json:"outgoing"`
	Implicit bool   `json:"implicit"`
	Error    string `json:"error"`
	From     node   `json:"from"`
	Reply    *struct {
		Direct  map[string]interface{} `json:"direct"`
		Channel map[string]interface{} `json:"channel"`
	} `json:"reply"`
	// Filled in on "nodes" events.
	MyNode *node  `json:"my_node"`
	Nodes  []node `json:"nodes"`
}

// request is a !weather waiting for the node list.
type request struct {
	reply map[string]interface{}
	place string
}

type bot struct {
	commands io.Writer
	log      *log.Logger
	now      func() time.Time
	weather  func(location string) (string, error)
	maxHops  int
	maxAge   time.Duration
	// interval throttles the bot as a whole: at most one weather lookup and
	// one reply to the mesh per interval, no matter how many nodes ask, so a
	// hostile mesh cannot turn the bot into a flood source.
	interval  time.Duration
	lastReply time.Time
	pending   map[string]request
	counter   int
}

func newBot(commands io.Writer, logger *log.Logger) *bot {
	return &bot{
		commands: commands,
		log:      logger,
		now:      time.Now,
		weather:  wttr("u"),
		maxHops:  2,
		maxAge:   24 * time.Hour,
		interval: 5 * time.Minute,
		pending:  make(map[string]request),
	}
}

// handle processes one line from the stream.
func (b *bot) handle(line string) error {
	var ev event
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return fmt.Errorf("unreadable line from stream: %w", err)
	}

	switch ev.Type {
	case "ready":
		b.log.Printf("connected, waiting for !weather")
	case "message":
		return b.onMessage(ev)
	case "nodes":
		return b.onNodes(ev)
	case "sent", "queued":
		b.log.Printf("%s: %s", ev.ID, ev.Type)
	case "delivered":
		if ev.Implicit {
			b.log.Printf("%s: relayed by a neighbour", ev.ID)
		} else {
			b.log.Printf("%s: delivered", ev.ID)
		}
	case "failed", "error":
		b.log.Printf("%s: %s: %s", ev.ID, ev.Type, ev.Error)
	}
	return nil
}

// onMessage reacts to "!weather [place]" by asking the stream for the node
// list; the lookup happens when the list arrives.
func (b *bot) onMessage(ev event) error {
	if ev.Outgoing || ev.Reply == nil {
		return nil
	}
	fields := strings.Fields(ev.Text)
	if len(fields) == 0 || !strings.EqualFold(fields[0], "!weather") {
		return nil
	}

	// One request per interval, bot-wide. The slot is taken now, before the
	// lookup, so a burst of requests cannot all get through.
	if since := b.now().Sub(b.lastReply); since < b.interval {
		b.log.Printf("ignoring !weather from %s: last reply %s ago, next allowed in %s", ev.From.NodeID, since.Round(time.Second), (b.interval - since).Round(time.Second))
		return nil
	}
	b.lastReply = b.now()

	reply := ev.Reply.Channel
	if ev.Direct {
		reply = ev.Reply.Direct
	}
	b.counter++
	id := fmt.Sprintf("weather-%d", b.counter)
	b.pending[id] = request{reply: reply, place: strings.Join(fields[1:], " ")}
	b.log.Printf("%s: !weather from %s", id, ev.From.NodeID)
	return b.write(map[string]interface{}{"id": id, "command": "nodes"})
}

// onNodes finishes a pending request once the node list is available.
func (b *bot) onNodes(ev event) error {
	req, ok := b.pending[ev.ID]
	if !ok {
		return nil
	}
	delete(b.pending, ev.ID)

	location, source := req.place, "for "+req.place
	if location == "" {
		var ok bool
		location, source, ok = b.locate(ev.MyNode, ev.Nodes)
		if !ok {
			return b.send(req.reply, ev.ID, "Sorry, I don't know where I am: no GPS and no positioned nodes nearby.")
		}
	}

	report, err := b.weather(location)
	if err != nil {
		b.log.Printf("%s: weather lookup failed: %v", ev.ID, err)
		return b.send(req.reply, ev.ID, "Weather lookup failed, try again later.")
	}
	return b.send(req.reply, ev.ID, fmt.Sprintf("%s (%s)", report, source))
}

// locate returns the radio's position as "lat,lon" and a note on where it
// came from: the radio's GPS, or the mean of nearby nodes heard over the air.
func (b *bot) locate(me *node, nodes []node) (location, source string, ok bool) {
	if me != nil && me.hasPosition() {
		return coords(me.Latitude, me.Longitude), "at the radio's GPS position", true
	}

	// Candidates: positioned nodes heard over the air within the hop limit.
	var candidates, recent []node
	for _, n := range nodes {
		if !n.hasPosition() || n.ViaMQTT || n.HopsAway > b.maxHops {
			continue
		}
		if me != nil && n.NodeID == me.NodeID {
			continue
		}
		candidates = append(candidates, n)
		if b.maxAge <= 0 || time.Duration(n.LastHeardSecondsAgo)*time.Second <= b.maxAge {
			recent = append(recent, n)
		}
	}

	// Prefer recently heard nodes, but radios without a time source stamp
	// everything with a stale clock, so fall back to all candidates rather
	// than refusing when too few look recent.
	chosen, note := recent, "estimated from %d nearby nodes"
	if len(recent) < minRecentNodes && len(candidates) > 0 {
		chosen, note = candidates, "estimated from %d nearby nodes, some not heard recently"
	}
	if len(chosen) == 0 {
		return "", "", false
	}

	var lat, lon float64
	for _, n := range chosen {
		lat += n.Latitude
		lon += n.Longitude
	}
	count := float64(len(chosen))
	return coords(lat/count, lon/count), fmt.Sprintf(note, len(chosen)), true
}

// minRecentNodes is how many recently heard positioned nodes are needed before
// older ones are ignored.
const minRecentNodes = 3

func coords(lat, lon float64) string {
	return fmt.Sprintf("%.3f,%.3f", lat, lon)
}

// send writes a reply command built from a reply template.
func (b *bot) send(template map[string]interface{}, id, text string) error {
	command := make(map[string]interface{}, len(template)+2)
	for k, v := range template {
		command[k] = v
	}
	command["id"] = id
	command["text"] = text
	return b.write(command)
}

func (b *bot) write(command map[string]interface{}) error {
	line, err := json.Marshal(command)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(b.commands, "%s\n", line)
	return err
}

// wttr returns a lookup function for wttr.in. units is "u" for US customary
// or "m" for metric.
func wttr(units string) func(string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	return func(location string) (string, error) {
		format := "%c %C %t (feels %f), wind %w, humidity %h"
		u := "https://wttr.in/" + url.PathEscape(location) + "?format=" + url.QueryEscape(format) + "&" + units
		req, err := http.NewRequest(http.MethodGet, u, nil)
		if err != nil {
			return "", err
		}
		// wttr.in returns plain text for command-line user agents.
		req.Header.Set("User-Agent", "curl/8.0 (meshtastic weatherbot)")
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("wttr.in returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}
		return strings.Join(strings.Fields(string(body)), " "), nil
	}
}

func main() {
	port := flag.String("port", "", "serial port or IP address of the radio; defaults to MESHTASTIC_RADIO_DEV or, on macOS, the only attached USB radio")
	bin := flag.String("bin", "meshtastic-go", "path to the meshtastic-go binary")
	units := flag.String("units", "u", "wttr.in units: u (US customary) or m (metric)")
	maxHops := flag.Int("max-hops", 2, "only use nodes at most this many hops away to estimate the position")
	maxAge := flag.Duration("max-age", 24*time.Hour, "only use nodes heard within this long to estimate the position")
	interval := flag.Duration("interval", 5*time.Minute, "answer at most one !weather (one lookup, one reply) per this interval")
	flag.Parse()

	// With no -port, meshtastic-go resolves the port itself from
	// MESHTASTIC_RADIO_DEV or the attached devices.
	args := []string{"stream", "--commands"}
	if *port != "" {
		args = append([]string{"--port", *port}, args...)
	}
	cmd := exec.Command(*bin, args...)
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatalf("starting %s: %v", *bin, err)
	}

	// Closing the stream's stdin asks it to exit cleanly.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-stop
		stdin.Close()
	}()

	b := newBot(stdin, log.New(os.Stderr, "weatherbot: ", log.LstdFlags))
	b.weather = wttr(*units)
	b.maxHops = *maxHops
	b.maxAge = *maxAge
	b.interval = *interval

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		if err := b.handle(scanner.Text()); err != nil {
			b.log.Print(err)
		}
	}
	if err := scanner.Err(); err != nil {
		b.log.Printf("reading from meshtastic-go: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Fatalf("meshtastic-go exited: %v", err)
	}
}
