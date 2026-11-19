package gmtcli

import (
	"fmt"
	"reflect"
)

func printDoubleDivider() {
	fmt.Printf("%-100s", "=========================================================================================================================================================================================\n")
}

func printSingleDivider() {
	fmt.Printf("%-100s", "-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------\n")
}

// fieldString renders a reflected struct field for the human-readable views.
// Optional protobuf fields are pointers, so dereference them and show N/A
// when the radio did not report a value instead of printing an address.
// Messages and enums keep their protobuf text form via String().
func fieldString(v reflect.Value) string {
	if v.Kind() == reflect.Ptr && v.IsNil() {
		return "N/A"
	}
	if s, ok := v.Interface().(fmt.Stringer); ok {
		return s.String()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return fmt.Sprintf("%v", v.Interface())
}
