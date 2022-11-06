package controller

import (
	"fmt"

	"github.com/alext/heating-controller/thermostat"
	"github.com/alext/heating-controller/units"
)

//go:generate go run golang.org/x/tools/cmd/stringer -type=Action

type Action uint8

const (
	Off Action = iota
	On
	SetTarget
	IncreaseTarget
	DecreaseTarget
)

func (a Action) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Action) UnmarshalText(data []byte) error {
	switch string(data) {
	case "On":
		*a = On
	case "Off":
		*a = Off
	case "SetTarget":
		*a = SetTarget
	case "IncreaseTarget":
		*a = IncreaseTarget
	case "DecreaseTarget":
		*a = DecreaseTarget
	default:
		return fmt.Errorf("Unrecognised action value '%s'", data)
	}
	return nil
}

type ThermostatAction struct {
	Action Action            `json:"action"`
	Param  units.Temperature `json:"param"`
}

func (ta ThermostatAction) String() string {
	return fmt.Sprintf("%s(%s)", ta.Action, ta.Param)
}

func (ta ThermostatAction) Apply(t thermostat.Thermostat) {
	switch ta.Action {
	case SetTarget:
		t.Set(ta.Param)
	case IncreaseTarget:
		if t.Target() < ta.Param {
			t.Set(ta.Param)
		}
	case DecreaseTarget:
		if t.Target() > ta.Param {
			t.Set(ta.Param)
		}
	}
}
