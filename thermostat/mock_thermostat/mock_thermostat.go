package mock_thermostat

import "github.com/alext/heating-controller/thermostat"

type mockThermostat struct {
	target thermostat.Temperature
}

func New(target thermostat.Temperature) thermostat.Thermostat {
	return &mockThermostat{
		target: target,
	}
}

func (t *mockThermostat) Current() thermostat.Temperature {
	return 18000
}
func (t *mockThermostat) Target() thermostat.Temperature {
	return t.target
}
func (t *mockThermostat) Set(value thermostat.Temperature) {
	t.target = value
}
func (t *mockThermostat) Close() {}
