package main

import (
	"github.com/denkhaus/sensor/store"
	"github.com/denkhaus/sensor/types"
)

func Script(ctx *types.ScriptContext) error {

	cond := ctx.SensorStore.Get(store.Conductivity)
	ctx.Logger.Infof("hello denkhaus from script! -> %f", cond)

	return nil
}
