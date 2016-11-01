package webserver_test

import "github.com/alext/heating-controller/thermostat"

type dummyThermostat struct {
	Cur thermostat.Temperature
	Tgt thermostat.Temperature
}

func (t *dummyThermostat) Current() thermostat.Temperature {
	return t.Cur
}
func (t *dummyThermostat) Target() thermostat.Temperature {
	return t.Tgt
}
func (t *dummyThermostat) Set(value thermostat.Temperature) {
	t.Tgt = value
}
func (t *dummyThermostat) Close() {}
