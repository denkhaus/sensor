package types

import (
	"github.com/sirupsen/logrus"

	"github.com/denkhaus/sensor/store"
)

type ScriptContext struct {
	Logger        *logrus.Logger
	SensorStore   store.SensorStore
	EmbeddedStore store.EmbeddedStore
}
