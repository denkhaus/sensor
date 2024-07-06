package store

import (
	"context"

	"github.com/denkhaus/sensor/config"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

var (
	sensorStoreInstance   SensorStore
	embeddedStoreInstance EmbeddedStore
)

func init() {
	sensorStoreInstance = NewSensorStore(100)
}

func Sensor() SensorStore {
	return sensorStoreInstance
}

func Embedded() EmbeddedStore {
	return embeddedStoreInstance
}

func Set(id DataID, data float64) {
	sensorStoreInstance.Set(id, data)
}

func Get(id DataID) float64 {
	return sensorStoreInstance.Get(id)
}

func Initialize(
	ctx context.Context,
	logger *logrus.Logger,
	config *config.Config,
	eg *errgroup.Group,
) (EmbeddedStore, error) {

	storage := NewEmbeddedStore(config.Storage.Id)
	if err := storage.Open(); err != nil {
		return nil, errors.Wrap(err, "open storage")
	}

	embeddedStoreInstance = storage
	return storage, nil
}
