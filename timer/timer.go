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

func (a Action) String() string {
	if a == TurnOn {
		return "On"
	}
	return "Off"
}

const (
	TurnOff Action = iota
	TurnOn
)

type Timer interface {
	Start()
	Stop()
	Running() bool
	AddEvent(Event)
	NextEvent() *Event
}

type Event struct {
	Hour   int
	Min    int
	Action Action
}

func (e *Event) actionTime(actionDate time.Time) time.Time {
	year, month, day := actionDate.Date()
	return time.Date(year, month, day, e.Hour, e.Min, 0, 0, time.Local)
}

func (e *Event) after(hour, min int) bool {
	return e.Hour > hour || (e.Hour == hour && e.Min > min)
}

func (e *Event) do(out output.Output) {
	var err error
	if e.Action == TurnOn {
		logger.Infof("[Timer:%s] Activating output", out.Id())
		err = out.Activate()
	} else {
		logger.Infof("[Timer:%s] Deactivating output", out.Id())
		err = out.Deactivate()
	}
	if err != nil {
		logger.Warnf("[Timer:%s] Output error: %v", out.Id(), err)
	}
}

func (e *Event) String() string {
	return fmt.Sprintf("%d:%d %s", e.Hour, e.Min, e.Action)
}

type eventList []*Event

func (el eventList) Len() int      { return len(el) }
func (el eventList) Swap(i, j int) { el[i], el[j] = el[j], el[i] }
func (el eventList) Less(i, j int) bool {
	a, b := el[i], el[j]
	return a.Hour < b.Hour || (a.Hour == b.Hour && a.Min < b.Min)
}

type timer struct {
	out      output.Output
	events   eventList
	running  bool
	lock     sync.Mutex
	newEvent chan *Event
	stop     chan bool
}

func New(out output.Output) Timer {
	return &timer{
		out:      out,
		events:   make(eventList, 0),
		newEvent: make(chan *Event),
		stop:     make(chan bool),
	}
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

func (t *timer) AddEvent(e Event) {
	t.lock.Lock()
	defer t.lock.Unlock()
	logger.Debugf("[Timer:%s] Adding event: %v", t.out.Id(), e)
	if t.running {
		t.newEvent <- &e
		return
	}
	t.events = append(t.events, &e)
	sort.Sort(t.events)
}

func (t *timer) NextEvent() *Event {
	t.lock.Lock()
	defer t.lock.Unlock()
	_, nextEvent := t.next(time_Now().Local())
	return nextEvent
}

func (t *timer) run() {
	t.setInitialState()
	for {
		now := time_Now().Local()
		at, event := t.next(now)
		logger.Debugf("[Timer:%s] Next entry at %v - %v", t.out.Id(), at, event)
		select {
		case now = <-time_After(at.Sub(now)):
			if event != nil {
				go event.do(t.out)
			}
		case event = <-t.newEvent:
			t.events = append(t.events, event)
			sort.Sort(t.events)
		case <-t.stop:
			return
		}
	}
}

func (t *timer) setInitialState() {
	if len(t.events) < 1 {
		return
	}
	hour, min, _ := time_Now().Local().Clock()
	var previous *Event
	for _, e := range t.events {
		if e.after(hour, min) {
			break
		}
		previous = e
	}
	if previous == nil {
		previous = t.events[len(t.events)-1]
	}
	previous.do(t.out)
}

func (t *timer) next(now time.Time) (at time.Time, e *Event) {
	if len(t.events) < 1 {
		return now.AddDate(0, 0, 1), nil
	}
	hour, min, _ := now.Clock()
	for _, event := range t.events {
		if event.after(hour, min) {
			return event.actionTime(now), event
		}
	}
	return t.events[0].actionTime(now.AddDate(0, 0, 1)), t.events[0]
}
