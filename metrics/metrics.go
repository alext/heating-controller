package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/alext/heating-controller/controller"
)

type Metrics struct {
	ctrl       *controller.Controller
	registry   *prometheus.Registry
	sensorDesc *prometheus.Desc
}

func newRegistry() *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.MustRegister(prometheus.NewGoCollector())
	return r
}

func New(ctrl *controller.Controller) *Metrics {
	m := &Metrics{
		ctrl:       ctrl,
		registry:   newRegistry(),
		sensorDesc: newDensorDesc(),
	}
	for _, z := range ctrl.Zones {
		m.AddZone(z)
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
