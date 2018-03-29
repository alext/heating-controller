package controller

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/thermostat"
	"github.com/alext/heating-controller/units"
)

var ErrInvalidEvent = errors.New("invalid event")

type Zone struct {
	ID         string
	Scheduler  scheduler.Scheduler
	Thermostat thermostat.Thermostat

	lock          sync.RWMutex
	events        eventList
	Out           output.Output
	schedDemand   bool
	thermDemand   bool
	currentDemand bool
}

func NewZone(id string, out output.Output) *Zone {
	z := &Zone{
		ID:          id,
		events:      make(eventList, 0),
		Out:         out,
		thermDemand: true, // always on until a thermostat is added
	}
	z.Scheduler = scheduler.New(z.ID, z.schedulerDemand)
	return z
}

func (z *Zone) AddEvent(e Event) error {
	if !e.Valid() {
		return ErrInvalidEvent
	}
	z.lock.Lock()
	defer z.lock.Unlock()

	z.events = append(z.events, e)
	sort.Sort(z.events)

	return z.Scheduler.AddEvent(e.toScheduler())
}

func (z *Zone) RemoveEvent(e Event) {
	z.lock.Lock()
	defer z.lock.Unlock()

	newEvents := make(eventList, 0)
	for _, ze := range z.events {
		if ze != e {
			newEvents = append(newEvents, ze)
		}
	}
	z.events = newEvents

	z.Scheduler.RemoveEvent(e.toScheduler())
}

func (z *Zone) NextEvent() *Event {
	e := eventFromScheduler(z.Scheduler.NextEvent())
	if e == nil {
		return nil
	}
	return e
}

func (z *Zone) ReadEvents() []Event {
	z.lock.RLock()
	defer z.lock.RUnlock()

	events := make([]Event, 0, len(z.events))
	for _, e := range z.events {
		events = append(events, e)
	}
	return events
}

func (z *Zone) Boosted() bool {
	return z.Scheduler.Boosted()
}

func (z *Zone) Boost(d time.Duration) {
	z.Scheduler.Boost(d)
}

func (z *Zone) CancelBoost() {
	z.Scheduler.CancelBoost()
}

func (z *Zone) SetupThermostat(source sensor.Sensor, initialTarget units.Temperature) {
	z.Thermostat = thermostat.New(z.ID, source, initialTarget, z.thermostatDemand)
}

func (z *Zone) Active() (bool, error) {
	z.lock.RLock()
	defer z.lock.RUnlock()
	return z.Out.Active()
}

func (z *Zone) SDemand() bool {
	z.lock.RLock()
	defer z.lock.RUnlock()
	return z.schedDemand
}

func (z *Zone) TDemand() bool {
	z.lock.RLock()
	defer z.lock.RUnlock()
	return z.thermDemand
}

func (z *Zone) schedulerDemand(a scheduler.Action) {
	z.lock.Lock()
	defer z.lock.Unlock()
	z.schedDemand = a == scheduler.TurnOn
	log.Printf("[Zone:%s] received scheduler demand : %t", z.ID, z.schedDemand)
	z.updateDemand()
}

func (z *Zone) thermostatDemand(demand bool) {
	z.lock.Lock()
	defer z.lock.Unlock()
	z.thermDemand = demand
	log.Printf("[Zone:%s] received thermostat demand : %t", z.ID, z.thermDemand)
	z.updateDemand()
}

// Must be called with the lock held for writing.
func (z *Zone) updateDemand() {
	targetDemand := z.schedDemand && z.thermDemand
	if targetDemand == z.currentDemand {
		// No change needed
		return
	}

	var err error
	if targetDemand {
		log.Printf("[Zone:%s] Activating output", z.ID)
		err = z.Out.Activate()
	} else {
		log.Printf("[Zone:%s] Deactivating output", z.ID)
		err = z.Out.Deactivate()
	}
	if err != nil {
		log.Printf("[Zone:%s] Output error: %v", z.ID, err)
	}
	z.currentDemand = targetDemand
}
