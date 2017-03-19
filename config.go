package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/alext/heating-controller/sensor"
)

type config struct {
	Port  int                   `json:"port"`
	Zones map[string]zoneConfig `json:"zones"`
}

type zoneConfig struct {
	Virtual    bool              `json:"virtual"`
	GPIOPin    int               `json:"gpio_pin"`
	Thermostat *thermostatConfig `json:"thermostat"`
}

type thermostatConfig struct {
	SensorURL     string             `json:"sensor_url"`
	DefaultTarget sensor.Temperature `json:"default_target"`
}

func loadConfig(filename string) (*config, error) {
	c := &config{
		Port: defaultPort,
	}

	file, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("[main] Config file '%s' not found, ignoring", filename)
			return c, nil
		}
		return nil, err
	}

	err = json.NewDecoder(file).Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil

}
