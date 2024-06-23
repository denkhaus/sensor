// Code generated by 'yaegi extract github.com/denkhaus/sensor/types'. DO NOT EDIT.

package symbols

import (
	"github.com/denkhaus/sensor/types"
	"reflect"
)

func init() {
	Symbols["github.com/denkhaus/sensor/types/types"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"NewTimespan":                 reflect.ValueOf(types.NewTimespan),
		"SwitchTimerStateInitialized": reflect.ValueOf(types.SwitchTimerStateInitialized),
		"SwitchTimerStateOff":         reflect.ValueOf(types.SwitchTimerStateOff),
		"SwitchTimerStateOn":          reflect.ValueOf(types.SwitchTimerStateOn),

		// type definitions
		"PulseTimer":       reflect.ValueOf((*types.PulseTimer)(nil)),
		"ScriptContext":    reflect.ValueOf((*types.ScriptContext)(nil)),
		"Span":             reflect.ValueOf((*types.Span)(nil)),
		"SwitchTimer":      reflect.ValueOf((*types.SwitchTimer)(nil)),
		"SwitchTimerState": reflect.ValueOf((*types.SwitchTimerState)(nil)),
	}
}
