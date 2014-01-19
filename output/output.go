package output

import (
	"github.com/davecheney/gpio"
)

// Variable indirection to facilitate testing.
var pinOpener = gpio.OpenPin

type Output interface {
	Id() string
	Active() bool
	Activate()
	Deactivate()
	Close() error
}

type output struct {
	id  string
	pin gpio.Pin
}

func NewOutput(id string, pinNo int) (out Output, err error) {
	pin, err := pinOpener(pinNo, gpio.ModeOutput)
	if err != nil {
		return nil, err
	}
	return &output{id: id, pin: pin}, nil
}

func (out *output) Id() string {
	return out.id
}

func (out *output) Active() bool {
	return out.pin.Get()
}

func (out *output) Activate() {
	out.pin.Set()
}

func (out *output) Deactivate() {
	out.pin.Clear()
}

func (out *output) Close() error {
	return out.pin.Close()
}
