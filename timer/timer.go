package timer

import (
	"time"

	"github.com/alext/heating-controller/output"
)

// variable indirection to enable testing
var (
	time_Now   = time.Now
	time_After = time.After
)

type Timer interface {
	Start()
	Stop()
}

type timer struct {
	out  output.Output
	stop chan bool
}

func New(out output.Output) Timer {
	return &timer{
		out:  out,
		stop: make(chan bool),
	}
}

func (t *timer) Start() {
	go t.run()
}

func (t *timer) Stop() {
	t.stop <- true
}

func (t *timer) run() {
	for {
		now := time_Now().Local()
		at, do := t.next(now)
		select {
		case now = <-time_After(at.Sub(now)):
			go do()
		case <-t.stop:
			return
		}
	}
}

type event struct {
	hour   int
	min    int
	active bool
}

func (e *event) actions(t *timer, actionDate time.Time) (at time.Time, do func()) {
	year, month, day := actionDate.Date()
	at = time.Date(year, month, day, e.hour, e.min, 0, 0, time.Local)
	if e.active {
		do = t.activate
	} else {
		do = t.deactivate
	}
	return
}

var events = [...]event{
	{6, 30, true},
	{7, 30, false},
	{17, 00, true},
	{21, 00, false},
}

func (t *timer) next(now time.Time) (at time.Time, do func()) {
	hour, min, _ := now.Clock()
	for _, event := range events {
		if event.hour > hour || (event.hour == hour && event.min > min) {
			return event.actions(t, now)
		}
	}
	return events[0].actions(t, now.AddDate(0, 0, 1))
}

func (t *timer) activate() {
	t.out.Activate()
}

func (t *timer) deactivate() {
	t.out.Deactivate()
}
