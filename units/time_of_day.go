package units

import (
	"fmt"
	"time"
)

type TimeOfDay uint

func NewTimeOfDay(hour, minute int) TimeOfDay {
	return TimeOfDay(hour*60 + minute)
}

func (t TimeOfDay) String() string {
	return fmt.Sprintf("%d:%02d", t.Hour(), t.Minute())
}

func (t TimeOfDay) Hour() int {
	return int(t / 60)
}

func (t TimeOfDay) Minute() int {
	return int(t % 60)
}

func (t TimeOfDay) NextOccuranceAfter(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), t.Hour(), t.Minute(), 0, 0, current.Location())
	if next.Before(current) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
