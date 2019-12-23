package thermostat

import (
	"sync"

	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/units"
)

//go:generate counterfeiter . Thermostat
type Thermostat interface {
	Current() units.Temperature
	Target() units.Temperature
	Set(units.Temperature)
	Close()
}

type demandFunc func(bool)

type thermostat struct {
	id       string
	sourceCh <-chan units.Temperature
	demand   demandFunc
	closeCh  chan struct{}

	lock    sync.RWMutex
	target  units.Temperature
	current units.Temperature
	active  bool
}

func New(id string, source sensor.Sensor, target units.Temperature, df demandFunc) Thermostat {
	initial, _ := source.Read()
	t := &thermostat{
		id:       id,
		sourceCh: source.Subscribe(),
		target:   target,
		current:  initial,
		demand:   df,
		closeCh:  make(chan struct{}),
	}

	// Set active so that a new thermostat defaults to active when within the
	// threshold.
	t.active = true
	t.trigger()

	go t.readLoop()
	return t
}

func (t *thermostat) Current() units.Temperature {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.current
}

func (t *thermostat) setCurrent(tmp units.Temperature) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.current = tmp
	t.trigger()
}

func (t *thermostat) Target() units.Temperature {
	t.lock.RLock()
	defer t.lock.RUnlock()
	return t.target
}

func (t *thermostat) Set(tmp units.Temperature) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.target = tmp
	t.trigger()
}

func (t *thermostat) Close() {
	if t.closeCh != nil {
		close(t.closeCh)
	}
}

func (t *thermostat) readLoop() {
	for {
		select {
		case tmp := <-t.sourceCh:
			t.setCurrent(tmp)
		case <-t.closeCh:
			return
		}
	}
}

const threshold = 200

// Must be called with the lock held for writing.
func (t *thermostat) trigger() {
	previousActive := t.active
	if t.current < (t.target - threshold) {
		t.active = true
	} else if t.current > t.target { // no threshold here due to hysteresis in system.
		t.active = false
	}
	if t.active != previousActive && t.demand != nil {
		go t.demand(t.active)
	}
}
