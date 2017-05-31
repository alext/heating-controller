package sensor

import (
	"time"

	"github.com/alext/heating-controller/units"
)

const initialValue = 21000

type pushSensor struct {
	baseSensor
	sensorID string
}

func NewPushSensor(id string) SettableSensor {
	return &pushSensor{
		sensorID: id,
		baseSensor: baseSensor{
			temp: initialValue,
		},
	}
}

func (s *pushSensor) Close() {
	// No-Op
}

func (s *pushSensor) Set(temp units.Temperature, updatedAt time.Time) {
	s.baseSensor.set(temp, updatedAt)
}
