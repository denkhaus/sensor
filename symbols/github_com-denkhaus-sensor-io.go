// Code generated by 'yaegi extract github.com/denkhaus/sensor/io'. DO NOT EDIT.

package symbols

import (
	"github.com/denkhaus/sensor/io"
	"reflect"
)

func init() {
	Symbols["github.com/denkhaus/sensor/io/io"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"NewIOOperator": reflect.ValueOf(io.NewIOOperator),
		"PinStateHigh":  reflect.ValueOf(io.PinStateHigh),
		"PinStateLow":   reflect.ValueOf(io.PinStateLow),

		// type definitions
		"Operator": reflect.ValueOf((*io.Operator)(nil)),
		"Pin":      reflect.ValueOf((*io.Pin)(nil)),
		"PinState": reflect.ValueOf((*io.PinState)(nil)),
	}
}
