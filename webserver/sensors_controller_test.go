package webserver_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/units"
	"github.com/alext/heating-controller/webserver"
)

type dummySensor struct {
	temp       units.Temperature
	updateTime time.Time
}

func (s *dummySensor) Read() (units.Temperature, time.Time) {
	return s.temp, s.updateTime
}

func (s *dummySensor) Subscribe() <-chan units.Temperature { return nil }

var _ = Describe("sensors controller", func() {
	var (
		ctrl   *controller.Controller
		server *webserver.WebServer
	)

	BeforeEach(func() {
		ctrl = controller.New()
		server = webserver.New(ctrl, 8080, "")
	})

	Describe("sensors index", func() {
		var (
			s1 *dummySensor
			s2 *dummySensor
		)

		BeforeEach(func() {
			s1 = &dummySensor{}
			s1.temp, s1.updateTime = 18345, time.Now()
			s2 = &dummySensor{}
			s2.temp, s2.updateTime = 19542, time.Now()
			ctrl.AddSensor("one", s1)
			ctrl.AddSensor("two", s2)
		})

		It("returns details of all sensors", func() {
			resp := doGetRequest(server, "/sensors")
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))

			data := decodeJsonResponse(resp)
			Expect(data).To(HaveKey("one"))
			Expect(data).To(HaveKey("two"))
			data1 := data["one"].(map[string]interface{})
			Expect(data1["temperature"]).To(BeEquivalentTo(18345))
			data2 := data["two"].(map[string]interface{})
			Expect(data2["temperature"]).To(BeEquivalentTo(19542))
		})
	})
	Describe("reading a sensor", func() {
		var (
			sensor *dummySensor
		)

		BeforeEach(func() {
			sensor = &dummySensor{}
			ctrl.AddSensor("one", sensor)
		})

		It("returns the sensor data as JSON", func() {
			updateTime := time.Now().Add(-3 * time.Minute)
			sensor.temp, sensor.updateTime = 15643, updateTime

			resp := doGetRequest(server, "/sensors/one")
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Header().Get("Content-Type")).To(Equal("application/json"))

			data := decodeJsonResponse(resp)
			Expect(data["temperature"]).To(BeEquivalentTo(15643))

			updateTimeStr, _ := updateTime.MarshalText()
			Expect(data["updated_at"]).To(Equal(string(updateTimeStr)))
		})

		It("returns 404 for a non-existent sensor", func() {
			resp := doGetRequest(server, "/sensors/non-existent")
			Expect(resp.Code).To(Equal(http.StatusNotFound))
		})
	})

	Describe("setting a sensor", func() {
		var (
			s1 sensor.SettableSensor
		)

		BeforeEach(func() {
			s1 = sensor.NewPushSensor("something")
			s1.Set(12345, time.Now().Add(-1*time.Hour))
			ctrl.AddSensor("one", s1)
		})

		It("updates the sensor with the given details", func() {
			data := map[string]interface{}{
				"temperature": 15643,
			}
			resp := doJSONPutRequest(server, "/sensors/one", data)
			Expect(resp.Code).To(Equal(http.StatusOK))
			temp, updated := s1.Read()
			Expect(temp).To(BeEquivalentTo(15643))
			Expect(updated).To(BeTemporally("~", time.Now(), 100*time.Millisecond))

			respData := decodeJsonResponse(resp)
			Expect(respData["temperature"]).To(BeEquivalentTo(15643))
		})

		It("returns a 400 for invalid data", func() {
			data := map[string]interface{}{
				"foo": "bar",
			}
			resp := doJSONPutRequest(server, "/sensors/one", data)
			Expect(resp.Code).To(Equal(http.StatusBadRequest))
			temp, _ := s1.Read()
			Expect(temp).To(BeEquivalentTo(12345))
		})

		It("returns 405 for a non-writable sensor", func() {
			s2 := &dummySensor{}
			s2.temp, s2.updateTime = 12345, time.Now().Add(-1*time.Hour)
			ctrl.AddSensor("two", s2)

			data := map[string]interface{}{
				"temperature": 15643,
			}
			resp := doJSONPutRequest(server, "/sensors/two", data)
			Expect(resp.Code).To(Equal(405))
			temp, _ := s1.Read()
			Expect(temp).To(BeEquivalentTo(12345))
		})
	})

})
