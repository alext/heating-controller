package config

import (
	"encoding/json"
	"io"

	"github.com/alext/heating-controller/units"
)

const DefaultPort = 8080

type Config struct {
	Port    int                     `json:"port"`
	MQTT    *MQTTConfig             `json:"mqtt"`
	Sensors map[string]SensorConfig `json:"sensors"`
	Zones   map[string]ZoneConfig   `json:"zones"`
}

type MQTTConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type SensorConfig struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Topic string `json:"topic"`
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

func (c *Config) populateDefaults() {
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.MQTT != nil && c.MQTT.Port == 0 {
		c.MQTT.Port = 1883
	}
}

func New() *Config {
	c := &Config{
		Sensors: make(map[string]SensorConfig),
		Zones:   make(map[string]ZoneConfig),
	}
	c.populateDefaults()
	return c
}

func LoadConfig(input io.Reader) (*Config, error) {
	var c Config

	err := json.NewDecoder(input).Decode(&c)
	if err != nil {
		return nil, err
	}
	c.populateDefaults()
	return &c, nil
}
