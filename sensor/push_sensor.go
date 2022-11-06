package sensor

import (
	"time"

	"github.com/alext/heating-controller/units"
)

type pushSensor struct {
	baseSensor
}

func NewPushSensor(name, id string) SettableSensor {
	return &pushSensor{
		baseSensor: newBaseSensor(name, id),
	}
}

func (s *pushSensor) Close() {
	// No-Op
}

func (s *pushSensor) Set(temp units.Temperature, updatedAt time.Time) {
	s.baseSensor.set(temp, updatedAt)
}
