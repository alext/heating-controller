package sensor

import (
	"sync"
	"time"
)

type baseSensor struct {
	lock          sync.RWMutex
	temp          Temperature
	updatedAt     time.Time
	subscriptions []chan Temperature
}

func (s *baseSensor) Read() (Temperature, time.Time) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.temp, s.updatedAt
}

func (s *baseSensor) Subscribe() <-chan Temperature {
	ch := make(chan Temperature, 1)
	s.lock.Lock()
	defer s.lock.Unlock()
	s.subscriptions = append(s.subscriptions, ch)
	return ch
}

func (s *baseSensor) set(temp Temperature, updatedAt time.Time) {
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
