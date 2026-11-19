// Command pingbot is the smallest useful bot built on
// "meshtastic-go stream --commands": it answers "!ping" with "pong".
//
// It runs meshtastic-go as a subprocess, reads the JSON lines it prints, and
// writes reply commands back to its stdin. A message that arrived on a
// channel is answered on that channel; a direct message is answered
// directly. Delivery events for the replies are logged to stderr.
//
//	go run ./examples/pingbot -port /dev/cu.usbmodem143101
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// event holds the parts of a stream line this bot looks at. Unknown fields
// are ignored, so the bot keeps working as the stream grows new ones.
type event struct {
	Type     string `json:"type"`
	ID       string `json:"id"`
	Text     string `json:"text"`
	Direct   bool   `json:"direct"`
	Outgoing bool   `json:"outgoing"`
	Implicit bool   `json:"implicit"`
	Error    string `json:"error"`
	From     struct {
		NodeID   string `json:"node_id"`
		LongName string `json:"longname"`
	} `json:"from"`
	// Reply carries the ready-made command templates from the stream.
	Reply *struct {
		Direct  map[string]interface{} `json:"direct"`
		Channel map[string]interface{} `json:"channel"`
	} `json:"reply"`
}

// bot reacts to stream lines and writes commands to the stream's stdin.
type bot struct {
	commands io.Writer
	log      *log.Logger
	replies  int
}

// handle processes one line from the stream.
func (b *bot) handle(line string) error {
	var ev event
	if err := json.Unmarshal([]byte(line), &ev); err != nil {
		return fmt.Errorf("unreadable line from stream: %w", err)
	}

	switch ev.Type {
	case "ready":
		b.log.Printf("connected, waiting for !ping")
	case "message":
		if ev.Outgoing || !strings.EqualFold(strings.TrimSpace(ev.Text), "!ping") {
			return nil
		}
		return b.pong(ev)
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

// pong answers a !ping the same way it arrived: directly for a direct
// message, on the channel otherwise.
func (b *bot) pong(ev event) error {
	if ev.Reply == nil {
		return nil
	}
	command := ev.Reply.Channel
	how := "on the channel"
	if ev.Direct {
		command = ev.Reply.Direct
		how = "directly"
	}

	b.replies++
	command["id"] = fmt.Sprintf("pong-%d", b.replies)
	command["text"] = "pong"

	line, err := json.Marshal(command)
	if err != nil {
		return err
	}
	who := ev.From.LongName
	if who == "" {
		who = ev.From.NodeID
	}
	b.log.Printf("%s: !ping from %s, answering %s", command["id"], who, how)
	_, err = fmt.Fprintf(b.commands, "%s\n", line)
	return err
}

func main() {
	port := flag.String("port", "", "serial port or IP address of the radio (required)")
	bin := flag.String("bin", "meshtastic-go", "path to the meshtastic-go binary")
	flag.Parse()
	if *port == "" {
		flag.Usage()
		os.Exit(2)
	}

	cmd := exec.Command(*bin, "--port", *port, "stream", "--commands")
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

	b := &bot{commands: stdin, log: log.New(os.Stderr, "pingbot: ", log.LstdFlags)}
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
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
