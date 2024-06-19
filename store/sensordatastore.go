package store

//go:generate stringer -type=DataID

type DataID int

const (
	Humidity DataID = iota
	Temperature
	Conductivity
	Salinity
	TDS
)

type SensorStore interface {
	Set(id DataID, data float64)
	Get(id DataID) float64
}

type sensorStore struct {
	data map[DataID]float64
}

func (p *sensorStore) Set(id DataID, data float64) {
	p.data[id] = data
}

func (p *sensorStore) Get(id DataID) float64 {
	return p.data[id]
}

func NewSensorStore() SensorStore {
	return &sensorStore{data: make(map[DataID]float64)}
}
