package store

//go:generate stringer -type=DataID
var (
	dataStore = NewDataStore()
)

type DataID int

const (
	Humidity DataID = iota
	Temperature
	Conductivity
	Salinity
	TDS
)

func Store() *DataStore {
	return dataStore
}

type DataStore struct {
	data map[DataID]float64
}

func (p *DataStore) Set(id DataID, data float64) {
	p.data[id] = data
}

func (p *DataStore) Get(id DataID) float64 {
	return p.data[id]
}

func NewDataStore() *DataStore {
	return &DataStore{data: make(map[DataID]float64)}
}

func Set(id DataID, data float64) {
	dataStore.Set(id, data)
}

func Get(id DataID) float64 {
	return dataStore.Get(id)
}
