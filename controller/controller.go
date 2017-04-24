package controller

type Controller struct {
	Zones map[string]*Zone
}

func New() *Controller {
	return &Controller{
		Zones: make(map[string]*Zone),
	}
}

func (c *Controller) AddZone(z *Zone) {
	c.Zones[z.ID] = z
}
