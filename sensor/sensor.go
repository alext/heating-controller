package sensor

import (
	"fmt"
	iofs "io/fs"
	"os"
	"time"

	"github.com/alext/heating-controller/config"
	"github.com/alext/heating-controller/units"
)

var fs iofs.FS = os.DirFS("")

type Sensor interface {
	ID() string
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
