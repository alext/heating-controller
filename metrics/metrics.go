package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/sensor"
)

type Metrics struct {
	sensors map[string]sensor.Sensor
	lock    sync.RWMutex
}

func New() (*Metrics, error) {
	m := &Metrics{
		sensors: make(map[string]sensor.Sensor),
	}
	err := prometheus.Register(m)
	return m, err
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.Handler()
}

func (m *Metrics) AddSensor(name string, s sensor.Sensor) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.sensors[name] = s
	return
}

func (m *Metrics) AddZone(z *controller.Zone) {
	g := prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Namespace:   "house",
			Name:        "zone_active",
			ConstLabels: prometheus.Labels{"id": z.ID},
		},
		func() float64 {
			if z.Active() {
				return 1
			}
			return 0
		},
	)
	prometheus.MustRegister(g)
}
