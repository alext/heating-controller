package controller

import (
	"sort"
	"sync"
	"time"

	"github.com/alext/heating-controller/scheduler"
)

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
	lock   sync.RWMutex
	events eventList
	demand func(Event)
	sched  scheduler.Scheduler
}

func NewEventHandler(s scheduler.Scheduler, demand func(Event)) EventHandler {
	return &eventHandler{
		sched:  s,
		demand: demand,
		events: make(eventList, 0),
	}
}

func (eh *eventHandler) AddEvent(e Event) error {
	if !e.Valid() {
		return ErrInvalidEvent
	}
	eh.lock.Lock()
	defer eh.lock.Unlock()

	eh.events = append(eh.events, e)
	sort.Sort(eh.events)

	return eh.sched.AddEvent(e.buildSchedulerEvent(eh.demand))
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

	eh.sched.RemoveEvent(e.buildSchedulerEvent(eh.demand))
}

func (eh *eventHandler) NextEvent() *Event {
	se := eh.sched.NextEvent()
	if se == nil {
		return nil
	}
	e := &Event{
		Hour: se.Hour,
		Min:  se.Min,
	}
	if se.Label == "On" {
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
	return eh.sched.Boosted()
}

func (eh *eventHandler) Boost(d time.Duration) {
	eh.sched.Boost(d, func() {
		eh.demand(Event{Action: TurnOn})
	})
}

func (eh *eventHandler) CancelBoost() {
	eh.sched.CancelBoost()
	// FIXME: restore previous state
	eh.demand(Event{Action: TurnOff})
}
