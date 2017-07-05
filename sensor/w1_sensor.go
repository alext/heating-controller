package sensor

import (
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/alext/heating-controller/units"
)

const w1DevicesPath = "/sys/bus/w1/devices/"

type w1Sensor struct {
	baseSensor
	closeCh chan struct{}
}

func NewW1Sensor(deviceID string) Sensor {
	s := &w1Sensor{
		baseSensor: baseSensor{
			deviceID: deviceID,
		},
		closeCh: make(chan struct{}),
	}
	s.readTemperature(time.Now())
	go s.readLoop()
	return s
}

func (s *w1Sensor) readLoop() {
	t := newTicker(time.Minute)
	for {
		select {
		case t := <-t.Channel():
			s.readTemperature(t)
		case <-s.closeCh:
			t.Stop()
			close(s.closeCh)
			return
		}
	}
}

func (s *w1Sensor) Close() {
	s.closeCh <- struct{}{}
	<-s.closeCh
}

var temperatureRegexp = regexp.MustCompile(`t=(-?\d+)`)

func (s *w1Sensor) readTemperature(updateTime time.Time) {
	file, err := fs.Open(w1DevicesPath + s.baseSensor.deviceID + "/w1_slave")
	if err != nil {
		log.Printf("[sensor:%s] Error opening device file: %s", s.baseSensor.deviceID, err.Error())
		return
	}
	defer file.Close()
	data, err := ioutil.ReadAll(file)
	if err != nil {
		log.Printf("[sensor:%s] Error reading device: %s", s.baseSensor.deviceID, err.Error())
		return
	}
	matches := temperatureRegexp.FindStringSubmatch(string(data))
	if matches == nil {
		log.Printf("[sensor:%s] Failed to match temperature in data:\n%s", s.baseSensor.deviceID, string(data))
		return
	}

	temp, err := strconv.Atoi(matches[1])
	if err != nil {
		log.Printf("[sensor:%s] Error parsing temperature value '%s': %s", s.baseSensor.deviceID, matches[1], err.Error())
		return
	}

	s.baseSensor.set(units.Temperature(temp), updateTime)
}
