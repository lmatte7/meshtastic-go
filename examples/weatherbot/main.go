// Command weatherbot answers "!weather" on the mesh with the current weather
// for where the radio is, looked up from wttr.in.
//
// The radio's own GPS position is used when it has one. Otherwise the
// position is estimated as the mean of the positions of nodes heard over the
// air (no MQTT, at most a few hops away), preferring ones heard recently. The
// reply
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
	"maps"
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
	HopsAway            *int    `json:"hops_away"` // nil when the radio did not say
	ViaMQTT             bool    `json:"via_mqtt"`
	LastHeard           string  `json:"last_heard"` // empty if the stream has not heard the node
	LastHeardSecondsAgo int     `json:"last_heard_seconds_ago"`
}

func (n node) hasPosition() bool { return n.Latitude != 0 || n.Longitude != 0 }

// eventID is an event's id, which is a number on message and packet lines
// (the mesh packet id) and a string on the results of our own commands.
type eventID string

func (id *eventID) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err == nil {
		*id = eventID(s)
		return nil
	}
	*id = eventID(string(b)) // a number: keep its digits
	return nil
}

// event holds the parts of a stream line this bot looks at.
type event struct {
	Type     string  `json:"type"`
	ID       eventID `json:"id"`
	Text     string  `json:"text"`
	Direct   bool    `json:"direct"`
	Outgoing bool    `json:"outgoing"`
	Implicit bool    `json:"implicit"`
	Error    string  `json:"error"`
	From     node    `json:"from"`
	Reply    *struct {
		Direct  map[string]any `json:"direct"`
		Channel map[string]any `json:"channel"`
	} `json:"reply"`
	// Filled in on "nodes" events.
	MyNode *node  `json:"my_node"`
	Nodes  []node `json:"nodes"`
}

// request is a !weather waiting for the node list.
type request struct {
	reply map[string]any
	place string
}

type bot struct {
	commands io.Writer
	log      *log.Logger
	debug    bool
	now      func() time.Time
	weather  func(location string) (string, error)
	maxHops  int
	maxAge   time.Duration
	// interval throttles the bot as a whole: at most one weather lookup and
	// one reply to the mesh per interval, no matter how many nodes ask, so a
	// hostile mesh cannot turn the bot into a flood source.
	interval  time.Duration
	lastReply time.Time
	// noticeSent records whether a throttled sender has been told the
	// cooldown in the current period; only one such notice goes out per
	// period, and further throttled requests are ignored silently.
	noticeSent bool
	pending    map[string]request
	counter    int
}

func newBot(commands io.Writer, logger *log.Logger) *bot {
	return &bot{
		commands: commands,
		log:      logger,
		now:      time.Now,
		weather:  wttr(),
		maxHops:  2,
		maxAge:   24 * time.Hour,
		interval: 5 * time.Minute,
		pending:  make(map[string]request),
	}
}

// debugf logs only when -debug is on.
func (b *bot) debugf(format string, args ...any) {
	if b.debug {
		b.log.Printf("debug: "+format, args...)
	}
}

// handle processes one line from the stream.
func (b *bot) handle(line string) error {
	var ev event
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return fmt.Errorf("unreadable line from stream: %w: %.200s", err, line)
	}
	b.debugf("recv %s id=%s from=%s direct=%v outgoing=%v text=%q", ev.Type, ev.ID, ev.From.NodeID, ev.Direct, ev.Outgoing, ev.Text)

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
	if ev.Outgoing {
		b.debugf("ignoring our own outgoing message")
		return nil
	}
	if ev.Reply == nil {
		b.debugf("ignoring message without reply templates")
		return nil
	}
	fields := strings.Fields(ev.Text)
	if len(fields) == 0 || !strings.EqualFold(fields[0], "!weather") {
		b.debugf("not a !weather request")
		return nil
	}

	// One request per interval, bot-wide. The slot is taken now, before the
	// lookup, so a burst of requests cannot all get through. The first
	// sender turned away in a period is told the cooldown, by direct message
	// even if they asked on a channel; anyone after that is ignored silently.
	if since := b.now().Sub(b.lastReply); since < b.interval {
		b.log.Printf("throttled !weather from %s: last reply %s ago", ev.From.NodeID, since.Round(time.Second))
		if b.noticeSent {
			b.debugf("cooldown notice already sent this period; staying silent")
			return nil
		}
		b.noticeSent = true
		b.counter++
		return b.send(ev.Reply.Direct, fmt.Sprintf("cooldown-%d", b.counter),
			fmt.Sprintf("!weather is available once every %s.", b.interval))
	}
	b.lastReply = b.now()
	b.noticeSent = false

	reply := ev.Reply.Channel
	if ev.Direct {
		reply = ev.Reply.Direct
	}
	b.counter++
	id := fmt.Sprintf("weather-%d", b.counter)
	b.pending[id] = request{reply: reply, place: strings.Join(fields[1:], " ")}
	b.log.Printf("%s: !weather from %s", id, ev.From.NodeID)
	return b.write(map[string]any{"id": id, "command": "nodes"})
}

// onNodes finishes a pending request once the node list is available.
func (b *bot) onNodes(ev event) error {
	req, ok := b.pending[string(ev.ID)]
	if !ok {
		b.debugf("nodes event %s is not for a pending request", ev.ID)
		return nil
	}
	delete(b.pending, string(ev.ID))
	b.debugf("nodes event %s: %d nodes, my_node position %v,%v", ev.ID, len(ev.Nodes), ev.MyNode != nil && ev.MyNode.Latitude != 0, ev.MyNode != nil && ev.MyNode.Longitude != 0)

	location, source := req.place, "for "+req.place
	if location == "" {
		var ok bool
		location, source, ok = b.locate(ev.MyNode, ev.Nodes)
		if !ok {
			return b.send(req.reply, string(ev.ID), "Sorry, I don't know where I am: no GPS and no positioned nodes nearby.")
		}
	}

	// The reply is the weather and nothing else: where the position came
	// from is logged here and never sent to the mesh.
	b.debugf("looking up weather for %q (%s)", location, source)
	report, err := b.weather(location)
	if err != nil {
		b.log.Printf("%s: weather lookup failed: %v", ev.ID, err)
		return b.send(req.reply, string(ev.ID), "Weather lookup failed, try again later.")
	}
	return b.send(req.reply, string(ev.ID), report)
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
		// A node whose hop count the radio did not report could be anywhere,
		// so it cannot be used to estimate where we are.
		if !n.hasPosition() || n.ViaMQTT || n.HopsAway == nil || *n.HopsAway > b.maxHops {
			continue
		}
		if me != nil && n.NodeID == me.NodeID {
			continue
		}
		candidates = append(candidates, n)
		heard := n.LastHeard != "" // the stream has received a packet from it
		if b.maxAge <= 0 || (heard && time.Duration(n.LastHeardSecondsAgo)*time.Second <= b.maxAge) {
			recent = append(recent, n)
		}
	}

	// Prefer nodes heard recently, but right after start-up few nodes have
	// been heard yet, so fall back to all candidates rather than refusing
	// when too few are recent.
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
func (b *bot) send(template map[string]any, id, text string) error {
	command := make(map[string]any, len(template)+2)
	maps.Copy(command, template)
	command["id"] = id
	command["text"] = text
	return b.write(command)
}

func (b *bot) write(command map[string]any) error {
	line, err := json.Marshal(command)
	if err != nil {
		return err
	}
	b.debugf("send %s", line)
	_, err = fmt.Fprintf(b.commands, "%s\n", line)
	return err
}

// wttr returns a lookup function for wttr.in, always in metric units.
// wttr.in occasionally drops a connection, so a lookup is retried a few
// times before it is reported as failed.
func wttr() func(string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	format := "%c %C %t (feels %f), wind %w, humidity %h"
	fetch := func(location string) (string, error) {
		u := "https://wttr.in/" + url.PathEscape(location) + "?format=" + url.QueryEscape(format) + "&m"
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
	return func(location string) (string, error) {
		var err error
		for attempt := 1; attempt <= 3; attempt++ {
			var report string
			if report, err = fetch(location); err == nil {
				return report, nil
			}
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
		return "", fmt.Errorf("after 3 attempts: %w", err)
	}
}

func main() {
	port := flag.String("port", "", "serial port or IP address of the radio; defaults to MESHTASTIC_RADIO_DEV or, on macOS, the only attached USB radio")
	bin := flag.String("bin", "meshtastic-go", "path to the meshtastic-go binary")
	maxHops := flag.Int("max-hops", 2, "only use nodes at most this many hops away to estimate the position")
	maxAge := flag.Duration("max-age", 24*time.Hour, "only use nodes heard within this long to estimate the position")
	interval := flag.Duration("interval", 5*time.Minute, "answer at most one !weather (one lookup, one reply) per this interval")
	debug := flag.Bool("debug", false, "log every line received from the stream, every command sent, and why messages are ignored")
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
	b.maxHops = *maxHops
	b.maxAge = *maxAge
	b.interval = *interval
	b.debug = *debug

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
