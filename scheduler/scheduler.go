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

func (s *scheduler) Start() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if !s.running {
		logger.Infof("[Scheduler:%s] Starting", s.out.Id())
		s.running = true
		go s.run()
	}
}

func (s *scheduler) Stop() {
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.running {
		logger.Infof("[Scheduler:%s] Stopping", s.out.Id())
		s.running = false
		s.stop <- true
	}
}

func (s *scheduler) Running() bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.running
}

func (s *scheduler) AddEvent(e Event) {
	s.lock.Lock()
	defer s.lock.Unlock()
	logger.Debugf("[Scheduler:%s] Adding event: %v", s.out.Id(), e)
	if s.running {
		s.newEvent <- &e
		return
	}
	s.events = append(s.events, &e)
	sort.Sort(s.events)
}

func (s *scheduler) NextEvent() *Event {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, nextEvent := s.next(time_Now().Local())
	return nextEvent
}

func (s *scheduler) run() {
	s.setInitialState()
	var event *Event
	var at time.Time
	tmr := newTimer(100 * time.Hour) // arbitrary duration that will be reset in the loop
	for {
		if event == nil {
			now := time_Now().Local()
			at, event = s.next(now)
			tmr.Reset(at.Sub(now))
			logger.Debugf("[Scheduler:%s] Next entry at %v - %v", s.out.Id(), at, event)
		}
		select {
		case <-tmr.Channel():
			if event != nil {
				go event.do(s.out)
				event = nil
			}
		case newEvent := <-s.newEvent:
			s.events = append(s.events, newEvent)
			sort.Sort(s.events)
			if _, e := s.next(time_Now().Local()); e == newEvent {
				// let the new event be picked up at the top of the loop
				event = nil
			}
		case <-s.stop:
			tmr.Stop()
			return
		}
	}
}

func (s *scheduler) setInitialState() {
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
	previous.do(s.out)
}

func (s *scheduler) next(now time.Time) (at time.Time, e *Event) {
	if len(s.events) < 1 {
		return now.AddDate(0, 0, 1), nil
	}
	hour, min, _ := now.Clock()
	for _, event := range s.events {
		if event.after(hour, min) {
			return event.actionTime(now), event
		}
	}
	return s.events[0].actionTime(now.AddDate(0, 0, 1)), s.events[0]
}
