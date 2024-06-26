package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/denkhaus/sensor/config"
	"github.com/denkhaus/sensor/logging"
	"github.com/denkhaus/sensor/script"
	"github.com/denkhaus/sensor/store"
	"github.com/pkg/errors"

	"go.bug.st/serial"
	"golang.org/x/sync/errgroup"
)

var (
	BuildVersion = "0.0.0"
	BuildDate    = "n/a"
	BuildCommit  = "n/a"
)

var (
	logger = logging.Logger()
	cnf    config.Config
)

// startup initializes and opens a serial port for communication.
//
// Parameters:
// - config: the config structure containing the name of the serial port to open (string).
//
// Returns:
// - serial.Port: the opened serial port (serial.Port).
// - error: an error if the serial port could not be opened (error).
func startup(config *config.Config) (serial.Port, error) {
	// Check if inputPort is empty
	if config.Usb.Port == "" {
		return nil, errors.New("usb inputPort cannot be empty")
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
		if port == config.Usb.Port {
			found = true
			break
		}
	}

	if config.Usb.Port != "auto" && !found {
		logger.Infof("available ports: %v", ports)
		return nil, errors.Errorf("the port %v you defined was not found", config.Usb.Port)
	}

	mode := &serial.Mode{
		BaudRate: 4800,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	usbInputPort := config.Usb.Port
	if usbInputPort == "auto" {
		usbInputPort = ports[0]
		logger.Infof("-usb-port was set to auto -> choose : %v", usbInputPort)
	}

	logger.Infof("open port: %s", usbInputPort)
	port, err := serial.Open(usbInputPort, mode)
	if err != nil {
		return nil, errors.Wrap(err, "open serial port failed")
	}

	if err = port.SetReadTimeout(
		time.Duration(time.Second * time.Duration(config.Usb.ReadTimeout)),
	); err != nil {
		return nil, errors.Wrap(err, "set port read timeout")
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

	if err := config.Parse(&cnf); err != nil {
		logger.Fatalf("can't create input flags: %v", err)
	}

	if cnf.Version {
		fmt.Printf("%s %s (%s)-(%s)\n", os.Args[0], BuildVersion, BuildCommit, BuildDate)
		os.Exit(0)
	}

	logging.SwitchLogLevel(cnf.LogLevel)

	eg, ctx := errgroup.WithContext(context.Background())

	port, err := startup(&cnf)
	if err != nil {
		logger.Fatalf("startup error: %v", err)
	}

	storage, err := store.Initialize(ctx, logger, &cnf, eg)
	if err != nil {
		logger.Fatalf("initialize storage: %v", err)
	}

	defer storage.Close()

	r := NewDataReader(port)
	if err := r.process(ctx, &cnf, eg); err != nil {
		logger.Fatalf("process data: %v", err)
	}

	if err := script.Initialize(ctx, logger, &cnf, eg); err != nil {
		logger.Fatalf("initialize scriptrunner: %v", err)
	}

	if err := eg.Wait(); err != nil {
		logger.Fatalf("main error: %v", err)
	}
}
