package controller

import (
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/timer"
)

type Controller interface {
	Outputs() map[string]output.Output
	Output(string) output.Output
	AddOutput(output.Output)
	Timers() map[string]timer.Timer
	Timer(string) timer.Timer
	AddTimer(timer.Timer)
}

type controller struct {
	outputs map[string]output.Output
	timers  map[string]timer.Timer
}

func New() *controller {
	return &controller{
		outputs: make(map[string]output.Output),
		timers:  make(map[string]timer.Timer),
	}
}

func (c *controller) Outputs() map[string]output.Output {
	return c.outputs
}

func (c *controller) Output(id string) output.Output {
	return c.outputs[id]
}

func (c *controller) AddOutput(out output.Output) {
	c.outputs[out.Id()] = out
}

func (c *controller) Timers() map[string]timer.Timer {
	return c.timers
}

func (c *controller) Timer(id string) timer.Timer {
	return c.timers[id]
}

func (c *controller) AddTimer(t timer.Timer) {
	c.timers[t.Id()] = t
}
