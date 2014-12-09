package scheduler

import (
	"sort"
	"sync"
	"time"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
)

// variable indirection to enable testing
var time_Now = time.Now

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

type Scheduler interface {
	Start()
	Stop()
	Running() bool
	AddEvent(Event)
	NextEvent() *Event
}

type scheduler struct {
	out      output.Output
	events   eventList
	running  bool
	lock     sync.Mutex
	newEvent chan *Event
	stop     chan bool
}

func New(out output.Output) Scheduler {
	return &scheduler{
		out:      out,
		events:   make(eventList, 0),
		newEvent: make(chan *Event),
		stop:     make(chan bool),
	}
}

func (t *scheduler) Start() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if !t.running {
		logger.Infof("[Scheduler:%s] Starting", t.out.Id())
		t.running = true
		go t.run()
	}
}

func (t *scheduler) Stop() {
	t.lock.Lock()
	defer t.lock.Unlock()
	if t.running {
		logger.Infof("[Scheduler:%s] Stopping", t.out.Id())
		t.running = false
		t.stop <- true
	}
}

func (t *scheduler) Running() bool {
	t.lock.Lock()
	defer t.lock.Unlock()
	return t.running
}

func (t *scheduler) AddEvent(e Event) {
	t.lock.Lock()
	defer t.lock.Unlock()
	logger.Debugf("[Scheduler:%s] Adding event: %v", t.out.Id(), e)
	if t.running {
		t.newEvent <- &e
		return
	}
	t.events = append(t.events, &e)
	sort.Sort(t.events)
}

func (t *scheduler) NextEvent() *Event {
	t.lock.Lock()
	defer t.lock.Unlock()
	_, nextEvent := t.next(time_Now().Local())
	return nextEvent
}

func (t *scheduler) run() {
	t.setInitialState()
	var event *Event
	var at time.Time
	tmr := newClockTimer(100 * time.Hour) // arbitrary duration that will be reset in the loop
	for {
		if event == nil {
			now := time_Now().Local()
			at, event = t.next(now)
			tmr.Reset(at.Sub(now))
			logger.Debugf("[Scheduler:%s] Next entry at %v - %v", t.out.Id(), at, event)
		}
		select {
		case <-tmr.Channel():
			if event != nil {
				go event.do(t.out)
				event = nil
			}
		case newEvent := <-t.newEvent:
			t.events = append(t.events, newEvent)
			sort.Sort(t.events)
			if _, e := t.next(time_Now().Local()); e == newEvent {
				// let the new event be picked up at the top of the loop
				event = nil
			}
		case <-t.stop:
			tmr.Stop()
			return
		}
	}
}

func (t *scheduler) setInitialState() {
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

func (t *scheduler) next(now time.Time) (at time.Time, e *Event) {
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
