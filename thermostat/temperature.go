package thermostat

import "strconv"

type Temperature int

func (t Temperature) String() string {
	return strconv.FormatFloat(float64(t)/1000, 'f', -1, 64) + "°C"
}
