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

type Action int

const (
	TurnOn Action = iota
	TurnOff
)

type Timer interface {
	Start()
	Stop()
	Running() bool
	AddEntry(hour, minute int, a Action)
}

type entry struct {
	hour   int
	min    int
	action Action
}

func (e *entry) actionTime(actionDate time.Time) time.Time {
	year, month, day := actionDate.Date()
	return time.Date(year, month, day, e.hour, e.min, 0, 0, time.Local)
}

func (e *entry) after(hour, min int) bool {
	return e.hour > hour || (e.hour == hour && e.min > min)
}

func (e *entry) do(out output.Output) {
	if e.action == TurnOn {
		out.Activate()
	} else {
		out.Deactivate()
	}
}

type entryList []*entry

func (el entryList) Len() int      { return len(el) }
func (el entryList) Swap(i, j int) { el[i], el[j] = el[j], el[i] }
func (el entryList) Less(i, j int) bool {
	a, b := el[i], el[j]
	return a.hour < b.hour || (a.hour == b.hour && a.min < b.min)
}

type timer struct {
	out      output.Output
	entries  entryList
	running  bool
	lock     sync.Mutex
	newEntry chan *entry
	stop     chan bool
}

func New(out output.Output) Timer {
	return &timer{
		out:      out,
		entries:  make(entryList, 0),
		newEntry: make(chan *entry),
		stop:     make(chan bool),
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

func (t *timer) AddEntry(hour, min int, a Action) {
	t.lock.Lock()
	defer t.lock.Unlock()
	e := &entry{hour: hour, min: min, action: a}
	if t.running {
		t.newEntry <- e
		return
	}
	t.entries = append(t.entries, e)
}

func (t *timer) run() {
	sort.Sort(t.entries)
	t.setInitialState()
	for {
		now := time_Now().Local()
		at, entry := t.next(now)
		select {
		case now = <-time_After(at.Sub(now)):
			if entry != nil {
				go entry.do(t.out)
			}
		case entry = <-t.newEntry:
			t.entries = append(t.entries, entry)
			sort.Sort(t.entries)
		case <-t.stop:
			return
		}
	}
}

func (t *timer) setInitialState() {
	if len(t.entries) < 1 {
		return
	}
	hour, min, _ := time_Now().Local().Clock()
	var previous *entry
	for _, e := range t.entries {
		if e.after(hour, min) {
			break
		}
		previous = e
	}
	if previous == nil {
		previous = t.entries[len(t.entries)-1]
	}
	previous.do(t.out)
}

func (t *timer) next(now time.Time) (at time.Time, e *entry) {
	if len(t.entries) < 1 {
		return now.AddDate(0, 0, 1), nil
	}
	hour, min, _ := now.Clock()
	for _, entry := range t.entries {
		if entry.after(hour, min) {
			return entry.actionTime(now), entry
		}
	}
	return t.entries[0].actionTime(now.AddDate(0, 0, 1)), t.entries[0]
}
