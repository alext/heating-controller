package sensor

import (
	"sync"
	"time"
)

const initialValue = 21000

type pushSensor struct {
	sensorID string

	mu        sync.RWMutex
	temp      Temperature
	updatedAt time.Time
}

func NewPushSensor(id string) SettableSensor {
	return &pushSensor{
		sensorID: id,
		temp:     initialValue,
	}
}

func (s *pushSensor) Read() (Temperature, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.temp, s.updatedAt
}

func (s *pushSensor) Close() {
	// No-Op
}

func (s *pushSensor) Set(temp Temperature, updatedAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.temp = temp
	s.updatedAt = updatedAt
}
