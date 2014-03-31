package timer

import (
	"sync"
	"time"

	"github.com/alext/heating-controller/output"
)

// variable indirection to enable testing
var (
	time_Now   = time.Now
	time_After = time.After
)

type action int

const (
	TurnOn action = iota
	TurnOff
)

type Timer interface {
	Start()
	Stop()
	Running() bool
	AddEntry(hour, minute int, a action)
}

type timer struct {
	out     output.Output
	running bool
	lock    sync.Mutex
	stop    chan bool
}

func New(out output.Output) Timer {
	return &timer{
		out:  out,
		stop: make(chan bool),
	}
}

func (t *timer) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.running {
		t.running = true
		go t.run()
	}
}

func (t *timer) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.running {
		t.running = false
		t.stop <- true
	}
}

func (t *timer) Running() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.running
}

func (t *timer) AddEntry(hour, min int, a action) {
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
	action action
}

func (e *event) actions(t *timer, actionDate time.Time) (at time.Time, do func()) {
	year, month, day := actionDate.Date()
	at = time.Date(year, month, day, e.hour, e.min, 0, 0, time.Local)
	if e.action == TurnOn {
		do = t.activate
	} else {
		do = t.deactivate
	}
	return
}

var events = [...]event{
	{6, 30, TurnOn},
	{7, 30, TurnOff},
	{17, 00, TurnOn},
	{21, 00, TurnOff},
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
