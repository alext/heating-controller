package webserver_test

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/units"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("schedule controller", func() {
	var (
		ctrl        *controller.Controller
		server      *webserver.WebServer
		tempDataDir string
	)

	BeforeEach(func() {
		tempDataDir, _ = ioutil.TempDir("", "schedule_controller_test")
		controller.DataDir = tempDataDir
		ctrl = controller.New()
		server = webserver.New(ctrl, 8080, "")
	})

	AfterEach(func() {
		os.RemoveAll(tempDataDir)
	})

	Describe("adding an event", func() {
		var (
			zone1  *controller.Zone
			values url.Values
		)

		BeforeEach(func() {
			zone1 = controller.NewZone("one", output.Virtual("one"))
			ctrl.AddZone(zone1)
			zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(7, 30), Action: controller.On})
			zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(8, 30), Action: controller.Off})

			values = url.Values{}
			values.Set("hour", "10")
			values.Set("min", "24")
			values.Set("action", "On")
			values.Set("therm_action", "")
			values.Set("therm_param", "")
		})

		It("should add the event to the schedule and redirect to the schedule", func() {

			w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			events := zone1.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events).To(ContainElement(controller.Event{Time: units.NewTimeOfDay(10, 24), Action: controller.On}))
		})

		It("should add the event with a thermostat action when requested", func() {
			values.Set("therm_action", "SetTarget")
			values.Set("therm_param", "19.5")
			w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			events := zone1.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events).To(ContainElement(controller.Event{
				Time: units.NewTimeOfDay(10, 24), Action: controller.On,
				ThermAction: &controller.ThermostatAction{Action: controller.SetTarget, Param: 19500},
			}))
		})

		It("should save the zone state", func() {
			doRequestWithValues(server, "POST", "/zones/one/schedule", values)

			data := readFile(controller.DataDir + "/one.json")
			expected, _ := json.Marshal(map[string]interface{}{
				"events": []map[string]interface{}{
					{"time": "7:30", "action": "On"},
					{"time": "8:30", "action": "Off"},
					{"time": "10:24", "action": "On"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})

		Context("with invalid input", func() {
			It("should return an error with an invalid action", func() {
				values.Set("action", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("invalid action"))
				Expect(zone1.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with a non-numeric hour", func() {
				values.Set("hour", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("hour must be a number"))
				Expect(zone1.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with a non-numeric minute", func() {
				values.Set("min", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("minute must be a number"))
				Expect(zone1.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with an invalid thermostat action", func() {
				values.Set("therm_action", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("invalid thermostat action"))
				Expect(zone1.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with an invalid thermostat param", func() {
				values.Set("therm_action", "IncreaseTarget")
				values.Set("therm_param", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("thermostat param must be a number"))
				Expect(zone1.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with a well-formed, but invalid event", func() {
				values.Set("hour", "25")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("invalid event"))
				Expect(zone1.ReadEvents()).To(HaveLen(2))
			})
		})
	})

	Describe("removing an event", func() {
		var (
			zone1 *controller.Zone
		)

		BeforeEach(func() {
			zone1 = controller.NewZone("one", output.Virtual("one"))
			ctrl.AddZone(zone1)
			zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(7, 30), Action: controller.On})
			zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(8, 30), Action: controller.Off})
		})

		It("should remove the matching event and redirect to the schedule", func() {
			w := doFakeDeleteRequest(server, "/zones/one/schedule/7:30")

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			events := zone1.ReadEvents()
			Expect(events).To(HaveLen(1))
			Expect(events).NotTo(ContainElement(controller.Event{Time: units.NewTimeOfDay(7, 30), Action: controller.On}))
		})

		It("should save the zone state", func() {
			doFakeDeleteRequest(server, "/zones/one/schedule/7:30")

			data := readFile(controller.DataDir + "/one.json")
			expected, _ := json.Marshal(map[string]interface{}{
				"events": []map[string]interface{}{
					{"time": "8:30", "action": "Off"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})

		It("should do nothing for a non-existent event", func() {
			w := doFakeDeleteRequest(server, "/zones/one/schedule/7:40")

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			Expect(zone1.ReadEvents()).To(HaveLen(2))
		})

		It("should 404 for non-numerical times in URL", func() {
			w := doFakeDeleteRequest(server, "/zones/one/schedule/foo-bar")
			Expect(w.Code).To(Equal(404))
		})
	})
})

func readFile(filename string) []byte {
	data, err := ioutil.ReadFile(filename)
	ExpectWithOffset(1, err).NotTo(HaveOccurred())
	return data
}
