package metrics_test

import (
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/metrics"
)

var _ = Describe("Metrics handling", func() {
	var (
		m       *metrics.Metrics
		handler http.Handler
	)

	BeforeEach(func() {
		m = metrics.New(controller.New(nil))
		handler = m.Handler()
	})

	Describe("adding an info metric", func() {
		It("returns an info metric with the given version", func() {
			m.AddInfo("1.2.3")
			lines := getMetricsLines(handler)
			Expect(lines).To(ContainElement("# TYPE heating_controller_info gauge"))
			Expect(lines).To(ContainElement(`heating_controller_info{version="1.2.3"} 1`))
		})
	})
})
