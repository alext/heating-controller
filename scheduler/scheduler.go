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
	Boost(time.Duration, func())
	CancelBoost()
	NextEvent() *Event
	ReadEvents() []Event
	Override(Event)
	CancelOverride()
}

type scheduler struct {
	id        string
	events    eventList
	running   bool
	boosted   bool
	lock      sync.Mutex
	commandCh chan func()

	nextEvent *Event
	nextAt    time.Time
	tmr       timer
}

func New(id string) Scheduler {
	return &scheduler{
		id:        id,
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
			s.nextEvent = nil // cause the next event to be recalculated
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
			s.nextEvent = nil // cause the next event to be recalculated
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

func (s *scheduler) Boost(d time.Duration, action func()) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		return
	}

	s.commandCh <- func() {
		go action()
		s.boosted = true
		if d == 0 && len(s.events) == 0 {
			d = time.Hour
		}
		if d == 0 {
			log.Printf("[Scheduler:%s] Boosting until next event", s.id)
			return
		}
		now := time_Now().Local()
		endTime := now.Add(d)
		s.nextAt = endTime
		s.nextEvent = &Event{
			Hour:   endTime.Hour(),
			Min:    endTime.Minute(),
			Action: func() { s.commandCh <- s.setCurrentState },
		}
		s.tmr.Reset(s.nextAt.Sub(now))
		log.Printf("[Scheduler:%s] Boosting until %v", s.id, s.nextAt)
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

func (s *scheduler) Override(e Event) {
	s.commandCh <- func() {
		now := time_Now().Local()
		s.nextAt = e.nextOccuranceAfter(now)
		s.nextEvent = &e
		s.tmr.Reset(s.nextAt.Sub(now))
		log.Printf("[Scheduler:%s] Override job at %v - %v", s.id, s.nextAt, s.nextEvent)
	}
}

func (s *scheduler) CancelOverride() {
	s.commandCh <- func() {
		s.nextEvent = nil
	}
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
				go s.nextEvent.Action()
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
		if e.Hour != event.Hour || e.Min != event.Min || e.Label != event.Label {
			newEvents = append(newEvents, e)
		}
	}
	s.events = newEvents
}

func (s *scheduler) setCurrentState() {
	if len(s.events) < 1 {
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
	go previous.Action()
}

func (s *scheduler) next(now time.Time) (at time.Time, e *Event) {
	if len(s.events) < 1 {
		return now.Add(24 * time.Hour), nil
	}
	hour, min, _ := now.Clock()
	for _, event := range s.events {
		if event.after(hour, min) {
			return event.nextOccuranceAfter(now), event
		}
	}
	return s.events[0].nextOccuranceAfter(now), s.events[0]
}
