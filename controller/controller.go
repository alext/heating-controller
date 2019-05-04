package controller

import (
	"fmt"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/sensor"
)

type Controller struct {
	SensorsByName map[string]sensor.Sensor
	SensorsByID   map[string]sensor.Sensor
	Zones         map[string]*Zone
}

func New() *Controller {
	return &Controller{
		SensorsByName: make(map[string]sensor.Sensor),
		SensorsByID:   make(map[string]sensor.Sensor),
		Zones:         make(map[string]*Zone),
	}
}

func (c *Controller) AddSensor(name string, s sensor.Sensor) {
	c.SensorsByName[name] = s
	c.SensorsByID[s.ID()] = s
}

func (c *Controller) AddZone(z *Zone) {
	c.Zones[z.ID] = z
}

var outputNew = output.New // variable indirection to facilitate testing

func (c *Controller) Setup(cfg *config.Config) error {
	for name, sensorConfig := range cfg.Sensors {
		s, err := sensor.New(name, sensorConfig)
		if err != nil {
			return err
		}
		c.AddSensor(name, s)
	}

	for name, zoneConfig := range cfg.Zones {
		var out output.Output
		if zoneConfig.Virtual {
			out = output.Virtual(name)
		} else {
			var err error
			out, err = outputNew(name, zoneConfig.GPIOPin)
			if err != nil {
				return err
			}
		}
		z := NewZone(name, out)
		if zoneConfig.Thermostat != nil {
			s, ok := c.SensorsByName[zoneConfig.Thermostat.Sensor]
			if !ok {
				return fmt.Errorf("Non-existent sensor '%s' for zone '%s'", zoneConfig.Thermostat.Sensor, name)
			}
			z.SetupThermostat(s, zoneConfig.Thermostat.DefaultTarget)
		}
		z.Restore()
		z.Scheduler.Start()
		c.AddZone(z)
	}

	return nil
}
