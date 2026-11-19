package gmtcli

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"github.com/urfave/cli/v2"
)

// locateMaxAge is how recently a node must have been heard to help estimate
// the radio's position.
const locateMaxAge = 15 * time.Minute

// Location is where the radio is, either from its own GPS or estimated from
// the nodes it hears directly.
type Location struct {
	// Source is "gps" for the radio's own fix or "estimate" for a position
	// computed from nearby nodes.
	Source           string  `json:"source"`
	LatitudeDegrees  float64 `json:"latitude_degrees"`
	LongitudeDegrees float64 `json:"longitude_degrees"`
	AltitudeMeters   *int    `json:"altitude_meters,omitempty"`
	// Nodes lists the nodes an estimate was computed from, strongest first.
	Nodes []LocationNode `json:"nodes,omitempty"`
}

// LocationNode is one node that contributed to an estimate.
type LocationNode struct {
	ID               string  `json:"node_id"`
	LongName         string  `json:"longname,omitempty"`
	ShortName        string  `json:"shortname,omitempty"`
	SNR              float64 `json:"snr_db"`
	Weight           float64 `json:"weight"`
	LatitudeDegrees  float64 `json:"latitude_degrees"`
	LongitudeDegrees float64 `json:"longitude_degrees"`
	LastHeardHuman   string  `json:"last_heard_human,omitempty"`
}

// locateRadio is the action for the locate command.
func locateRadio(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	responses, err := radio.GetRadioInfo()
	if err != nil {
		return cli.Exit(err, 1)
	}

	var myNode uint32
	var nodes []MeshNode
	for _, response := range responses {
		switch payload := response.GetPayloadVariant().(type) {
		case *gomeshproto.FromRadio_MyInfo:
			myNode = payload.MyInfo.GetMyNodeNum()
		case *gomeshproto.FromRadio_NodeInfo:
			nodes = append(nodes, ProtoNodeToMeshNode(payload))
		}
	}

	location, err := estimateLocation(myNode, nodes, locateMaxAge)
	if err != nil {
		return cli.Exit(err, 1)
	}

	if c.Bool("json") {
		out, err := json.Marshal(location)
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}
	printLocation(location)
	return nil
}

// estimateLocation returns the radio's own GPS position when it has one.
// Otherwise it estimates the position from nodes heard directly (zero hops,
// not via MQTT) within maxAge that report a position, as the mean of their
// positions weighted by received power (10^(SNR/10)), so a strong, close
// neighbour counts for more than a faint, distant one.
func estimateLocation(myNode uint32, nodes []MeshNode, maxAge time.Duration) (Location, error) {
	var candidates []MeshNode
	for _, node := range nodes {
		hasPosition := node.LatitudeDegrees != 0 || node.LongitudeDegrees != 0
		if node.NodeNumber == myNode {
			if hasPosition {
				altitude := node.AltitudeMeters
				return Location{
					Source:           "gps",
					LatitudeDegrees:  node.LatitudeDegrees,
					LongitudeDegrees: node.LongitudeDegrees,
					AltitudeMeters:   &altitude,
				}, nil
			}
			continue
		}
		// Direct means the radio said zero hops; a node whose hop count the
		// radio did not report cannot be assumed to be nearby.
		if !hasPosition || node.ViaMQTT || node.HopsAway == nil || *node.HopsAway != 0 || node.LastHeard == "" {
			continue
		}
		if time.Duration(node.LastHeardSecondsAgo)*time.Second > maxAge {
			continue
		}
		candidates = append(candidates, node)
	}

	if len(candidates) == 0 {
		return Location{}, fmt.Errorf("cannot locate: no GPS fix, and no node with a position was heard directly in the last %s", maxAge)
	}

	var total, lat, lon float64
	location := Location{Source: "estimate"}
	for _, node := range candidates {
		weight := math.Pow(10, node.SNR/10)
		total += weight
		lat += weight * node.LatitudeDegrees
		lon += weight * node.LongitudeDegrees
		location.Nodes = append(location.Nodes, LocationNode{
			ID:               node.ID,
			LongName:         node.LongName,
			ShortName:        node.ShortName,
			SNR:              node.SNR,
			Weight:           weight,
			LatitudeDegrees:  node.LatitudeDegrees,
			LongitudeDegrees: node.LongitudeDegrees,
			LastHeardHuman:   node.LastHeardHuman,
		})
	}
	if total == 0 {
		return Location{}, errors.New("cannot locate: nodes have no usable signal strength")
	}

	// 5 decimals is about a metre, far finer than the estimate deserves.
	location.LatitudeDegrees = roundTo(lat/total, 5)
	location.LongitudeDegrees = roundTo(lon/total, 5)
	for i := range location.Nodes {
		location.Nodes[i].Weight = roundTo(location.Nodes[i].Weight/total, 3)
	}
	sort.Slice(location.Nodes, func(i, j int) bool { return location.Nodes[i].Weight > location.Nodes[j].Weight })
	return location, nil
}

func printLocation(location Location) {
	switch location.Source {
	case "gps":
		fmt.Printf("GPS fix: %.7f, %.7f", location.LatitudeDegrees, location.LongitudeDegrees)
		if location.AltitudeMeters != nil {
			fmt.Printf(" at %dm", *location.AltitudeMeters)
		}
		fmt.Println()
	default:
		fmt.Printf("Estimated position: %.5f, %.5f\n", location.LatitudeDegrees, location.LongitudeDegrees)
		fmt.Printf("No GPS fix; SNR-weighted mean of %d directly heard nodes:\n", len(location.Nodes))
		for _, node := range location.Nodes {
			name := node.LongName
			if name == "" {
				name = node.ID
			}
			fmt.Printf("  %5.1f%%  %+6.2f dB  %-30s %.5f, %.5f  heard %s ago\n",
				node.Weight*100, node.SNR, name, node.LatitudeDegrees, node.LongitudeDegrees, node.LastHeardHuman)
		}
	}
}
