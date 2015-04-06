package zone

import (
	"log"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
)

type OverrideMode uint8

const (
	ModeNormal OverrideMode = iota
	ModeOverrideOn
	ModeOverrideOff
)

type Zone struct {
	ID                   string
	Out                  output.Output
	Scheduler            scheduler.Scheduler
	overrideMode         OverrideMode
	schedulerDemandState scheduler.Action
}

func New(id string, out output.Output) *Zone {
	z := &Zone{
		ID:  id,
		Out: out,
	}
	z.Scheduler = scheduler.New(z.ID, z.schedulerDemand)
	return z
}

func (z *Zone) Active() (bool, error) {
	return z.Out.Active()
}

func (z *Zone) SetOverride(s OverrideMode) {
	z.overrideMode = s
	z.applyOutputState()
}

func (z *Zone) schedulerDemand(a scheduler.Action) {
	z.schedulerDemandState = a
	z.applyOutputState()
}

func (z *Zone) applyOutputState() {
	switch z.overrideMode {
	case ModeOverrideOn:
		z.setOutput(true)
	case ModeOverrideOff:
		z.setOutput(false)
	default:
		z.setOutput(z.schedulerDemandState == scheduler.TurnOn)
	}
}

func (z *Zone) setOutput(on bool) {
	var err error
	if on {
		log.Printf("[Zone:%s] Activating output", z.ID)
		err = z.Out.Activate()
	} else {
		log.Printf("[Zone:%s] Deactivating output", z.ID)
		err = z.Out.Deactivate()
	}
	if err != nil {
		log.Printf("[Zone:%s] Output error: %v", z.ID, err)
	}
}
