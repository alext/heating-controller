package controller

import "github.com/alext/heating-controller/sensor"

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
