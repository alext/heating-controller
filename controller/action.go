package controller

import "fmt"

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
