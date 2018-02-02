package sensor

import (
	"time"

	"github.com/alext/heating-controller/units"
)

const initialValue = 21000

type pushSensor struct {
	baseSensor
}

func NewPushSensor(name, id string) SettableSensor {
	return &pushSensor{
		baseSensor: baseSensor{
			name:     name,
			deviceID: id,
			temp:     initialValue,
		},
	}
}

func (s *pushSensor) Close() {
	// No-Op
}

func (s *pushSensor) Set(temp units.Temperature, updatedAt time.Time) {
	s.baseSensor.set(temp, updatedAt)
}
