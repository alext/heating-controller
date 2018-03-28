package controller

import (
	"fmt"

	"github.com/alext/heating-controller/scheduler"
)

type Action int8

const (
	TurnOff Action = iota
	TurnOn
)

func (a Action) String() string {
	if a == TurnOn {
		return "On"
	}
	return "Off"
}

func (a Action) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

func (a *Action) UnmarshalText(data []byte) error {
	switch string(data) {
	case "On":
		*a = TurnOn
	case "Off":
		*a = TurnOff
	default:
		return fmt.Errorf("Unrecognised action value '%s'", data)
	}
	return nil
}

// FIXME: temp function to ease refactoring
func (a Action) toScheduler() scheduler.Action {
	if a == TurnOn {
		return scheduler.TurnOn
	}
	return scheduler.TurnOff
}
