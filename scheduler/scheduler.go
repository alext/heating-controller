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

type scheduler struct {
	id        string
	demand    demandFunc
	events    eventList
	running   bool
	boosted   bool
	lock      sync.Mutex
	commandCh chan func()

	nextEvent *Event
	nextAt    time.Time
	tmr       timer
}

func New(zoneID string, df demandFunc) Scheduler {
	return &scheduler{
		id:        zoneID,
		demand:    df,
		events:    make(eventList, 0),
		commandCh: make(chan func()),
	}
}

func (s *scheduler) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		log.Printf("[Scheduler:%s] Starting", s.id)
		s.tmr = newTimer(100 * time.Hour) // arbitrary duration that will be reset in the run loop
		s.running = true
		go s.run()
	}
}

func (s *scheduler) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		log.Printf("[Scheduler:%s] Stopping", s.id)
		s.commandCh <- nil
		s.tmr.Stop()
		s.running = false
	}
}

func (s *scheduler) Running() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.running
}

func (s *scheduler) AddEvent(event Event) error {
	if !event.Valid() {
		return ErrInvalidEvent
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	log.Printf("[Scheduler:%s] Adding event: %v", s.id, event)
	if s.running {
		s.commandCh <- func() {
			s.addEvent(&event)
			if _, e := s.next(time_Now().Local()); *e == event {
				// let the new event be picked up by the run loop
				s.nextEvent = nil
			}
		}
	} else {
		s.addEvent(&event)
	}
	return nil
}

func (s *scheduler) RemoveEvent(event Event) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- func() {
			s.removeEvent(&event)
			if event == *s.nextEvent {
				// let the new event be picked up by the run loop
				s.nextEvent = nil
			}
		}
	} else {
		s.removeEvent(&event)
	}
}

func (s *scheduler) Boosted() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		return false
	}

	retCh := make(chan bool, 1)
	s.commandCh <- func() {
		retCh <- s.boosted
	}
	return <-retCh
}

func (s *scheduler) Boost(d time.Duration) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		return
	}

	s.commandCh <- func() {
		go s.demand(TurnOn)
		s.boosted = true
		if d == 0 && len(s.events) == 0 {
			d = time.Hour
		}
		now := time_Now().Local()
		endTime := now.Add(d)
		if d != 0 && (s.nextEvent == nil || s.nextEvent.Action == TurnOff || endTime.Before(s.nextAt)) {
			s.nextAt = endTime
			s.nextEvent = &Event{
				Hour:   endTime.Hour(),
				Min:    endTime.Minute(),
				Action: TurnOff,
			}
			s.tmr.Reset(s.nextAt.Sub(now))
			log.Printf("[Scheduler:%s] Boosting until %v", s.id, s.nextAt)
		} else {
			log.Printf("[Scheduler:%s] Boosting until next event", s.id)
		}
	}
}

func (s *scheduler) CancelBoost() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		s.commandCh <- func() {
			log.Printf("[Scheduler:%s] Cancelling boost", s.id)
			s.setCurrentState()
			s.nextEvent = nil
		}
	}
}

func (s *scheduler) NextEvent() *Event {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		retCh := make(chan *Event, 1)
		s.commandCh <- func() {
			retCh <- s.nextEvent
		}
		return <-retCh
	}
	_, nextEvent := s.next(time_Now().Local())
	return nextEvent
}

func (s *scheduler) ReadEvents() []Event {
	result := make([]Event, 0)

	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		var wg sync.WaitGroup
		wg.Add(1)
		s.commandCh <- func() {
			for _, e := range s.events {
				result = append(result, *e)
			}
			wg.Done()
		}
		wg.Wait()
	} else {
		for _, e := range s.events {
			result = append(result, *e)
		}
	}
	return result
}

func (s *scheduler) run() {
	s.setCurrentState()
	for {
		if s.nextEvent == nil {
			now := time_Now().Local()
			s.nextAt, s.nextEvent = s.next(now)
			s.tmr.Reset(s.nextAt.Sub(now))
			s.boosted = false
			log.Printf("[Scheduler:%s] Next event at %v - %v", s.id, s.nextAt, s.nextEvent)
		}
		select {
		case <-s.tmr.Channel():
			if s.nextEvent != nil {
				go s.demand(s.nextEvent.Action)
				s.nextEvent = nil
			}
		case f := <-s.commandCh:
			if f == nil {
				// Scheduler is stopping. Exit.
				return
			}
			f()
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
