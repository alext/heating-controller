package sensor

import (
	"time"

	"github.com/alext/heating-controller/units"
	"github.com/spf13/afero"
)

var fs afero.Fs = &afero.OsFs{}

type Sensor interface {
	Read() (units.Temperature, time.Time)
	Subscribe() <-chan units.Temperature
}

type SettableSensor interface {
	Sensor
	Set(units.Temperature, time.Time)
}
