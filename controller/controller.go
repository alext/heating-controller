package controller

import (
	"fmt"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/sensor"
)

type Controller struct {
	Sensors map[string]sensor.Sensor
	Zones   map[string]*Zone
}

func New() *Controller {
	return &Controller{
		Sensors: make(map[string]sensor.Sensor),
		Zones:   make(map[string]*Zone),
	}
}

func (c *Controller) AddSensor(name string, s sensor.Sensor) {
	c.Sensors[name] = s
}

func (c *Controller) AddZone(z *Zone) {
	c.Zones[z.ID] = z
}

func (c *Controller) SetupSensors(sensors map[string]config.SensorConfig) error {
	for name, sensorConfig := range sensors {
		s, err := sensor.New(sensorConfig)
		if err != nil {
			return err
		}
		c.AddSensor(name, s)
	}
	return nil
}

var outputNew = output.New // variable indirection to facilitate testing

func (c *Controller) SetupZones(zones map[string]config.ZoneConfig) error {
	for id, config := range zones {
		var out output.Output
		if config.Virtual {
			out = output.Virtual(id)
		} else {
			var err error
			out, err = outputNew(id, config.GPIOPin)
			if err != nil {
				return err
			}
		}
		z := NewZone(id, out)
		if config.Thermostat != nil {
			s, ok := c.Sensors[config.Thermostat.Sensor]
			if !ok {
				return fmt.Errorf("Non-existent sensor: '%s'", config.Thermostat.Sensor)
			}
			z.SetupThermostat(s, config.Thermostat.DefaultTarget)
		}
		z.Restore()
		z.Scheduler.Start()
		c.AddZone(z)
	}
	return nil
}
