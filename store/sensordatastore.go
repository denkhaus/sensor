package store

import (
	"sync"
)

//go:generate stringer -type=DataID

type ValueStore struct {
	data     []float64
	capacity int
}

// GetAverage calculates the average value stored in the ValueStore.
//
// It checks if there are any values stored in the ValueStore and returns 0 if there are none.
// Otherwise, it iterates over the data stored in the ValueStore and calculates the sum of all values.
// The average is then calculated by dividing the sum by the length of the data slice.
//
// Returns:
// - float64: the average value stored in the ValueStore, or 0 if there are no values.
func (p *ValueStore) GetAverage() float64 {
	if len(p.data) == 0 {
		return 0.0
	}

	var value float64
	for _, v := range p.data {
		value += v
	}

	return value / float64(len(p.data))
}

// Set updates the ValueStore with a new value.
//
// It takes a float64 value as a parameter.
func (p *ValueStore) Set(value float64) {
	if p == nil {
		return
	}

	if len(p.data) > p.capacity {
		p.data = p.data[1:]
	}

	p.data = append(p.data, value)
}

// NewValueStore creates a new instance of ValueStore with the given capacity.
//
// Parameters:
// - size: The capacity of the ValueStore.
//
// Returns:
// - *ValueStore: A pointer to the newly created ValueStore.
func NewValueStore(size int) *ValueStore {
	return &ValueStore{
		data:     []float64{},
		capacity: size,
	}
}

type DataID int

const (
	Humidity DataID = iota
	Temperature
	Conductivity
	Salinity
	TDS
	ConductivityWeighted
	ConductivityRaw
)

type SensorStore interface {
	Set(id DataID, data float64)
	Get(id DataID) float64
}

type sensorStore struct {
	mutex    sync.RWMutex
	data     map[DataID]*ValueStore
	capacity int
}

// Set sets the value of a sensor data in the sensor store.
//
// It takes a DataID and a float64 value as parameters.
// It does not return anything.
func (p *sensorStore) Set(id DataID, data float64) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if store, ok := p.data[id]; ok {
		store.Set(data)
		return
	}

	store := NewValueStore(p.capacity)
	p.data[id] = store
	store.Set(data)
}

// Get retrieves the value of a sensor data from the sensor store.
//
// It takes a DataID as a parameter and returns a float64.
func (p *sensorStore) Get(id DataID) float64 {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if store, ok := p.data[id]; ok {
		return store.GetAverage()
	}

	return 0.0
}

// NewSensorStore creates a new instance of SensorStore with the given size.
//
// Parameters:
// - size: The capacity of the SensorStore.
// Returns:
// - SensorStore: A pointer to the newly created SensorStore.
func NewSensorStore(size int) SensorStore {
	return &sensorStore{
		data:     make(map[DataID]*ValueStore),
		capacity: size,
	}
}
