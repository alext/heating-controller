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
	registry *prometheus.Registry
	sensors  map[string]sensor.Sensor
	lock     sync.RWMutex
}

func New() *Metrics {
	m := &Metrics{
		registry: prometheus.NewRegistry(),
		sensors:  make(map[string]sensor.Sensor),
	}
	m.registry.MustRegister(m)
	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
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
	m.registry.MustRegister(g)
}
