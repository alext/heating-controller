package timer

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
)

// variable indirection to enable testing
var (
	time_Now   = time.Now
	time_After = time.After
)

type Action int

const (
	TurnOff Action = iota
	TurnOn
)

type Timer interface {
	Id() string
	Start()
	Stop()
	Running() bool
	OutputActive() bool
	AddEntry(hour, minute int, a Action)
}

type entry struct {
	t      *timer
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

func (e *entry) do() {
	var err error
	if e.action == TurnOn {
		logger.Infof("[Timer:%s] Activating output", e.t.out.Id())
		e.t.outputActive = true
		err = e.t.out.Activate()
	} else {
		logger.Infof("[Timer:%s] Deactivating output", e.t.out.Id())
		e.t.outputActive = false
		err = e.t.out.Deactivate()
	}
	if err != nil {
		logger.Warnf("[Timer:%s] Output error: %v", e.t.out.Id(), err)
	}
}

func (e *entry) String() string {
	if e.action == TurnOn {
		return fmt.Sprintf("%d:%d On", e.hour, e.min)
	} else {
		return fmt.Sprintf("%d:%d Off", e.hour, e.min)
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
	out          output.Output
	entries      entryList
	running      bool
	outputActive bool
	lock         sync.Mutex
	newEntry     chan *entry
	stop         chan bool
}

func New(out output.Output) Timer {
	return &timer{
		out:      out,
		entries:  make(entryList, 0),
		newEntry: make(chan *entry),
		stop:     make(chan bool),
	}
}

func (t *timer) Id() string {
	return t.out.Id()
}

func (t *timer) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.running {
		logger.Infof("[Timer:%s] Starting", t.out.Id())
		t.running = true
		go t.run()
	}
}

func (t *timer) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.running {
		logger.Infof("[Timer:%s] Stopping", t.out.Id())
		t.running = false
		t.stop <- true
	}
}

func (t *timer) Running() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.running
}

func (t *timer) OutputActive() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.running && t.outputActive
}

func (t *timer) AddEntry(hour, min int, a Action) {
	t.lock.Lock()
	defer t.lock.Unlock()
	e := &entry{t: t, hour: hour, min: min, action: a}
	logger.Debugf("[Timer:%s] Adding entry: %v", t.out.Id(), e)
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
		logger.Debugf("[Timer:%s] Next entry at %v - %v", t.out.Id(), at, entry)
		select {
		case now = <-time_After(at.Sub(now)):
			if entry != nil {
				go entry.do()
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
	previous.do()
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
