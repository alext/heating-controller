package webserver_test

import (
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/metrics"
	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("serving metrics", func() {
	var (
		ctrl   *controller.Controller
		server *webserver.WebServer
	)

	BeforeEach(func() {
		ctrl = controller.New()
		m := metrics.New(ctrl)
		server = webserver.New(ctrl, 8080, "", m.Handler())
	})

	It("serves the metrics", func() {
		sens := sensor.NewPushSensor("foo", "1234")
		now := time.Now()
		sens.Set(19500, now)
		ctrl.AddSensor("foo", sens)

		resp := doGetRequest(server, "/metrics")
		Expect(resp.Code).To(Equal(http.StatusOK))
		Expect(resp.Header().Get("Content-Type")).To(Equal("text/plain; version=0.0.4; charset=utf-8"))

		body, err := ioutil.ReadAll(resp.Body)
		Expect(err).NotTo(HaveOccurred())

		Expect(string(body)).To(ContainSubstring("house_temperature_celcius{name=\"foo\"} 19.5 %d", now.Unix()))
	})
})
