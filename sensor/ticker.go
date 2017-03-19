package sensor

import "time"

var newTicker = newRealTicker

// Interface to specify a wrapper around time.Ticker in order to alow the
// substitution of another implementation in the tests.
type ticker interface {
	Channel() <-chan time.Time
	Stop()
}

type realTicker struct {
	*time.Ticker
}

func newRealTicker(d time.Duration) ticker {
	return realTicker{time.NewTicker(d)}
}

func (t realTicker) Channel() <-chan time.Time {
	return t.Ticker.C
}
