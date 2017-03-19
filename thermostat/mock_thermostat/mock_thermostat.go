package mock_thermostat

import (
	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/thermostat"
)

type mockThermostat struct {
	target sensor.Temperature
}

func New(target sensor.Temperature) thermostat.Thermostat {
	return &mockThermostat{
		target: target,
	}
}

func (t *mockThermostat) Current() sensor.Temperature {
	return 18000
}
func (t *mockThermostat) Target() sensor.Temperature {
	return t.target
}
func (t *mockThermostat) Set(value sensor.Temperature) {
	t.target = value
}
func (t *mockThermostat) Close() {}
