package mock_thermostat

import (
	"github.com/alext/heating-controller/thermostat"
	"github.com/alext/heating-controller/units"
)

type mockThermostat struct {
	target units.Temperature
}

func New(target units.Temperature) thermostat.Thermostat {
	return &mockThermostat{
		target: target,
	}
}

func (t *mockThermostat) Current() units.Temperature {
	return 18000
}
func (t *mockThermostat) Target() units.Temperature {
	return t.target
}
func (t *mockThermostat) Set(value units.Temperature) {
	t.target = value
}
func (t *mockThermostat) Close() {}
