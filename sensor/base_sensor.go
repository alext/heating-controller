package sensor

import (
	"sync"
	"time"
)

type baseSensor struct {
	lock      sync.RWMutex
	temp      Temperature
	updatedAt time.Time
}

func (s *baseSensor) Read() (Temperature, time.Time) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.temp, s.updatedAt
}

func (s *baseSensor) set(temp Temperature, updatedAt time.Time) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.temp = temp
	s.updatedAt = updatedAt
}
