package controller

import (
	"fmt"
	"time"

	"github.com/alext/heating-controller/scheduler"
)

type Event struct {
	Hour   int    `json:"hour"`
	Min    int    `json:"min"`
	Action Action `json:"action"`
}

func (e Event) Valid() bool {
	return e.Hour >= 0 && e.Hour < 24 && e.Min >= 0 && e.Min < 60
}

func (e Event) NextOccurance() time.Time {
	return e.nextOccuranceAfter(time.Now().Local())
}

func (e Event) nextOccuranceAfter(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), e.Hour, e.Min, 0, 0, time.Local)
	if next.Before(current) {
		current = current.AddDate(0, 0, 1)
		next = time.Date(current.Year(), current.Month(), current.Day(), e.Hour, e.Min, 0, 0, time.Local)
	}
	return next
}

func (e Event) String() string {
	return fmt.Sprintf("%d:%02d %s", e.Hour, e.Min, e.Action)
}

// FIXME: temp function to ease refactoring
func (e Event) toScheduler() scheduler.Event {
	return scheduler.Event{
		Hour:   e.Hour,
		Min:    e.Min,
		Action: e.Action.toScheduler(),
	}
}

// FIXME: temp function to ease refactoring
func eventFromScheduler(se *scheduler.Event) *Event {
	if se == nil {
		return nil
	}
	e := &Event{
		Hour: se.Hour,
		Min:  se.Min,
	}
	if se.Action == scheduler.TurnOn {
		e.Action = TurnOn
	} // TurnOff is zero value
	return e
}
