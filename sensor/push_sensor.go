package sensor

import "time"

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

func (s *pushSensor) Set(temp Temperature, updatedAt time.Time) {
	s.baseSensor.set(temp, updatedAt)
}
