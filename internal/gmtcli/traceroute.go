package gmtcli

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/proto"
)

// Traceroute is the result of tracing the route to a node: the hops the
// request took to reach it and the hops the reply took back, each link with
// the SNR the receiving node measured.
type Traceroute struct {
	From    TracerouteNode  `json:"from"`
	To      TracerouteNode  `json:"to"`
	Towards []TracerouteHop `json:"towards"`
	Back    []TracerouteHop `json:"back,omitempty"`
}

// TracerouteNode identifies a node on the route.
type TracerouteNode struct {
	ID        string `json:"node_id"`
	LongName  string `json:"longname,omitempty"`
	ShortName string `json:"shortname,omitempty"`
}

// TracerouteHop is one link: the node reached and the SNR it heard the
// previous node with. SNR is nil when the node did not report it.
type TracerouteHop struct {
	TracerouteNode
	SNR *float64 `json:"snr_db,omitempty"`
}

// unknownSNR is the value the firmware uses in a RouteDiscovery for a link
// whose SNR was not recorded (quarter-dB units, int8 minimum).
const unknownSNR = -128

// traceroute is the action for the traceroute command.
func traceroute(c *cli.Context) error {
	if c.NArg() != 1 {
		return cli.Exit("usage: traceroute <node>   (a node number, !xxxxxxxx id, or node name)", 2)
	}

	radio := getRadio(c)
	defer radio.Close()

	myNode, nodes, err := radioNodes(radio)
	if err != nil {
		return cli.Exit(err, 1)
	}
	byNumber := make(map[uint32]MeshNode, len(nodes))
	for _, node := range nodes {
		byNumber[node.NodeNumber] = node
	}

	dest, err := lookupNode(c.Args().First(), byNumber)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if dest == myNode {
		return cli.Exit("that is this radio", 1)
	}

	id := rand.Uint32()
	for id == 0 {
		id = rand.Uint32()
	}
	request, err := tracerouteRequest(dest, id)
	if err != nil {
		return cli.Exit(err, 1)
	}
	if err := radio.SendPacket(request); err != nil {
		return cli.Exit(fmt.Sprintf("sending traceroute: %v", err), 1)
	}

	timeout := c.Duration("timeout")
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		responses, err := radio.ReadResponse(false)
		if err != nil {
			return cli.Exit(fmt.Sprintf("reading from radio: %v", err), 1)
		}
		for _, response := range responses {
			packet := response.GetPacket()
			decoded := packet.GetDecoded()
			if decoded == nil || decoded.GetRequestId() != id {
				continue
			}
			switch decoded.GetPortnum() {
			case gomeshproto.PortNum_TRACEROUTE_APP:
				var discovery gomeshproto.RouteDiscovery
				if err := proto.Unmarshal(decoded.GetPayload(), &discovery); err != nil {
					return cli.Exit(fmt.Sprintf("unreadable traceroute reply: %v", err), 1)
				}
				result := buildTraceroute(myNode, dest, &discovery, byNumber)
				if c.Bool("json") {
					out, err := json.Marshal(result)
					if err != nil {
						return err
					}
					fmt.Println(string(out))
				} else {
					fmt.Print(formatTraceroute(result))
				}
				return nil
			case gomeshproto.PortNum_ROUTING_APP:
				var routing gomeshproto.Routing
				if err := proto.Unmarshal(decoded.GetPayload(), &routing); err == nil {
					if reason := routing.GetErrorReason(); reason != gomeshproto.Routing_NONE {
						return cli.Exit(fmt.Sprintf("traceroute failed: %s", reason), 1)
					}
				}
			}
		}
	}
	return cli.Exit(fmt.Sprintf("no traceroute reply within %s", timeout), 1)
}

// tracerouteRequest builds the ToRadio message carrying a traceroute request
// to dest with the given packet id.
func tracerouteRequest(dest, id uint32) ([]byte, error) {
	payload, err := proto.Marshal(&gomeshproto.RouteDiscovery{})
	if err != nil {
		return nil, err
	}
	return proto.Marshal(&gomeshproto.ToRadio{PayloadVariant: &gomeshproto.ToRadio_Packet{Packet: &gomeshproto.MeshPacket{
		To: dest,
		Id: id,
		PayloadVariant: &gomeshproto.MeshPacket_Decoded{Decoded: &gomeshproto.Data{
			Portnum:      gomeshproto.PortNum_TRACEROUTE_APP,
			Payload:      payload,
			WantResponse: true,
		}},
	}}})
}

// buildTraceroute turns a RouteDiscovery reply into a Traceroute. The reply's
// route lists the relays between us and the destination; snr_towards has one
// entry per link including the final one into the destination. route_back
// and snr_back describe the reply's path from the destination to us.
func buildTraceroute(from, to uint32, discovery *gomeshproto.RouteDiscovery, nodes map[uint32]MeshNode) Traceroute {
	name := func(num uint32) TracerouteNode {
		if num == broadcastNodeNumber {
			// The firmware records a relay it could not identify as the
			// broadcast address.
			return TracerouteNode{ID: nodeIDFromNumber(num), LongName: "unknown relay"}
		}
		node, ok := nodes[num]
		if !ok {
			return TracerouteNode{ID: nodeIDFromNumber(num)}
		}
		return TracerouteNode{ID: node.ID, LongName: node.LongName, ShortName: node.ShortName}
	}
	hops := func(route []uint32, snrs []int32, last uint32) []TracerouteHop {
		path := append(append([]uint32{}, route...), last)
		out := make([]TracerouteHop, 0, len(path))
		for i, num := range path {
			hop := TracerouteHop{TracerouteNode: name(num)}
			if i < len(snrs) && snrs[i] != unknownSNR {
				snr := float64(snrs[i]) / 4 // quarter-dB units
				hop.SNR = &snr
			}
			out = append(out, hop)
		}
		return out
	}

	result := Traceroute{
		From:    name(from),
		To:      name(to),
		Towards: hops(discovery.GetRoute(), discovery.GetSnrTowards(), to),
	}
	if len(discovery.GetRouteBack()) > 0 || len(discovery.GetSnrBack()) > 0 {
		result.Back = hops(discovery.GetRouteBack(), discovery.GetSnrBack(), from)
	}
	return result
}

// formatTraceroute renders a Traceroute for a person.
func formatTraceroute(t Traceroute) string {
	label := func(n TracerouteNode) string {
		if n.LongName != "" {
			return fmt.Sprintf("%s (%s)", n.LongName, n.ID)
		}
		return n.ID
	}
	path := func(start TracerouteNode, hops []TracerouteHop) string {
		var b strings.Builder
		fmt.Fprintf(&b, "  %s\n", label(start))
		for _, hop := range hops {
			if hop.SNR != nil {
				fmt.Fprintf(&b, "    %+.2f dB\n", *hop.SNR)
			} else {
				b.WriteString("    ? dB\n")
			}
			fmt.Fprintf(&b, "  %s\n", label(hop.TracerouteNode))
		}
		return b.String()
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Route to %s, %d hop(s):\n", label(t.To), len(t.Towards))
	b.WriteString(path(t.From, t.Towards))
	if len(t.Back) > 0 {
		fmt.Fprintf(&b, "Route back, %d hop(s):\n", len(t.Back))
		b.WriteString(path(t.To, t.Back))
	}
	return b.String()
}
