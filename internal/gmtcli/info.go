package gmtcli

import (
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"time"

	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
	"google.golang.org/protobuf/encoding/protojson"
)

func showRadioInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	responses, err := radio.GetRadioInfo()
	if err != nil {
		return err
	}

	printRadioInfo(responses)

	return nil

}

func showAllRadioInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return getRadioInfo(radio, c.Bool("json"))
}

func showNodeInfo(c *cli.Context) error {

	radio := getRadio(c)
	defer radio.Close()

	return displayNodes(radio, c.Bool("json"))
}

func factoryResetRadio(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	err := radio.FactoryRest()
	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}

func showPositionInfo(c *cli.Context) error {

	positionPacket := &gomeshproto.FromRadio{}

	r := getRadio(c)
	defer r.Close()

	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	for _, packet := range responses {
		if config := packet.GetConfig(); config != nil {
			if gpsConfig := config.GetPosition(); gpsConfig != nil {
				positionPacket = packet
			}
		}
	}

	displayPositionInfo(positionPacket)

	return nil
}

func displayNodes(r gomesh.Radio, jsonFormat bool) error {
	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	for _, response := range responses {
		if nodeInfo, ok := response.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}
	}

	if jsonFormat {
		return printNodesJSON(nodes)
	}
	return printNodes(nodes)
}

func getRadioInfo(r gomesh.Radio, json bool) error {

	responses, err := r.GetRadioInfo()
	if err != nil {
		return err
	}

	if json {
		printJsonRadioInfo(responses)
	} else {
		printRadioInfo(responses)
	}

	return nil
}

func showModemOptions(c *cli.Context) error {
	fmt.Println("Modem Options")
	printDoubleDivider()
	fmt.Printf("'lf' for %s\n", gomeshproto.Config_LoRaConfig_LONG_FAST.String())
	fmt.Printf("'vls' for %s\n", gomeshproto.Config_LoRaConfig_VERY_LONG_SLOW.String())
	fmt.Printf("'ms' for %s\n", gomeshproto.Config_LoRaConfig_MEDIUM_SLOW.String())
	fmt.Printf("'mf' for %s\n", gomeshproto.Config_LoRaConfig_MEDIUM_FAST.String())
	fmt.Printf("'sl' for %s\n", gomeshproto.Config_LoRaConfig_SHORT_SLOW.String())
	fmt.Printf("'sf' for %s\n", gomeshproto.Config_LoRaConfig_SHORT_FAST.String())
	fmt.Printf("'lm' for %s\n", gomeshproto.Config_LoRaConfig_LONG_MODERATE.String())

	return nil
}

func setModemOption(c *cli.Context) error {
	radio := getRadio(c)
	defer radio.Close()

	err := radio.SetModemMode(c.String("option"))
	if err != nil {
		return cli.Exit(err, 0)
	}

	return nil
}

// MeshNode is a flattened, human-friendly view of a node on the mesh. Enums
// are resolved to their names, the MAC address is formatted, coordinates are
// in degrees and the last-heard time is given both as a timestamp and as an
// age. Both the JSON output and the node table are produced from it, so any
// improvement here shows up in both views.
type MeshNode struct {
	ID                  string  `json:"node_id,omitempty"`
	NodeNumber          uint32  `json:"node_number,omitempty"`
	LongName            string  `json:"longname,omitempty"`
	ShortName           string  `json:"shortname,omitempty"`
	AltitudeMeters      int     `json:"altitude_meters,omitempty"`
	LatitudeDegrees     float64 `json:"latitude_degrees,omitempty"`
	LongitudeDegrees    float64 `json:"longitude_degrees,omitempty"`
	MACAddress          string  `json:"mac_address,omitempty"`
	IsLicensed          bool    `json:"is_licensed"`
	Role                string  `json:"role,omitempty"`
	HardwareModel       string  `json:"hardware_model,omitempty"`
	BatteryLevel        int     `json:"battery_level_percent,omitempty"`
	Voltage             float64 `json:"battery_volts,omitempty"`
	ChannelUtilization  float64 `json:"channel_utilization_percent,omitempty"`
	AirUtilTx           float64 `json:"air_utilization_transmit_percent,omitempty"`
	LastHeard           string  `json:"last_heard,omitempty"`
	LastHeardSecondsAgo int     `json:"last_heard_seconds_ago,omitempty"`
	LastHeardHuman      string  `json:"last_heard_human,omitempty"`
	// LastHeardSource is set in stream nodes events to say where the
	// last-heard fields come from: "stream", "radio_corrected" or "radio".
	LastHeardSource string `json:"last_heard_source,omitempty"`
	// SNR is the signal-to-noise ratio of the last packet heard from the node.
	SNR float64 `json:"snr_db,omitempty"`
	// HopsAway is how many hops away the node was last heard; 0 means heard
	// directly.
	HopsAway int `json:"hops_away"`
	// ViaMQTT is true when the node was last heard through an MQTT gateway
	// rather than over the air.
	ViaMQTT bool `json:"via_mqtt"`
}

// macAddress returns the node's MAC address, or "" if the radio did not report
// one. Meshtastic has deprecated the macaddr field and newer firmware omits
// it, while current radios still send it. It is read by name through protobuf
// reflection rather than the deprecated accessor, so the address stays in the
// output while radios provide it and the code keeps compiling if a future
// regeneration of the protobufs drops the field.
func macAddress(user *gomeshproto.User) string {
	message := user.ProtoReflect()
	field := message.Descriptor().Fields().ByName("macaddr")
	if field == nil || !message.Has(field) {
		return ""
	}
	return net.HardwareAddr(message.Get(field).Bytes()).String()
}

// ProtoNodeToMeshNode converts a NodeInfo packet from the radio into a MeshNode.
func ProtoNodeToMeshNode(node *gomeshproto.FromRadio_NodeInfo) MeshNode {
	return protoNodeToMeshNodeAt(node, time.Now())
}

// protoNodeToMeshNodeAt is ProtoNodeToMeshNode with an explicit "now" so the
// last-heard age can be computed deterministically.
func protoNodeToMeshNodeAt(node *gomeshproto.FromRadio_NodeInfo, now time.Time) MeshNode {
	meshNode := MeshNode{}

	if node == nil || node.NodeInfo == nil {
		return meshNode
	}
	info := node.NodeInfo

	meshNode.NodeNumber = info.Num
	// Meshtastic derives a node's id from its number, so one can always be
	// given even when the radio has not received the node's user info.
	meshNode.ID = nodeIDFromNumber(info.Num)
	meshNode.SNR = float64(info.Snr)
	meshNode.HopsAway = int(info.HopsAway)
	meshNode.ViaMQTT = info.ViaMqtt

	if info.LastHeard != 0 {
		lastSeen := time.Unix(int64(info.LastHeard), 0).UTC()
		sinceLastSeen := now.Sub(lastSeen)
		meshNode.LastHeard = lastSeen.Format(time.RFC3339)
		meshNode.LastHeardSecondsAgo = int(sinceLastSeen.Seconds())
		meshNode.LastHeardHuman = sinceLastSeen.Round(time.Second).String()
	}

	if user := info.GetUser(); user != nil {
		if user.Id != "" {
			meshNode.ID = user.Id
		}
		meshNode.LongName = user.LongName
		meshNode.ShortName = user.ShortName
		meshNode.IsLicensed = user.IsLicensed
		meshNode.Role = user.Role.String()
		if user.HwModel != gomeshproto.HardwareModel_UNSET {
			meshNode.HardwareModel = user.HwModel.String()
		}
		meshNode.MACAddress = macAddress(user)
	}

	if metrics := info.GetDeviceMetrics(); metrics != nil {
		meshNode.BatteryLevel = int(metrics.BatteryLevel)
		meshNode.Voltage = float64(metrics.Voltage)
		meshNode.ChannelUtilization = float64(metrics.ChannelUtilization)
		meshNode.AirUtilTx = float64(metrics.AirUtilTx)
	}

	if position := info.GetPosition(); position != nil {
		meshNode.AltitudeMeters = int(position.Altitude)
		// The radio reports coordinates as integers in units of 1e-7 degrees.
		meshNode.LatitudeDegrees = float64(position.LatitudeI) / 1e7
		meshNode.LongitudeDegrees = float64(position.LongitudeI) / 1e7
	}

	return meshNode
}

func printNodesJSON(nodes []*gomeshproto.FromRadio_NodeInfo) error {
	meshNodes := make([]MeshNode, 0, len(nodes))
	for _, node := range nodes {
		meshNodes = append(meshNodes, ProtoNodeToMeshNode(node))
	}
	out, err := json.Marshal(meshNodes)
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// orNA returns "N/A" for empty values so the table never shows blank cells.
func orNA(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}

func printNodes(nodes []*gomeshproto.FromRadio_NodeInfo) error {
	fmt.Printf("\n")
	fmt.Printf("Nodes in Mesh:\n")

	printDoubleDivider()
	fmt.Printf("| %-10s| ", "Node ID")
	fmt.Printf("%-11s| ", "Node Number")
	fmt.Printf("%-40s| ", "Long Name")
	fmt.Printf("%-6s| ", "Short")
	fmt.Printf("%-14s| ", "Role")
	fmt.Printf("%-24s| ", "Hardware")
	fmt.Printf("%-8s| ", "Battery")
	fmt.Printf("%-9s| ", "Altitude")
	fmt.Printf("%-13s| ", "Latitude")
	fmt.Printf("%-13s| ", "Longitude")
	fmt.Printf("%-14s|\n", "Last Heard")
	printSingleDivider()
	for _, node := range nodes {
		if node.NodeInfo == nil {
			continue
		}
		mn := ProtoNodeToMeshNode(node)

		// Zero is a valid reading (sea level, 0% battery), so decide what to
		// show from whether the radio reported the section at all.
		battery, altitude, latitude, longitude, lastHeard := "N/A", "N/A", "N/A", "N/A", "N/A"
		if node.NodeInfo.DeviceMetrics != nil {
			battery = fmt.Sprintf("%d%%", mn.BatteryLevel)
		}
		if position := node.NodeInfo.Position; position != nil {
			altitude = fmt.Sprintf("%dm", mn.AltitudeMeters)
			// A position with both coordinates exactly zero means no fix.
			if position.LatitudeI != 0 || position.LongitudeI != 0 {
				latitude = fmt.Sprintf("%.7f", mn.LatitudeDegrees)
				longitude = fmt.Sprintf("%.7f", mn.LongitudeDegrees)
			}
		}
		if mn.LastHeardHuman != "" {
			lastHeard = mn.LastHeardHuman + " ago"
		}

		fmt.Printf("| %-10s| ", orNA(mn.ID))
		fmt.Printf("%-11d| ", mn.NodeNumber)
		fmt.Printf("%-40s| ", orNA(mn.LongName))
		fmt.Printf("%-6s| ", orNA(mn.ShortName))
		fmt.Printf("%-14s| ", orNA(mn.Role))
		fmt.Printf("%-24s| ", orNA(mn.HardwareModel))
		fmt.Printf("%-8s| ", battery)
		fmt.Printf("%-9s| ", altitude)
		fmt.Printf("%-13s| ", latitude)
		fmt.Printf("%-13s| ", longitude)
		fmt.Printf("%-14s|\n", lastHeard)
	}
	printDoubleDivider()
	return nil
}

func printJsonRadioInfo(info []*gomeshproto.FromRadio) {
	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	channels := make([]*gomeshproto.Channel, 0)

	fmt.Print("{ \"packets\": [")
	for i, packet := range info {
		if nodeInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
			continue
		}
		if channelInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_Channel); ok {
			channels = append(channels, channelInfo.Channel)
			continue
		}
		out, err := protojson.Marshal(packet)
		if err != nil {
			fmt.Println("N/A")
		}

		fmt.Println(string(out))
		if i < (len(info) - 1) {
			fmt.Print(",")
		} else {
			fmt.Print("],")
		}
	}

	fmt.Print("\"channels\":[")
	for i, channel := range channels {
		out, _ := protojson.Marshal(channel)
		fmt.Println(string(out))
		if i < (len(channels) - 1) {
			fmt.Print(",")
		} else {
			fmt.Print("],")
		}
	}

	fmt.Print("\"nodes\":[")
	for i, node := range nodes {
		out, _ := protojson.Marshal(node.NodeInfo)
		fmt.Println(string(out))
		if i < (len(nodes) - 1) {
			fmt.Print(",")
		} else {
			fmt.Print("]}\n")
		}
	}
}

func printRadioInfo(info []*gomeshproto.FromRadio) {
	fmt.Printf("%s", "\nRadio Settings: \n")
	nodes := make([]*gomeshproto.FromRadio_NodeInfo, 0)
	channels := make([]*gomeshproto.Channel, 0)
	positionPacket := &gomeshproto.FromRadio{}

	for _, packet := range info {
		if nodeInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_NodeInfo); ok {
			nodes = append(nodes, nodeInfo)
		}
		if channelInfo, ok := packet.GetPayloadVariant().(*gomeshproto.FromRadio_Channel); ok {
			channels = append(channels, channelInfo.Channel)
		}
		if config := packet.GetConfig(); config != nil {
			if gpsConfig := config.GetPosition(); gpsConfig != nil {
				positionPacket = packet
			}
			if deviceInfo := config.GetDevice(); deviceInfo != nil {
				fmt.Printf("%s", "\nDevice Settings\n")
				v := reflect.ValueOf(*deviceInfo)
				for i := 0; i < v.NumField(); i++ {
					if v.Field(i).CanInterface() {
						fmt.Printf("%-25s", v.Type().Field(i).Name)
						fmt.Printf("%v\n", v.Field(i))
					}
				}
			}
		}

		if metaInfo := packet.GetMetadata(); metaInfo != nil {
			fmt.Printf("%s", "Radio Metadata\n")
			v := reflect.ValueOf(*metaInfo)
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanInterface() {
					fmt.Printf("%-25s", v.Type().Field(i).Name)
					fmt.Printf("%v\n", v.Field(i))
				}
			}
		}
		if nodeInfo := packet.GetNodeInfo(); nodeInfo != nil {
			fmt.Printf("%s", "\n\nNode Info\n")
			v := reflect.ValueOf(*nodeInfo)
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanInterface() {
					fmt.Printf("%-25s", v.Type().Field(i).Name)
					fmt.Printf("%v\n", v.Field(i))
				}
			}
		}
	}

	displayPositionInfo(positionPacket)
	printNodes(nodes)
	printChannels(channels)

}

func displayPositionInfo(packet *gomeshproto.FromRadio) {
	if config := packet.GetConfig(); config != nil {
		if gpsConfig := config.GetPosition(); gpsConfig != nil {
			fmt.Printf("%s", "\n\nPosition Settings\n")
			v := reflect.ValueOf(*gpsConfig)
			for i := 0; i < v.NumField(); i++ {
				if v.Field(i).CanInterface() {
					fmt.Printf("%-35s", v.Type().Field(i).Name)
					fmt.Printf("%v\n", v.Field(i))
				}
			}
		}
	}
}
