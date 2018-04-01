package controller

import (
	"sort"
	"sync"
	"time"

	"github.com/alext/heating-controller/scheduler"
)

//go:generate counterfeiter . EventHandler
type EventHandler interface {
	AddEvent(Event) error
	RemoveEvent(Event)
	ReadEvents() []Event
	NextEvent() *Event

	Boost(time.Duration)
	CancelBoost()
	Boosted() bool
}

type eventHandler struct {
	lock    sync.RWMutex
	events  eventList
	demand  func(Event)
	sched   scheduler.Scheduler
	boosted bool
}

func NewEventHandler(s scheduler.Scheduler, demand func(Event)) EventHandler {
	return &eventHandler{
		sched:  s,
		demand: demand,
		events: make(eventList, 0),
	}
}

func (eh *eventHandler) trigger(e Event) {
	eh.lock.Lock()
	defer eh.lock.Unlock()
	eh.boosted = false
	eh.demand(e)
}

// nextEvent returns the next regular event, ignoring any overrides. This is
// different from NextEvent, which queries the scheduler.
func (eh *eventHandler) nextEvent() *Event {
	if len(eh.events) < 1 {
		return nil
	}
	hour, min, _ := timeNow().Local().Clock()
	for _, e := range eh.events {
		if e.after(hour, min) {
			return &e
		}
	}
	return &eh.events[0]
}

func (eh *eventHandler) previousEvent() *Event {
	if len(eh.events) < 1 {
		return nil
	}
	hour, min, _ := timeNow().Local().Clock()
	for i, e := range eh.events {
		if e.after(hour, min) {
			if i > 0 {
				return &eh.events[i-1]
			}
			break
		}
	}
	return &eh.events[len(eh.events)-1]
}

func (eh *eventHandler) AddEvent(e Event) error {
	if !e.Valid() {
		return ErrInvalidEvent
	}
	eh.lock.Lock()
	defer eh.lock.Unlock()

	eh.events = append(eh.events, e)
	sort.Sort(eh.events)

	return eh.sched.AddJob(e.buildSchedulerJob(eh.trigger))
}

func (eh *eventHandler) RemoveEvent(e Event) {
	eh.lock.Lock()
	defer eh.lock.Unlock()

	newEvents := make(eventList, 0)
	for _, ee := range eh.events {
		if ee != e {
			newEvents = append(newEvents, ee)
		}
	}
	eh.events = newEvents

	eh.sched.RemoveJob(e.buildSchedulerJob(eh.demand))
}

func (eh *eventHandler) NextEvent() *Event {
	j := eh.sched.NextJob()
	if j == nil {
		return nil
	}
	e := &Event{
		Hour: j.Hour,
		Min:  j.Min,
	}
	if j.Label == "On" {
		e.Action = TurnOn
	}
	return e
}

func (eh *eventHandler) ReadEvents() []Event {
	eh.lock.RLock()
	defer eh.lock.RUnlock()

	events := make([]Event, 0, len(eh.events))
	for _, e := range eh.events {
		events = append(events, e)
	}
	return events
}

func (eh *eventHandler) Boosted() bool {
	eh.lock.RLock()
	defer eh.lock.RUnlock()
	return eh.boosted
}

func (eh *eventHandler) Boost(d time.Duration) {
	eh.lock.Lock()
	defer eh.lock.Unlock()
	eh.boosted = true
	eh.demand(Event{Action: TurnOn})

	if d == 0 {
		return
	}

	endTime := timeNow().Local().Add(d)
	endEvent := Event{
		Hour:   endTime.Hour(),
		Min:    endTime.Minute(),
		Action: TurnOff,
	}

	nextEvent := eh.nextEvent()

	if nextEvent == nil || endEvent.NextOccurance().Before(nextEvent.NextOccurance()) {
		eh.sched.Override(endEvent.buildSchedulerJob(eh.trigger))
	}
}

func (eh *eventHandler) CancelBoost() {
	eh.lock.Lock()
	defer eh.lock.Unlock()
	if !eh.boosted {
		return
	}
	eh.boosted = false
	eh.sched.CancelOverride()

	previous := eh.previousEvent()
	if previous != nil {
		eh.demand(*previous)
	} else {
		eh.demand(Event{Action: TurnOff})
	}
}
