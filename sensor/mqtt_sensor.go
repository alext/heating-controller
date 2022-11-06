package sensor

import (
	"context"
	"log"
	"time"

	"github.com/alext/heating-controller/units"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TopicSubscriber
type TopicSubscriber interface {
	Subscribe(topic string) (<-chan string, error)
}

type mqttSensor struct {
	baseSensor
	ch <-chan string

	cancel context.CancelFunc
}

func NewMQTTSensor(name, topic string, sub TopicSubscriber) (Sensor, error) {
	ch, err := sub.Subscribe(topic)
	if err != nil {
		return nil, err
	}
	s := &mqttSensor{
		// Use name as the ID as we don't really have an ID for these.
		baseSensor: newBaseSensor(name, name),
		ch:         ch,
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	go s.readLoop(ctx)
	return s, nil
}

func (s *mqttSensor) Close() {
	s.cancel()
}

func (s *mqttSensor) readLoop(ctx context.Context) {
	for {
		select {
		case msg := <-s.ch:
			s.handleMessage(msg)
		case <-ctx.Done():
			return
		}
	}
}

func (s *mqttSensor) handleMessage(msg string) {
	t, err := units.ParseTemperature(msg)
	if err != nil {
		log.Printf("[sensor:%s] Failed to parse temperature '%s': %s", s.baseSensor.name, msg, err.Error())
		return
	}
	s.baseSensor.set(t, time.Now())
}
