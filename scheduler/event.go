package scheduler

import (
	"fmt"
	"time"
)

const (
	Sunday    uint8 = 1 << uint8(time.Sunday)
	Monday    uint8 = 1 << uint8(time.Monday)
	Tuesday   uint8 = 1 << uint8(time.Tuesday)
	Wednesday uint8 = 1 << uint8(time.Wednesday)
	Thursday  uint8 = 1 << uint8(time.Thursday)
	Friday    uint8 = 1 << uint8(time.Friday)
	Saturday  uint8 = 1 << uint8(time.Saturday)
)

type Event struct {
	Hour   int    `json:"hour"`
	Min    int    `json:"min"`
	Action Action `json:"action"`
	Days   uint8  `json:"days"`
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

func (e Event) ActiveOn(day time.Weekday) bool {
	if e.Days == 0 {
		// No days set, always active
		return true
	}
	return (1<<uint8(day))&e.Days > 0
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
