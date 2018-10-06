package units

import (
	"fmt"
	"time"
)

type TimeOfDay uint

const (
	unitsPerHour = 60
	maxValidTOD  = 24*unitsPerHour - 1
)

func NewTimeOfDay(hour, minute int) TimeOfDay {
	return TimeOfDay(hour*unitsPerHour + minute)
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%d:%02d", t.Hour(), t.Minute())
}

func (t TimeOfDay) Hour() int {
	return int(t / unitsPerHour)
}

func (t TimeOfDay) Minute() int {
	return int(t % unitsPerHour)
}

func (t TimeOfDay) Valid() bool {
	return t <= maxValidTOD
}

func (t TimeOfDay) NextOccuranceAfter(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), t.Hour(), t.Minute(), 0, 0, current.Location())
	if next.Before(current) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
