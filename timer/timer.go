package timer

import (
	"sort"
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

type entry struct {
	hour   int
	min    int
	action action
}

func (e *entry) actionTime(actionDate time.Time) time.Time {
	year, month, day := actionDate.Date()
	return time.Date(year, month, day, e.hour, e.min, 0, 0, time.Local)
}

func (e *entry) do(out output.Output) {
	if e.action == TurnOn {
		out.Activate()
	} else {
		out.Deactivate()
	}
}

type byTime []*entry

func (a byTime) Len() int      { return len(a) }
func (a byTime) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byTime) Less(i, j int) bool {
	return a[i].hour < a[j].hour || (a[i].hour == a[j].hour && a[i].min < a[j].min)
}

type timer struct {
	out     output.Output
	entries []*entry
	running bool
	lock    sync.Mutex
	stop    chan bool
}

func New(out output.Output) Timer {
	return &timer{
		out:     out,
		entries: make([]*entry, 0),
		stop:    make(chan bool),
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
	e := &entry{hour: hour, min: min, action: a}
	t.entries = append(t.entries, e)
	sort.Sort(byTime(t.entries))
}

func (t *timer) run() {
	for {
		now := time_Now().Local()
		at, entry := t.next(now)
		select {
		case now = <-time_After(at.Sub(now)):
			if entry != nil {
				go entry.do(t.out)
			}
		case <-t.stop:
			return
		}
	}
}

func (t *timer) next(now time.Time) (at time.Time, e *entry) {
	if len(t.entries) < 1 {
		return now.AddDate(0, 0, 1), nil
	}
	hour, min, _ := now.Clock()
	for _, entry := range t.entries {
		if entry.hour > hour || (entry.hour == hour && entry.min > min) {
			return entry.actionTime(now), entry
		}
	}
	return t.entries[0].actionTime(now.AddDate(0, 0, 1)), t.entries[0]
}
