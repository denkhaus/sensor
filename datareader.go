package main

import (
	"context"
	"fmt"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/denkhaus/sensor/config"
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

// process runs the data reading process.
//
// It takes a configuration and error group as parameters.
// Returns an error.
func (p *DataReader) process(
	ctx context.Context,
	config *config.Config,
	eg *errgroup.Group,
) error {

	comChan := make(chan SensorData, ChannelSize)
	durUpdateInterval := time.Second * time.Duration(config.UpdateInterval)

	eg.Go(func() error {
		ticker := time.NewTicker(durUpdateInterval)

		for range ticker.C {
			for dataID := store.DataID(0); dataID < store.DataID(len(sendData)); dataID++ {
				rec, err := p.readSensorData(dataID)
				if err != nil {
					ticker.Stop()
					close(comChan)
					return errors.Wrapf(err, "error reading sensor data for id %d", dataID)
				}

				select {
				case <-ctx.Done():
					ticker.Stop()
					close(comChan)
					logger.Info("data-reader: done received -> closing")
					return nil
				default:
					data := SensorData{id: dataID, data: rec}
					// decode data here to ensure, data is written to the store
					data.Decode()
					if len(comChan) == ChannelSize {
						logger.Warn("sensor data channel is full, dropping data")
					} else {
						comChan <- data
					}
				}
			}
		}
		return nil
	})

	eg.Go(func() error {
		opts := mqtt.NewClientOptions().AddBroker(config.Mqtt.Endpoint)
		opts.SetClientID(config.Mqtt.ClientID).
			SetUsername(config.Mqtt.Username).
			SetPassword(config.Mqtt.Password)

		client := mqtt.NewClient(opts)
		if token := client.Connect(); token.Wait() && token.Error() != nil {
			return errors.Wrap(token.Error(), "mqtt connect error")
		}

		qos := 0
		for sensorData := range comChan {
			topic := fmt.Sprintf("%s/%s/SENSOR", config.Mqtt.TopicPrefix, config.Mqtt.ClientID)
			val := sensorData.Payload()
			token := client.Publish(topic, byte(qos), false, val)

			if token.Wait() && token.Error() != nil {
				return errors.Wrapf(token.Error(), "mqtt publish error for topic %s", topic)
			}

			logger.Debugf("mqtt:sent->%s: %v", topic, val)
		}

		logger.Info("mqtt-writer: channel closed -> closing")
		return nil
	})

	return nil
}
