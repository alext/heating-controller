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
	registry   *prometheus.Registry
	sensors    map[string]sensor.Sensor
	sensorDesc *prometheus.Desc
	lock       sync.RWMutex
}

func newRegistry() *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.MustRegister(prometheus.NewGoCollector())
	return r
}

func New() *Metrics {
	m := &Metrics{
		registry:   newRegistry(),
		sensors:    make(map[string]sensor.Sensor),
		sensorDesc: newDensorDesc(),
	}
	m.registry.MustRegister(m)
	return m
}

func (m *Metrics) Handler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		m.registry,
		promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}),
	)
}

func (m *Metrics) AddSensor(name string, s sensor.Sensor) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.sensors[name] = s
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
