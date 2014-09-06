package output

import (
	"sync"

	"github.com/alext/gpio"
)

// Variable indirection to facilitate testing.
var pinOpener = gpio.OpenPin

type Output interface {
	Id() string
	Active() (bool, error)
	Activate() error
	Deactivate() error
	Close() error
}

type output struct {
	id  string
	pin gpio.Pin
	mu  sync.Mutex
}

func New(id string, pinNo int) (out Output, err error) {
	pin, err := pinOpener(pinNo, gpio.ModeOutput)
	if err != nil {
		return nil, err
	}
	return &output{id: id, pin: pin}, nil
}

func (out *output) Id() string {
	return out.id
}

func (out *output) Active() (bool, error) {
	out.mu.Lock()
	defer out.mu.Unlock()
	return out.pin.Get()
}

func (out *output) Activate() error {
	out.mu.Lock()
	defer out.mu.Unlock()
	return out.pin.Set()
}

func (out *output) Deactivate() error {
	out.mu.Lock()
	defer out.mu.Unlock()
	return out.pin.Clear()
}

func (out *output) Close() error {
	out.mu.Lock()
	defer out.mu.Unlock()
	return out.pin.Close()
}
