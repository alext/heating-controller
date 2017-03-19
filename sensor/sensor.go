package sensor

import (
	"time"

	"github.com/spf13/afero"
)

var fs afero.Fs = &afero.OsFs{}

type Sensor interface {
	Read() (Temperature, time.Time)
}

type SettableSensor interface {
	Sensor
	Set(Temperature, time.Time)
}
