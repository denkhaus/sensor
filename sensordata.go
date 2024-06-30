package main

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/denkhaus/sensor/store"
)

type SensorData struct {
	id   store.DataID
	data []byte
}

func NewSensorData(id store.DataID, data []byte) *SensorData {
	return &SensorData{
		id:   id,
		data: data,
	}
}

func (s *SensorData) Decode() string {
	var decodedValue float64

	switch s.id {
	case store.Humidity:
		decodedValue = float64(binary.BigEndian.Uint16(s.data[3:5])) / 10.0
		store.Set(store.Humidity, decodedValue)
	case store.Temperature:
		curTemp := float64(binary.BigEndian.Uint16(s.data[3:5])) / 10.0
		store.Set(store.Temperature, curTemp)
		decodedValue = curTemp
	case store.Conductivity:
		cond := (float64(binary.BigEndian.Uint16(s.data[3:5])) / 1000.0) * 2.0
		store.Set(store.Conductivity, cond)
		decodedValue = cond
	case store.Salinity:
		decodedValue = float64(binary.BigEndian.Uint16(s.data[3:5]))
		store.Set(store.Salinity, decodedValue)
	case store.TDS:
		decodedValue = float64(binary.BigEndian.Uint16(s.data[3:5]))
		store.Set(store.TDS, decodedValue)
	}

	return strconv.FormatFloat(decodedValue, 'f', 2, 64)
}

func (s *SensorData) Payload() ([]byte, error) {
	data := map[string]interface{}{
		"data": map[string]float64{
			"humidity":     store.Get(store.Humidity),
			"temperature":  store.Get(store.Temperature),
			"conductivity": store.Get(store.Conductivity),
			"salinity":     store.Get(store.Salinity),
			"tds":          store.Get(store.TDS),
		},
	}

	return json.Marshal(data)
}
