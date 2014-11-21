package zone

import (
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/timer"
)

type Zone struct {
	ID    string
	Out   output.Output
	Timer timer.Timer
}

func New(id string, out output.Output) *Zone {
	return &Zone{
		ID:    id,
		Out:   out,
		Timer: timer.New(out),
	}
}
