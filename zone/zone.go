package zone

import (
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
)

type Zone struct {
	ID        string
	Out       output.Output
	Scheduler scheduler.Scheduler
}

func New(id string, out output.Output) *Zone {
	return &Zone{
		ID:        id,
		Out:       out,
		Scheduler: scheduler.New(out),
	}
}

func (z *Zone) Active() (bool, error) {
	return z.Out.Active()
}
