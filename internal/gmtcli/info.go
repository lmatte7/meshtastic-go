package gmtcli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/golang/protobuf/jsonpb"
	"github.com/lmatte7/gomesh"
	"github.com/lmatte7/gomesh/github.com/meshtastic/gomeshproto"
	"github.com/urfave/cli/v2"
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
	} else {
		return printNodes(nodes)
	}
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
	fmt.Printf(
		"'lf' for %s\n",
		gomeshproto.Config_LoRaConfig_LONG_FAST.String(),
	)
	fmt.Printf(
		"'vls' for %s\n",
		gomeshproto.Config_LoRaConfig_VERY_LONG_SLOW.String(),
	)
	fmt.Printf(
		"'ms' for %s\n",
		gomeshproto.Config_LoRaConfig_MEDIUM_SLOW.String(),
	)
	fmt.Printf(
		"'mf' for %s\n",
		gomeshproto.Config_LoRaConfig_MEDIUM_FAST.String(),
	)
	fmt.Printf(
		"'sl' for %s\n",
		gomeshproto.Config_LoRaConfig_SHORT_SLOW.String(),
	)
	fmt.Printf(
		"'sf' for %s\n",
		gomeshproto.Config_LoRaConfig_SHORT_FAST.String(),
	)
	fmt.Printf(
		"'lm' for %s\n",
		gomeshproto.Config_LoRaConfig_LONG_MODERATE.String(),
	)

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

type MeshNode struct {
	ID                  string  `json:"node_id,omitempty"`
	NodeNumber          int     `json:"node_number,omitempty"`
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
	ChannelUtilization  float64 `json:"channel_utilization_percentage,omitempty"`
	AirUtilTx           float64 `json:"air_utilization_transmit_percentage,omitempty"`
	LastHeard           string  `json:"last_heard,omitempty"`
	LastHeardSecondsAgo int     `json:"last_heard_seconds_ago,omitempty"`
	LastHeardHuman      string  `json:"last_heard_human,omitempty"`
}

func ProtoNodeToMeshNode(node *gomeshproto.FromRadio_NodeInfo) MeshNode {
	meshNode := MeshNode{}

	if node.NodeInfo != nil {
		meshNode.NodeNumber = int(node.NodeInfo.Num)
	}

	if node.NodeInfo.LastHeard != 0 {
		lastSeen := time.Unix(int64(node.NodeInfo.LastHeard), 0).UTC()
		meshNode.LastHeard = lastSeen.Format(time.RFC3339)

		//calculate seconds ago
		secondsAgo := time.Now().UTC().Sub(lastSeen).Seconds()
		meshNode.LastHeardSecondsAgo = int(secondsAgo)
		meshNode.LastHeardHuman = fmt.Sprintf(
			"%s",
			time.Since(lastSeen).Round(time.Second).String(),
		)
	}

	if node.NodeInfo.User == nil {
		return meshNode
	}
	if node.NodeInfo.User.LongName != "" {
		meshNode.LongName = fmt.Sprintf(
			"%s",
			node.NodeInfo.User.LongName,
		)
	}

	if node.NodeInfo.User.Id != "" {
		meshNode.ID = fmt.Sprintf(
			"%s",
			node.NodeInfo.User.Id,
		)
	}

	if node.NodeInfo.User.ShortName != "" {
		meshNode.ShortName = fmt.Sprintf(
			"%s",
			node.NodeInfo.User.ShortName,
		)
	}

	if node.NodeInfo.User.Macaddr != nil {
		meshNode.MACAddress = fmt.Sprintf(
			//format as 48 bit mac address in hex
			"%02x:%02x:%02x:%02x:%02x:%02x",
			node.NodeInfo.User.Macaddr[0],
			node.NodeInfo.User.Macaddr[1],
			node.NodeInfo.User.Macaddr[2],
			node.NodeInfo.User.Macaddr[3],
			node.NodeInfo.User.Macaddr[4],
			node.NodeInfo.User.Macaddr[5],
		)
	}

	if node.NodeInfo.User.IsLicensed {
		meshNode.IsLicensed = true
	}

	//lookup HardwareModel in the protobuf
	if node.NodeInfo.User.HwModel != 0 {
		meshNode.HardwareModel = fmt.Sprintf(
			"%s",
			gomeshproto.HardwareModel_name[int32(node.NodeInfo.User.HwModel)],
		)
	}

	//lookup role in the protobuf
	meshNode.Role = fmt.Sprintf(
		"%s",
		gomeshproto.Config_DeviceConfig_Role_name[int32(node.NodeInfo.User.Role)],
	)

	if node.NodeInfo.DeviceMetrics == nil {
		return meshNode
	}

	if node.NodeInfo.DeviceMetrics.BatteryLevel != 0 {
		meshNode.BatteryLevel = int(node.NodeInfo.DeviceMetrics.BatteryLevel)
	}

	if node.NodeInfo.DeviceMetrics.Voltage != 0 {
		meshNode.Voltage = float64(node.NodeInfo.DeviceMetrics.Voltage)
	}

	if node.NodeInfo.DeviceMetrics.ChannelUtilization != 0 {
		meshNode.ChannelUtilization = float64(
			node.NodeInfo.DeviceMetrics.ChannelUtilization,
		)
	}

	if node.NodeInfo.DeviceMetrics.AirUtilTx != 0 {
		meshNode.AirUtilTx = float64(node.NodeInfo.DeviceMetrics.AirUtilTx)
	}

	if node.NodeInfo.Position == nil {
		return meshNode
	}

	if node.NodeInfo.Position.Altitude != 0 {
		meshNode.AltitudeMeters = int(node.NodeInfo.Position.Altitude)
	}

	if node.NodeInfo.Position.LatitudeI != 0 {
		latds := fmt.Sprintf(
			"%.8f",
			float64(node.NodeInfo.Position.LatitudeI)/1e7,
		)
		//convert string to float
		latd, _ := strconv.ParseFloat(latds, 64)
		meshNode.LatitudeDegrees = latd
	}

	if node.NodeInfo.Position.LongitudeI != 0 {
		lods := fmt.Sprintf(
			"%.7f",
			float64(node.NodeInfo.Position.LongitudeI)/1e7,
		)
		lod, _ := strconv.ParseFloat(lods, 64)
		meshNode.LongitudeDegrees = lod
	}

	return meshNode
}

func printNodesJSON(nodes []*gomeshproto.FromRadio_NodeInfo) error {
	meshNodes := make([]MeshNode, 0)
	for _, node := range nodes {
		mn := ProtoNodeToMeshNode(node)
		meshNodes = append(meshNodes, mn)
	}
	json, err := json.Marshal(meshNodes)
	if err != nil {
		return err
	}
	fmt.Println(string(json))
	return nil
}

func printNodes(nodes []*gomeshproto.FromRadio_NodeInfo) error {
	fmt.Printf("\n")
	fmt.Printf("Nodes in Mesh:\n")

	printDoubleDivider()
	fmt.Printf("| %-15s| ", "Node Number")
	fmt.Printf("%-15s| ", "User")
	fmt.Printf("%-15s| ", "Battery")
	fmt.Printf("%-15s| ", "Altitude")
	fmt.Printf("%-15s| ", "Latitude")
	fmt.Printf("%-15s", "Longitude      |\n")
	printSingleDivider()
	for _, node := range nodes {
		if node.NodeInfo != nil {
			fmt.Printf("| %-15s| ", fmt.Sprint(node.NodeInfo.Num))
			if node.NodeInfo.User != nil {
				fmt.Printf("%-15s| ", node.NodeInfo.User.LongName)
			} else {
				fmt.Printf("%-15s| ", "N/A")
			}
			if node.NodeInfo.DeviceMetrics != nil {
				fmt.Printf(
					"%-15s| ",
					fmt.Sprint(node.NodeInfo.DeviceMetrics.BatteryLevel),
				)
			} else {
				fmt.Printf("%-15s| ", "N/A")
			}
			if node.NodeInfo.Position != nil {
				fmt.Printf(
					"%-15s| ",
					fmt.Sprint(node.NodeInfo.Position.Altitude),
				)
				fmt.Printf(
					"%-15s| ",
					fmt.Sprint(node.NodeInfo.Position.LatitudeI),
				)
				fmt.Printf(
					"%-15s",
					fmt.Sprint(node.NodeInfo.Position.LongitudeI),
				)
			} else {
				fmt.Printf("%-15s| ", "N/A")
				fmt.Printf("%-15s| ", "N/A")
				fmt.Printf("%-15s| ", "N/A")
			}
			fmt.Printf("%s", "|\n")
		}
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
		marshaler := jsonpb.Marshaler{}
		json, err := marshaler.MarshalToString(packet)
		if err != nil {
			fmt.Println("N/A")
		}

		fmt.Println(json)
		if i < (len(info) - 1) {
			fmt.Print(",")
		} else {
			fmt.Print("],")
		}
	}

	fmt.Print("\"channels\":[")
	for i, channel := range channels {
		marshaler := jsonpb.Marshaler{}
		json, _ := marshaler.MarshalToString(channel)
		fmt.Println(json)
		if i < (len(channels) - 1) {
			fmt.Print(",")
		} else {
			fmt.Print("],")
		}
	}

	fmt.Print("\"nodes\":[")
	for i, node := range nodes {
		marshaler := jsonpb.Marshaler{}
		json, _ := marshaler.MarshalToString(node.NodeInfo)
		fmt.Println(json)
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
				fmt.Printf("%-25s", v.Type().Field(i).Name)
				fmt.Printf("%v\n", v.Field(i))
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
