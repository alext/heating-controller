package scheduler

import "time"

// variable indirection to enable testing
var newTimer = newRealTimer

// Interface to specify a wrapper around time.Timer in order to alow the
// substitution of another implementation in the tests.
type timer interface {
	Reset(time.Duration) bool
	Stop() bool
	Channel() <-chan time.Time
}

type realTimer struct {
	*time.Timer
}

func newRealTimer(d time.Duration) timer {
	return realTimer{time.NewTimer(d)}
}

func (tmr realTimer) Channel() <-chan time.Time {
	return tmr.Timer.C
}
