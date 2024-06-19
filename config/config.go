package config

import (
	"flag"

	"github.com/itzg/go-flagsfiller"
)

type Config struct {
	Version        bool   `usage:"show version and exit" env:""`
	Port           string `default:"/dev/ttyUSB0" usage:"serial port to read from"`
	LogLevel       string `default:"info" usage:"log level"`
	UpdateInterval int    `default:"5" usage:"updateinterval for sensor data in seconds"`
	Script         struct {
		Path        string `default:"./sensor_script.go" usage:"path of the script to run"`
		RunInterval int    `default:"5" usage:"script run interval in seconds"`
	}
	Storage struct {
		Id string `default:"default_store" usage:"the storageid of the embedded datastore"`
	}
	Mqtt struct {
		TopicPrefix string `default:"tele" usage:"mqtt topic prefix"`
		Endpoint    string `default:"tcp://localhost:1883" usage:"mqtt endpoint to send sensor data to"`
		Username    string `default:"user" usage:"mqtt username"`
		Password    string `default:"" usage:"mqtt password"`
		ClientID    string `default:"sensor" usage:"mqtt client id"`
	}
}

func Parse(from interface{}) error {
	filler := flagsfiller.New(flagsfiller.WithEnv("sensor"))
	err := filler.Fill(flag.CommandLine, from)
	if err != nil {
		return err
	}

	flag.Parse()
	return nil
}
