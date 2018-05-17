package controller

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/alext/heating-controller/scheduler"
)

type Event struct {
	Hour        int               `json:"hour"`
	Min         int               `json:"min"`
	Action      Action            `json:"action"`
	ThermAction *ThermostatAction `json:"therm_action,omitempty"`
}

func (e Event) Valid() bool {
	return e.Hour >= 0 && e.Hour < 24 && e.Min >= 0 && e.Min < 60
}

func (e Event) NextOccurance() time.Time {
	return e.nextOccuranceAfter(timeNow().Local())
}

func (e Event) nextOccuranceAfter(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), e.Hour, e.Min, 0, 0, time.Local)
	if next.Before(current) {
		current = current.AddDate(0, 0, 1)
		next = time.Date(current.Year(), current.Month(), current.Day(), e.Hour, e.Min, 0, 0, time.Local)
	}
	return next
}

func (e Event) after(hour, min int) bool {
	return e.Hour > hour || (e.Hour == hour && e.Min > min)
}

func (e Event) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d:%02d %s", e.Hour, e.Min, e.Action)
	if e.ThermAction != nil {
		fmt.Fprintf(&b, " %s", e.ThermAction)
	}
	return b.String()
}

func (e Event) buildSchedulerJob(demand func(Event)) scheduler.Job {
	return scheduler.Job{
		Hour:   e.Hour,
		Min:    e.Min,
		Label:  e.Action.String(),
		Action: func() { demand(e) },
	}
}

func sortEvents(events []Event) {
	sort.Slice(events, func(i, j int) bool {
		a, b := events[i], events[j]
		return a.Hour < b.Hour || (a.Hour == b.Hour && a.Min < b.Min)
	})
}
