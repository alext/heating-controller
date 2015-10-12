package main

import (
	"encoding/json"
	"log"
	"os"
)

type config struct {
	Port  int                   `json:"port"`
	Zones map[string]zoneConfig `json:"zones"`
}

type zoneConfig struct {
	Virtual bool `json:"virtual"`
	GPIOPin int  `json:"gpio_pin"`
}

func loadConfig(filename string) (*config, error) {
	c := &config{
		Port: defaultPort,
	}

	file, err := fs.Open(filename)
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
