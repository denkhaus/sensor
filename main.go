package main

import (
	"flag"
	"os"

	"github.com/denkhaus/sensor/logging"
	"github.com/denkhaus/sensor/script"
	"github.com/pkg/errors"
	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
)

var (
	logger = logging.Logger()
)

// startup initializes and opens a serial port for communication.
//
// Parameters:
// - inputPort: the name of the serial port to open (string).
//
// Returns:
// - serial.Port: the opened serial port (serial.Port).
// - error: an error if the serial port could not be opened (error).
func startup(inputPort string) (serial.Port, error) {
	// Check if inputPort is empty
	if inputPort == "" {
		return nil, errors.New("inputPort cannot be empty")
	}

	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, errors.Wrap(err, "GetPortsList failed")
	}

	if len(ports) == 0 {
		return nil, errors.New("no serial ports found!")
	}

	found := false
	for _, port := range ports {
		if port == inputPort {
			found = true
			break
		}
	}

	if !found {
		logger.Infof("available ports: %v", ports)
		return nil, errors.Errorf("the port %v you defined was not found", inputPort)
	}

	mode := &serial.Mode{
		BaudRate: 4800,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	logger.Infof("open port: %s", inputPort)
	port, err := serial.Open(inputPort, mode)
	if err != nil {
		return nil, errors.Wrap(err, "open serial port failed")
	}

	return port, nil
}

// main is the entry point of the Go program.
//
// It parses the command line flags for the serial port, MQTT endpoint, MQTT client ID, and update interval.
// It initializes the serial port and starts two goroutines: one to read sensor data from the serial port and update the receiveData slice,
// and another to publish the receiveData to the MQTT broker at the specified interval.
// The main function waits for both goroutines to complete using the errgroup.Group.Wait() method.
//
// No parameters.
// No return values.

func main() {

	flags := flag.NewFlagSet("sensor", flag.ExitOnError)
	inputPort := flags.String("port", "/dev/ttyUSB0", "serial port to read from")
	mqttEndpoint := flags.String("mqttEndpoint", "tcp://localhost:1883", "mqtt endpoint to send sensor data to")
	mqttUsername := flags.String("mqttUsername", "user", "mqtt username")
	mqttPassword := flags.String("mqttPassword", "", "mqtt password")
	mqttClientID := flags.String("mqttClientID", "sensor", "mqtt client id")
	updateInterval := flags.Int("updateInterval", 5, "updateinterval in seconds")
	logLevel := flags.String("logLevel", "info", "log level")
	scriptPath := flags.String("scriptPath", "./script.go", "path of the script to run")
	scriptRunInterval := flags.Int("scriptRunInterval", 5, "script run interval in seconds")

	if err := flags.Parse(os.Args[1:]); err != nil {
		logger.Fatalf("can't parse input flags: %v", err)
	}

	eg := errgroup.Group{}
	if err := script.Initialize(*scriptPath, *scriptRunInterval, &eg); err != nil {
		logger.Warnf("can't initialize scriptrunner: %v", err)
	}

	logging.SwitchLogLevel(*logLevel)
	port, err := startup(*inputPort)
	if err != nil {
		logger.Fatalf("startup error: %v", err)
	}

	r := NewDataReader(port)
	if err := r.process(
		*updateInterval,
		*mqttEndpoint,
		*mqttClientID,
		*mqttUsername,
		*mqttPassword,
		&eg,
	); err != nil {
		logger.Fatalf("can't process data: %v", err)
	}
}
