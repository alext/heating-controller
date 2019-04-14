package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"

	"github.com/alext/heating-controller/controller"
)

type Metrics struct {
	ctrl       *controller.Controller
	registry   *prometheus.Registry
	sensorDesc *prometheus.Desc
	zoneDesc   *prometheus.Desc
}

func newRegistry() *prometheus.Registry {
	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	r.MustRegister(prometheus.NewGoCollector())
	r.MustRegister(version.NewCollector("heating_controller"))
	return r
}

func New(ctrl *controller.Controller) *Metrics {
	m := &Metrics{
		ctrl:       ctrl,
		registry:   newRegistry(),
		sensorDesc: newDensorDesc(),
		zoneDesc:   newZoneDesc(),
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
