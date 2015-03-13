package scheduler

import (
	"fmt"
	"time"
)

type Event struct {
	Hour   int    `json:"hour"`
	Min    int    `json:"min"`
	Action Action `json:"action"`
}

func (e Event) NextOccurance() time.Time {
	return e.nextOccuranceAfter(time_Now().Local())
}

func (e Event) Valid() bool {
	return e.Hour >= 0 && e.Hour < 24 && e.Min >= 0 && e.Min < 60
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
	return fmt.Sprintf("%d:%02d %s", e.Hour, e.Min, e.Action)
}

type eventList []*Event

func (el eventList) Len() int      { return len(el) }
func (el eventList) Swap(i, j int) { el[i], el[j] = el[j], el[i] }
func (el eventList) Less(i, j int) bool {
	a, b := el[i], el[j]
	return a.Hour < b.Hour || (a.Hour == b.Hour && a.Min < b.Min)
}
