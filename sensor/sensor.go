package sensor

import (
	"fmt"
	"time"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/units"
	"github.com/spf13/afero"
)

var fs afero.Fs = &afero.OsFs{}

type Sensor interface {
	DeviceId() string
	Read() (units.Temperature, time.Time)
	Subscribe() <-chan units.Temperature
}

type SettableSensor interface {
	Sensor
	Set(units.Temperature, time.Time)
}

func New(name string, cfg config.SensorConfig) (Sensor, error) {
	switch cfg.Type {
	case "w1":
		return NewW1Sensor(name, cfg.ID), nil
	case "push":
		return NewPushSensor(name, cfg.ID), nil
	default:
		return nil, fmt.Errorf("Unrecognised sensor type: '%s'", cfg.Type)
	}
}
