package config

import (
	"encoding/json"
	"io"

	"github.com/alext/heating-controller/units"
)

const DefaultPort = 8080

type Config struct {
	Port    int                     `json:"port"`
	Sensors map[string]SensorConfig `json:"sensors"`
	Zones   map[string]ZoneConfig   `json:"zones"`
}

type SensorConfig struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type ZoneConfig struct {
	Virtual    bool              `json:"virtual"`
	GPIOPin    int               `json:"gpio_pin"`
	Thermostat *ThermostatConfig `json:"thermostat"`
}

type ThermostatConfig struct {
	Sensor        string            `json:"sensor"`
	DefaultTarget units.Temperature `json:"default_target"`
}

func New() *Config {
	return &Config{
		Port: DefaultPort,
	}
}

func LoadConfig(input io.Reader) (*Config, error) {
	c := New()

	err := json.NewDecoder(input).Decode(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}
