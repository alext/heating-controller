package timer

import (
	"github.com/alext/heating-controller/output"
	"time"
)

type Timer interface {
	Start()
}

type timer struct {
	out output.Output
}

func New(out output.Output) Timer {
	return &timer{out: out}
}

func (t *timer) Start() {
	go t.run()
}

func (t *timer) run() {
	for {
		now := time.Now().Local()
		at, do := t.next(now)
		<- time.After(at.Sub(now))
		do()
	}
}

type event struct {
	hour   int
	min    int
	active bool
}

var events = [...]event{
	{6, 30, true},
	{7, 30, false},
	{17, 00, true},
	{21, 00, false},
}

func (t *timer) next(now time.Time) (at time.Time, do func()) {
	year, month, day := now.Date()
	hour, min, _ := now.Clock()
	var event event
	found := false
	for _, event = range events {
		if event.hour > hour || (event.hour == hour && event.min > min) {
			at = time.Date(year, month, day, event.hour, event.min, 0, 0, time.Local)
			found = true
			break
		}
	}
	if ! found {
		event = events[0]
		at = time.Date(year, month, day, event.hour, event.min, 0, 0, time.Local)
		at = at.AddDate(0, 0, 1)
	}

	if event.active {
		do = t.activate
	} else {
		do = t.deactivate
	}

	return
}

func (t *timer) activate() {
	t.out.Activate()
}

func (t *timer) deactivate() {
	t.out.Deactivate()
}
