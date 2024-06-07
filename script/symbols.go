package script

import (
	"go/constant"
	"go/token"
	"path/filepath"
	"reflect"

	"github.com/denkhaus/sensor/io"
	"github.com/denkhaus/sensor/store"
)

var Symbols = map[string]map[string]reflect.Value{}

func init() {
	Symbols["github.com/denkhaus/sensor"] = map[string]reflect.Value{
		// function, constant and variable definitions
		"Abs":  reflect.ValueOf(filepath.Abs),
		"Base": reflect.ValueOf(filepath.Base),

		"ListSeparator": reflect.ValueOf(constant.MakeFromLiteral("58", token.INT, 0)),

		// type definitions
		"DataStore":  reflect.ValueOf((*store.DataStore)(nil)),
		"IOOperator": reflect.ValueOf((*io.Operator)(nil)),
		"IOPin":      reflect.ValueOf((*io.Pin)(nil)),
	}
}
