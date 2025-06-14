package main

import (
	"encoding/binary"
	"encoding/json"
	"strconv"

	"github.com/denkhaus/containers"
	"github.com/denkhaus/sensor/store"
)

const (
	ConductivityDelta = 0.8
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
		cur_hum := float64(binary.BigEndian.Uint16(s.data[3:5])) / 10.0
		cur_hum = containers.Max(0.0, cur_hum)
		cur_hum = containers.Min(100.0, cur_hum)
		store.Set(store.Humidity, cur_hum)
		decodedValue = cur_hum
	case store.Temperature:
		cur_temp := float64(binary.BigEndian.Uint16(s.data[3:5])) / 10.0
		cur_temp = containers.Max(0.0, cur_temp)
		cur_temp = containers.Min(35.0, cur_temp)
		store.Set(store.Temperature, cur_temp)
		decodedValue = cur_temp
	case store.Conductivity:
		cond_raw := float64(binary.BigEndian.Uint16(s.data[3:5]))
		store.Set(store.ConductivityRaw, cond_raw)

		humidityDelta := 1.0
		humidity := store.Get(store.Humidity)
		if humidity != 0.0 {
			humidityDelta = 100.0 / humidity
		}

		cond := (((cond_raw / 1000.0) * humidityDelta) + 1.0) * ConductivityDelta
		cond = containers.Max(0.0, cond)
		cond = containers.Min(5.0, cond)

		store.Set(store.Conductivity, cond)
		decodedValue = cond
	case store.Salinity:
		cur_sal := float64(binary.BigEndian.Uint16(s.data[3:5]))
		store.Set(store.Salinity, cur_sal)
		decodedValue = cur_sal
	case store.TDS:
		cur_tds := float64(binary.BigEndian.Uint16(s.data[3:5]))
		store.Set(store.TDS, cur_tds)
		decodedValue = cur_tds
	}

	cond := store.Get(store.Conductivity)
	temp := store.Get(store.Temperature)

	if cond > 0.0 && temp > 0.0 {
		weightedCond25 := cond * (1 + 0.02*(25.0-temp))
		store.Set(store.ConductivityWeighted, weightedCond25)
	}

	return strconv.FormatFloat(decodedValue, 'f', 2, 64)
}

func (s *SensorData) Payload() ([]byte, error) {
	data := map[string]interface{}{
		"data": map[string]float64{
			"humidity":              store.Get(store.Humidity),
			"temperature":           store.Get(store.Temperature),
			"conductivity":          store.Get(store.Conductivity),
			"conductivity_weighted": store.Get(store.ConductivityWeighted),
			"conductivity_raw":      store.Get(store.ConductivityRaw),
			"salinity":              store.Get(store.Salinity),
			"tds":                   store.Get(store.TDS),
		},
	}

	return json.Marshal(data)
}
