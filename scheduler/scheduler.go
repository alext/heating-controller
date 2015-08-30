package scheduler

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"
)

// variable indirection to enable testing
var time_Now = time.Now

var ErrInvalidEvent = errors.New("invalid event")

type Scheduler interface {
	Start()
	Stop()
	Running() bool
	AddEvent(Event) error
	RemoveEvent(Event)
	Boosted() bool
	Boost(time.Duration)
	CancelBoost()
	NextEvent() *Event
	ReadEvents() []Event
}

type demandFunc func(Action)

type commandType uint8

const (
	stopCommand commandType = iota
	addEventCommand
	removeEventCommand
	nextEventCommand
	readEventsCommand
	boostCommand
	cancelBoostCommand
)

type command struct {
	cmdType commandType
	e       *Event
}

type scheduler struct {
	id        string
	demand    demandFunc
	events    eventList
	running   bool
	boosted   bool
	lock      sync.Mutex
	commandCh chan command
}

func New(zoneID string, df demandFunc) Scheduler {
	return &scheduler{
		id:        zoneID,
		demand:    df,
		events:    make(eventList, 0),
		commandCh: make(chan command),
	}
}

func (s *scheduler) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		log.Printf("[Scheduler:%s] Starting", s.id)
		s.running = true
		go s.run()
	}
}

func (s *scheduler) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		log.Printf("[Scheduler:%s] Stopping", s.id)
		s.running = false
		s.commandCh <- command{cmdType: stopCommand}
	}
}

func (s *scheduler) Running() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.running
}

func (s *scheduler) AddEvent(e Event) error {
	if !e.Valid() {
		return ErrInvalidEvent
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Printf("[Scheduler:%s] Adding event: %v", s.id, e)
	if s.running {
		s.commandCh <- command{cmdType: addEventCommand, e: &e}
		return nil
	}
	s.addEvent(&e)
	return nil
}

func (s *scheduler) RemoveEvent(e Event) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- command{cmdType: removeEventCommand, e: &e}
		return
	}
	s.removeEvent(&e)
}

func (s *scheduler) Boosted() bool {
	return s.boosted
}

func (s *scheduler) Boost(d time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		endTime := time_Now().Local().Add(d)
		s.commandCh <- command{cmdType: boostCommand, e: &Event{
			Hour:   endTime.Hour(),
			Min:    endTime.Minute(),
			Action: TurnOff,
		}}
	}
}

func (s *scheduler) CancelBoost() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- command{cmdType: cancelBoostCommand}
	}
}

func (s *scheduler) NextEvent() *Event {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- command{cmdType: nextEventCommand}
		cmd := <-s.commandCh
		return cmd.e
	}
	_, nextEvent := s.next(time_Now().Local())
	return nextEvent
}

func (s *scheduler) ReadEvents() []Event {
	result := make([]Event, 0)

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- command{cmdType: readEventsCommand}
		for cmd := range s.commandCh {
			if cmd.e == nil {
				break
			}
			result = append(result, *cmd.e)
		}
	} else {
		for _, e := range s.events {
			result = append(result, *e)
		}
	}
	return result
}

func (s *scheduler) run() {
	s.setCurrentState()
	var event *Event
	var at time.Time
	tmr := newTimer(100 * time.Hour) // arbitrary duration that will be reset in the loop
	for {
		if event == nil {
			now := time_Now().Local()
			at, event = s.next(now)
			tmr.Reset(at.Sub(now))
			s.boosted = false
			log.Printf("[Scheduler:%s] Next event at %v - %v", s.id, at, event)
		}
		select {
		case <-tmr.Channel():
			if event != nil {
				go s.demand(event.Action)
				event = nil
			}
		case cmd := <-s.commandCh:
			switch cmd.cmdType {
			case stopCommand:
				tmr.Stop()
				return
			case addEventCommand:
				s.addEvent(cmd.e)
				if _, e := s.next(time_Now().Local()); e == cmd.e {
					// let the new event be picked up at the top of the loop
					event = nil
				}
			case removeEventCommand:
				s.removeEvent(cmd.e)
				if *cmd.e == *event {
					// let the new event be picked up at the top of the loop
					event = nil
				}
			case nextEventCommand:
				cmd.e = event
				s.commandCh <- cmd
			case readEventsCommand:
				for _, e := range s.events {
					s.commandCh <- command{e: e}
				}
				s.commandCh <- command{}
			case boostCommand:
				go s.demand(TurnOn)
				s.boosted = true
				now := time_Now().Local()
				boostEnd := cmd.e.nextOccuranceAfter(now)
				if event == nil || event.Action == TurnOff || boostEnd.Before(at) {
					event = cmd.e
					at = boostEnd
					tmr.Reset(at.Sub(now))
					log.Printf("[Scheduler:%s] Boosting until %v", s.id, at)
				} else {
					log.Printf("[Scheduler:%s] Boosting until next event", s.id)
				}
			case cancelBoostCommand:
				log.Printf("[Scheduler:%s] Cancelling boost", s.id)
				s.setCurrentState()
				event = nil
			}
		}
	}
}

func (s *scheduler) addEvent(e *Event) {
	s.events = append(s.events, e)
	sort.Sort(s.events)
}

func (s *scheduler) removeEvent(event *Event) {
	newEvents := make(eventList, 0)
	for _, e := range s.events {
		if *e != *event {
			newEvents = append(newEvents, e)
		}
	}
	s.events = newEvents
}

func (s *scheduler) setCurrentState() {
	if len(s.events) < 1 {
		s.demand(TurnOff)
		return
	}
	hour, min, _ := time_Now().Local().Clock()
	var previous *Event
	for _, e := range s.events {
		if e.after(hour, min) {
			break
		}
		previous = e
	}
	if previous == nil {
		previous = s.events[len(s.events)-1]
	}
	s.demand(previous.Action)
}

func (s *scheduler) next(now time.Time) (at time.Time, e *Event) {
	if len(s.events) < 1 {
		return now.AddDate(0, 0, 1), nil
	}
	hour, min, _ := now.Clock()
	for _, event := range s.events {
		if event.after(hour, min) {
			return event.nextOccuranceAfter(now), event
		}
	}
	return s.events[0].nextOccuranceAfter(now), s.events[0]
}
