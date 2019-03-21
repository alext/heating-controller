package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/alext/heating-controller/sensor"
)

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for name, s := range m.sensors {
		ch <- sensorDesc(name, s)
	}
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.lock.RLock()
	defer m.lock.RUnlock()

	for name, s := range m.sensors {
		temp, ts := s.Read()

		metric, err := prometheus.NewConstMetric(sensorDesc(name, s), prometheus.GaugeValue, temp.Float())
		if err != nil {
			log.Printf("[metrics] Error constructing sensor metric for %s: %s", name, err.Error())
			continue
		}
		metric = prometheus.NewMetricWithTimestamp(ts, metric)
		ch <- metric
	}
}

func sensorDesc(name string, s sensor.Sensor) *prometheus.Desc {
	return prometheus.NewDesc(
		"temperature_celcius",
		"Current temperature in degrees Celcius",
		nil,
		prometheus.Labels{"name": name, "device_id": s.DeviceId()},
	)
}
