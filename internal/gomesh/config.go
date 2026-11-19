package gomesh

import (
	"errors"
	"reflect"
	"strconv"

	pb "github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
)

// GetRadioConfig returns a filtered list of raiod and module config settings
func (r *Radio) GetRadioConfig() (configPackets []*pb.FromRadio_Config, modulePackets []*pb.FromRadio_ModuleConfig, err error) {

	configResponses, err := r.GetRadioInfo()
	if err != nil {
		return
	}

	checks := 0
	for len(configResponses) == 0 || len(modulePackets) == 0 && checks < 5 {
		for _, response := range configResponses {
			if config, ok := response.GetPayloadVariant().(*pb.FromRadio_Config); ok {
				configPackets = append(configPackets, config)
			}
			if moduleConfig, ok := response.GetPayloadVariant().(*pb.FromRadio_ModuleConfig); ok {
				modulePackets = append(modulePackets, moduleConfig)
			}
		}

		checks++
	}

	return
}

// SetRadioConfig allows an freeform setting of values in the RadioConfig_UserPreferences struct
func (r *Radio) SetRadioConfig(key string, value string) error {

	keyFound := false

	configSettings, moduleSettings, err := r.GetRadioConfig()
	if err != nil {
		return err
	}

	// Loop through each config setting returned from the radio to check if
	// valid key was provided. I wish there was a better way to do this, since
	// new config values need to be added manually, but for now this works
	for _, config := range configSettings {

		if deviceConfig := config.Config.GetDevice(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Device{
								Device: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		} else if deviceConfig := config.Config.GetPosition(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Position{
								Position: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		} else if deviceConfig := config.Config.GetPower(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Power{
								Power: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		} else if deviceConfig := config.Config.GetNetwork(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Network{
								Network: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		} else if deviceConfig := config.Config.GetDisplay(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Display{
								Display: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		} else if deviceConfig := config.Config.GetLora(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Lora{
								Lora: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		} else if deviceConfig := config.Config.GetBluetooth(); deviceConfig != nil {
			if keyFound = reflectForKey(deviceConfig, key); keyFound {
				err := setProtoValue(deviceConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetConfig{
						SetConfig: &pb.Config{
							PayloadVariant: &pb.Config_Bluetooth{
								Bluetooth: deviceConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}

	}

	for _, module := range moduleSettings {

		if keyFound {
			break
		}
		if moduleConfig := module.ModuleConfig.GetMqtt(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_Mqtt{
								Mqtt: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetSerial(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_Serial{
								Serial: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetExternalNotification(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_ExternalNotification{
								ExternalNotification: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetStoreForward(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_StoreForward{
								StoreForward: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetRangeTest(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_RangeTest{
								RangeTest: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetTelemetry(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_Telemetry{
								Telemetry: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetCannedMessage(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_CannedMessage{
								CannedMessage: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetAudio(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_Audio{
								Audio: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetRemoteHardware(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_RemoteHardware{
								RemoteHardware: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetNeighborInfo(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_NeighborInfo{
								NeighborInfo: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetAmbientLighting(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_AmbientLighting{
								AmbientLighting: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetDetectionSensor(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_DetectionSensor{
								DetectionSensor: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
		if moduleConfig := module.ModuleConfig.GetPaxcounter(); moduleConfig != nil {
			if keyFound = reflectForKey(moduleConfig, key); keyFound {
				err := setProtoValue(moduleConfig, key, value)
				if err != nil {
					return err
				}
				adminMessage := pb.AdminMessage{
					PayloadVariant: &pb.AdminMessage_SetModuleConfig{
						SetModuleConfig: &pb.ModuleConfig{
							PayloadVariant: &pb.ModuleConfig_Paxcounter{
								Paxcounter: moduleConfig,
							},
						},
					},
				}
				err = sendAdminMessage(adminMessage, r)
				if err != nil {
					return err
				}
				break
			}
		}
	}

	if !keyFound {
		return errors.New("key not found")
	}

	return nil
}

func reflectForKey(t interface{}, key string) (keyFound bool) {

	keyFound = false
	crValues := reflect.ValueOf(t)

	crElements := crValues.Elem()

	fv := crElements.FieldByName(key)

	if fv.IsValid() && fv.CanSet() {
		keyFound = true
	}

	return
}

func setProtoValue(t interface{}, key string, value string) error {
	rPref := reflect.ValueOf(t)

	rPref = rPref.Elem()

	fv := rPref.FieldByName(key)
	vType := fv.Type()

	// The acceptable values that can be set from the command line are uint32 and bool, so only check for those
	switch vType.Kind() {
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fv.SetBool(boolValue)
	case reflect.Uint32:
		uintValue, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			return err
		}
		fv.SetUint(uintValue)
	case reflect.Int32:
		intValue, err := strconv.ParseInt(value, 10, 32)
		if err != nil {
			return err
		}
		fv.SetInt(intValue)
	case reflect.String:
		fv.SetString(value)

	}

	return nil
}

func sendAdminMessage(adminPacket pb.AdminMessage, r *Radio) error {
	out, err := proto.Marshal(&adminPacket)
	if err != nil {
		return err
	}

	nodeNum := r.nodeNum

	packet, err := r.createAdminPacket(nodeNum, out)
	if err != nil {
		return err
	}

	if err := r.SendPacket(packet); err != nil {
		return err
	}

	return nil
}
