package controller

import "fmt"

type Action int8

const (
	Off Action = iota
	On
)

func (a Action) String() string {
	if a == On {
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
		*a = On
	case "Off":
		*a = Off
	default:
		return fmt.Errorf("Unrecognised action value '%s'", data)
	}
	return nil
}
