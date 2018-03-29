package controller

import (
	"errors"
	"log"
	"sync"

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
	EventHandler

	lock          sync.RWMutex
	Out           output.Output
	schedDemand   bool
	thermDemand   bool
	currentDemand bool
}

func NewZone(id string, out output.Output) *Zone {
	z := &Zone{
		ID:          id,
		Out:         out,
		thermDemand: true, // always on until a thermostat is added
	}
	z.Scheduler = scheduler.New(z.ID)
	z.EventHandler = NewEventHandler(z.Scheduler, z.schedulerDemand)
	return z
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

func (z *Zone) schedulerDemand(demand bool) {
	z.lock.Lock()
	defer z.lock.Unlock()
	z.schedDemand = demand
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
