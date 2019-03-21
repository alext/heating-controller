package units

import (
	"strconv"
	"strings"
)

const unitStr = `°C`

type Temperature int

func ParseTemperature(input string) (Temperature, error) {
	input = strings.TrimSuffix(input, unitStr)
	f, err := strconv.ParseFloat(input, 64)
	if err != nil {
		return 0, err
	}
	return Temperature(f * 1000), nil
}

func (t Temperature) Float() float64 {
	return float64(t) / 1000
}

func (t Temperature) String() string {
	return strconv.FormatFloat(t.Float(), 'f', -1, 64) + unitStr
}
