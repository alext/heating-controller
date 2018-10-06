package units

import (
	"fmt"
	"strings"
	"time"
)

type TimeOfDay uint

const (
	unitsPerMinute = 60
	unitsPerHour   = 60 * unitsPerMinute
	maxValidTOD    = 24*unitsPerHour - 1
)

func NewTimeOfDay(hour, minute int, sec ...int) TimeOfDay {
	t := hour*unitsPerHour + minute*unitsPerMinute
	if len(sec) > 0 {
		t += sec[0]
	}
	return TimeOfDay(t)
}

func (t TimeOfDay) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d:%02d", t.Hour(), t.Minute())
	if t.Second() > 0 {
		fmt.Fprintf(&b, ":%02d", t.Second())
	}
	return b.String()
}

func (t TimeOfDay) Hour() int {
	return int(t / unitsPerHour)
}

func (t TimeOfDay) Minute() int {
	return int(t % unitsPerHour / unitsPerMinute)
}

func (t TimeOfDay) Second() int {
	return int(t % unitsPerMinute)
}

func (t TimeOfDay) Valid() bool {
	return t <= maxValidTOD
}

func (t TimeOfDay) NextOccuranceAfter(current time.Time) time.Time {
	next := time.Date(current.Year(), current.Month(), current.Day(), t.Hour(), t.Minute(), t.Second(), 0, current.Location())
	if next.Before(current) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}
