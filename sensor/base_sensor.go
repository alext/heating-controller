package sensor

import (
	"log"
	"sync"
	"time"

	"github.com/alext/heating-controller/units"
)

const initialValue = 21_000

type baseSensor struct {
	name          string
	id            string
	lock          sync.RWMutex
	temp          units.Temperature
	updatedAt     time.Time
	subscriptions []chan units.Temperature
}

func newBaseSensor(name, id string) baseSensor {
	s := baseSensor{
		name: name,
		id:   id,
		temp: initialValue,
	}
	return s
}

func (s *baseSensor) ID() string {
	return s.id
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
	log.Printf("[Sensor:%s] updated to %s, (updatedAt: %s)", s.name, temp, updatedAt)

	for _, ch := range s.subscriptions {
		select {
		case ch <- s.temp:
		default:
		}
	}
}
