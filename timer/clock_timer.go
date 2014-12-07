package timer

import "time"

// variable indirection to enable testing
var newClockTimer = newRealClockTimer

// Interface to specify a wrapper around time.Timer in order to alow the
// substitution of another implementation in the tests.
type clockTimer interface {
	Reset(time.Duration) bool
	Stop() bool
	Channel() <-chan time.Time
}

type realClockTimer struct {
	*time.Timer
}

func newRealClockTimer(d time.Duration) clockTimer {
	return &realClockTimer{time.NewTimer(d)}
}

func (tmr *realClockTimer) Channel() <-chan time.Time {
	return tmr.Timer.C
}
