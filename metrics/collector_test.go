package metrics_test

import (
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/metrics"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/sensor"
)

var _ = Describe("The custom collector", func() {
	var (
		ctrl    *controller.Controller
		handler http.Handler
	)

	BeforeEach(func() {
		ctrl = controller.New()
		m := metrics.New(ctrl)
		r := prometheus.NewPedanticRegistry()
		err := r.Register(m)
		Expect(err).NotTo(HaveOccurred())
		handler = promhttp.HandlerFor(r, promhttp.HandlerOpts{})
	})

	Describe("exposing sensors", func() {
		It("returns no metrics with an empty controller", func() {
			body := getMetricsBody(handler)
			Expect(body).NotTo(ContainSubstring("house_temperature_celcius"))
		})

		It("exposes sensors along with their updated timestamp", func() {
			s1 := sensor.NewPushSensor("one", "1234")
			ctrl.AddSensor("one", s1)
			s2 := sensor.NewPushSensor("two", "2345")
			ctrl.AddSensor("two", s2)
			t1 := time.Now().Add(-40 * time.Second)
			t2 := time.Now().Add(-27 * time.Second)
			s1.Set(19000, t1)
			s2.Set(19435, t2)

			lines := getMetricsLines(handler)
			Expect(lines).To(ContainElement("# TYPE house_temperature_celcius gauge"))
			Expect(lines).To(ContainElement(fmt.Sprintf(`house_temperature_celcius{name="one"} 19 %d`, timeMS(t1))))
			Expect(lines).To(ContainElement(fmt.Sprintf(`house_temperature_celcius{name="two"} 19.435 %d`, timeMS(t2))))
		})

		It("ignores any sensors that haven't had any readings yet", func() {
			s1 := sensor.NewPushSensor("one", "1234")
			ctrl.AddSensor("one", s1)
			s2 := sensor.NewPushSensor("two", "2345")
			ctrl.AddSensor("two", s2)
			s1.Set(19000, time.Now())

			lines := getMetricsLines(handler)
			Expect(lines).To(ContainElement("# TYPE house_temperature_celcius gauge"))
			Expect(lines).NotTo(ContainElement(ContainSubstring(`name="two"`)))
		})
	})

	Describe("exposing zones", func() {
		AfterEach(func() {
			for _, z := range ctrl.Zones {
				z.Scheduler.Stop()
			}
		})

		It("returns no metrics with an empty controller", func() {
			body := getMetricsBody(handler)
			Expect(body).NotTo(ContainSubstring("house_heating_zone_active"))
		})

		It("exposes zones current state", func() {
			z1 := controller.NewZone("one", output.Virtual("one"))
			z1.Scheduler.Start()
			z2 := controller.NewZone("two", output.Virtual("two"))
			z2.Scheduler.Start()
			ctrl.AddZone(z1)
			ctrl.AddZone(z2)
			z1.Boost(time.Hour)

			lines := getMetricsLines(handler)
			Expect(lines).To(ContainElement("# TYPE house_heating_zone_active gauge"))
			Expect(lines).To(ContainElement(`house_heating_zone_active{name="one"} 1`))
			Expect(lines).To(ContainElement(`house_heating_zone_active{name="two"} 0`))
		})
	})
})
