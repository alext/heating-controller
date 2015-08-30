package zone

import (
	"log"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
)

type Zone struct {
	ID        string
	Out       output.Output
	Scheduler scheduler.Scheduler
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

func (z *Zone) schedulerDemand(a scheduler.Action) {
	var err error
	if a == scheduler.TurnOn {
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
