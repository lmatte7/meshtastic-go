package gomesh

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	pb "github.com/lmatte7/meshtastic-go/internal/gomesh/gomeshproto"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
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

func reflectForKey(t any, key string) (keyFound bool) {

	keyFound = false
	crValues := reflect.ValueOf(t)

	crElements := crValues.Elem()

	fv := crElements.FieldByName(key)

	if fv.IsValid() && fv.CanSet() {
		keyFound = true
	}

	return
}

func setProtoValue(t any, key string, value string) error {
	fv := reflect.ValueOf(t).Elem().FieldByName(key)
	return setFieldValue(fv, value)
}

// enumType is implemented by every generated protobuf enum.
var enumType = reflect.TypeOf((*protoreflect.Enum)(nil)).Elem()

// setFieldValue parses value as text and stores it in a protobuf struct
// field. Enums accept their value names (case-insensitive, for example "US"
// or "LONG_FAST") as well as numbers; optional scalar fields, which the
// generated code represents as pointers, are set through a new pointer.
func setFieldValue(fv reflect.Value, value string) error {
	if !fv.IsValid() || !fv.CanSet() {
		return errors.New("unknown Field")
	}

	if fv.Kind() == reflect.Ptr && fv.Type().Elem().Kind() != reflect.Struct {
		target := reflect.New(fv.Type().Elem())
		if err := setFieldValue(target.Elem(), value); err != nil {
			return err
		}
		fv.Set(target)
		return nil
	}

	if fv.Type().Implements(enumType) {
		return setEnumValue(fv, value)
	}

	switch fv.Kind() {
	case reflect.Bool:
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fv.SetBool(boolValue)
	case reflect.Uint32, reflect.Uint64:
		uintValue, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		fv.SetUint(uintValue)
	case reflect.Int32, reflect.Int64:
		intValue, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		fv.SetInt(intValue)
	case reflect.Float32, reflect.Float64:
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		fv.SetFloat(floatValue)
	case reflect.String:
		fv.SetString(value)
	case reflect.Slice, reflect.Array:
		fv.SetBytes([]byte(value))
	default:
		return fmt.Errorf("cannot set a field of type %s from the command line", fv.Type())
	}
	return nil
}

// setEnumValue stores an enum given by name or number.
func setEnumValue(fv reflect.Value, value string) error {
	if number, err := strconv.ParseInt(value, 10, 32); err == nil {
		fv.SetInt(number)
		return nil
	}
	descriptor := fv.Interface().(protoreflect.Enum).Descriptor()
	values := descriptor.Values()
	for i := range values.Len() {
		if strings.EqualFold(string(values.Get(i).Name()), value) {
			fv.SetInt(int64(values.Get(i).Number()))
			return nil
		}
	}
	names := make([]string, 0, values.Len())
	for i := range values.Len() {
		names = append(names, string(values.Get(i).Name()))
	}
	return fmt.Errorf("unknown value %q; choose one of %s", value, strings.Join(names, ", "))
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
