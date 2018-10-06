package units

import (
	"fmt"
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
