package metrics

import (
	"log"

	"github.com/prometheus/client_golang/prometheus"
)

func newDensorDesc() *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("house", "", "temperature_celcius"),
		"Current temperature in degrees Celcius",
		[]string{"name", "device_id"},
		nil,
	)
}
func newZoneDesc() *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName("house", "heating", "zone_active"),
		"Heating zone active state - 1 or 0",
		[]string{"name"},
		nil,
	)
}

func (m *Metrics) Describe(ch chan<- *prometheus.Desc) {
	ch <- m.sensorDesc
	ch <- m.zoneDesc
}

func (m *Metrics) Collect(ch chan<- prometheus.Metric) {
	m.collectSensors(ch)
	m.collectZones(ch)
}

func (m *Metrics) collectSensors(ch chan<- prometheus.Metric) {
	for name, s := range m.ctrl.SensorsByName {
		temp, ts := s.Read()
		if ts.IsZero() {
			// sensor hasn't had a reading, so it returning initial values
			continue
		}

		metric, err := prometheus.NewConstMetric(m.sensorDesc, prometheus.GaugeValue, temp.Float(), name, s.DeviceId())
		if err != nil {
			log.Printf("[metrics] Error constructing sensor metric for %s: %s", name, err.Error())
			continue
		}
		metric = prometheus.NewMetricWithTimestamp(ts, metric)
		ch <- metric
	}
}

func (m *Metrics) collectZones(ch chan<- prometheus.Metric) {
	for _, z := range m.ctrl.Zones {
		var val float64 = 0
		if z.Active() {
			val = 1
		}
		metric, err := prometheus.NewConstMetric(m.zoneDesc, prometheus.GaugeValue, val, z.ID)
		if err != nil {
			log.Printf("[metrics] Error constructing zone metric for %s: %s", z.ID, err.Error())
			continue
		}
		ch <- metric
	}
}
