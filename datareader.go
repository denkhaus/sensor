package main

import (
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/denkhaus/sensor/store"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/pkg/errors"
	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
)

const (
	ChannelSize = 100
)

var (
	// Humidity:     01 03 00 00 00 01 84 0a
	// Temperatur:   01 03 00 01 00 01 d5 ca
	// Conductivity: 01 03 00 02 00 01 25 ca
	// Salinity:     01 03 00 03 00 01 74 0a
	// TDS:          01 03 00 04 00 01 c5 cb

	sendData = [][]byte{
		{0x01, 0x03, 0x00, 0x00, 0x00, 0x01, 0x84, 0x0a},
		{0x01, 0x03, 0x00, 0x01, 0x00, 0x01, 0xd5, 0xca},
		{0x01, 0x03, 0x00, 0x02, 0x00, 0x01, 0x25, 0xca},
		{0x01, 0x03, 0x00, 0x03, 0x00, 0x01, 0x74, 0x0a},
		{0x01, 0x03, 0x00, 0x04, 0x00, 0x01, 0xc5, 0xcb},
	}
)

type DataReader struct {
	port serial.Port
}

func NewDataReader(port serial.Port) *DataReader {
	reader := DataReader{port: port}
	return &reader
}

// readSensorData reads sensor data from the specified dataID and returns the received data as a byte slice.
//
// Parameters:
// - dataID: the ID of the data to be read from the sensor.
//
// Returns:
// - []byte: the received data as a byte slice.
// - error: an error if the dataReader is nil, the dataID is invalid, or there is an error writing or reading data from the sensor.

func (p *DataReader) readSensorData(dataID store.DataID) ([]byte, error) {
	if p == nil {
		return nil, errors.New("dataReader is nil")
	}

	if dataID < 0 || dataID >= store.DataID(len(sendData)) {
		return nil, errors.Errorf("invalid dataID: %d", dataID)
	}

	dataToSend := sendData[dataID]
	_, err := p.port.Write(dataToSend)
	if err != nil {
		return nil, errors.Errorf("can't write data to sensor: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	buff := make([]byte, 8)
	n, err := p.port.Read(buff)
	if err != nil {
		return nil, errors.Errorf("can't read data from sensor: %v", err)
	}

	if n == 0 {
		return nil, errors.New("EOF received")
	}

	result := buff[:n]
	logger.Debugf("tx: %s", spew.Sprint(dataToSend))
	logger.Debugf("rx: %s", spew.Sprint(result))

	return result, nil
}

// process reads sensor data at a specified interval and publishes it to an MQTT broker.
//
// Parameters:
// - updateInterval: the interval in seconds at which to read sensor data.
// - mqttEndpoint: the MQTT endpoint to publish sensor data to.
// - mqttClientID: the MQTT client ID.
//
// Returns:
// - error: an error if there was a problem reading sensor data or publishing it to the MQTT broker.
func (p *DataReader) process(
	updateInterval int,
	mqttEndpoint string,
	mqttClientID string,
	mqttUsername string,
	mqttPassword string,
	eg *errgroup.Group,
) error {

	comChan := make(chan SensorData, ChannelSize)
	durUpdateInterval := time.Second * time.Duration(updateInterval)

	eg.Go(func() error {
		ticker := time.NewTicker(durUpdateInterval)

		for range ticker.C {
			for dataID := store.DataID(0); dataID < store.DataID(len(sendData)); dataID++ {
				rec, err := p.readSensorData(dataID)
				if err != nil {
					close(comChan)
					return errors.Wrapf(err, "error reading sensor data for id %d", dataID)
				}

				comChan <- SensorData{id: dataID, data: rec}
			}
		}
		return nil
	})

	eg.Go(func() error {
		opts := mqtt.NewClientOptions().AddBroker(mqttEndpoint)
		opts.SetClientID(mqttClientID).
			SetUsername(mqttUsername).
			SetPassword(mqttPassword)

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			close(comChan)
			return errors.Wrap(token.Error(), "mqtt connect error")
		}

		qos := 0
		for sensorData := range comChan {
			topic := fmt.Sprintf("sensor/data/%s", sensorData.id)
			val := sensorData.Decode()
			token := client.Publish(topic, byte(qos), false, val)

			if token.Wait() && token.Error() != nil {
				close(comChan)
				return errors.Wrapf(token.Error(), "mqtt publish error for topic %s", topic)
			}

			logger.Infof("mqtt:sent->%s: %v", topic, val)
		}

		return nil
	})

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("main error: %v", err)
	}

	return nil
}
