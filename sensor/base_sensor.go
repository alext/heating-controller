package sensor

import (
	"sync"
	"time"

	"github.com/alext/heating-controller/units"
)

type baseSensor struct {
	deviceID      string
	lock          sync.RWMutex
	temp          units.Temperature
	updatedAt     time.Time
	subscriptions []chan units.Temperature
}

func (s *baseSensor) DeviceId() string {
	return s.deviceID
}

func (s *baseSensor) Read() (units.Temperature, time.Time) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.temp, s.updatedAt
}

func (s *baseSensor) Subscribe() <-chan units.Temperature {
	ch := make(chan units.Temperature, 1)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.subscriptions = append(s.subscriptions, ch)
	return ch
}

func (s *baseSensor) set(temp units.Temperature, updatedAt time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.temp = temp
	s.updatedAt = updatedAt
	go s.notifySubscribers()
}

func (s *baseSensor) notifySubscribers() {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, ch := range s.subscriptions {
		select {
		case ch <- s.temp:
		default:
		}
	}
}
