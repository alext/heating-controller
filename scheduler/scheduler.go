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
	Boost(time.Duration)
	NextEvent() *Event
}

type commandType int

const (
	stopCommand commandType = iota
	addEventCommand
	boostCommand
)

type command struct {
	cmdType commandType
	e       *Event
}

type scheduler struct {
	out       output.Output
	events    eventList
	running   bool
	lock      sync.Mutex
	commandCh chan command
}

func New(out output.Output) Scheduler {
	return &scheduler{
		out:       out,
		events:    make(eventList, 0),
		commandCh: make(chan command),
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
		s.commandCh <- command{cmdType: stopCommand}
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
		s.commandCh <- command{cmdType: addEventCommand, e: &e}
		return
	}
	s.events = append(s.events, &e)
	sort.Sort(s.events)
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
			logger.Debugf("[Scheduler:%s] Next event at %v - %v", s.out.Id(), at, event)
		}
		select {
		case <-tmr.Channel():
			if event != nil {
				go event.do(s.out)
				event = nil
			}
		case cmd := <-s.commandCh:
			switch cmd.cmdType {
			case stopCommand:
				tmr.Stop()
				return
			case addEventCommand:
				s.events = append(s.events, cmd.e)
				sort.Sort(s.events)
				if _, e := s.next(time_Now().Local()); e == cmd.e {
					// let the new event be picked up at the top of the loop
					event = nil
				}
			case boostCommand:
				go s.out.Activate()
				now := time_Now().Local()
				boostEnd := cmd.e.nextOccurance(now)
				if event == nil || event.Action == TurnOff || boostEnd.Before(at) {
					event = cmd.e
					at = boostEnd
					tmr.Reset(at.Sub(now))
				}
			}
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
			return event.nextOccurance(now), event
		}
	}
	return s.events[0].nextOccurance(now), s.events[0]
}
