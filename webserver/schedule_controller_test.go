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
	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("schedule controller", func() {
	var (
		server      *webserver.WebServer
		tempDataDir string
	)

	BeforeEach(func() {
		tempDataDir, _ = ioutil.TempDir("", "schedule_controller_test")
		controller.DataDir = tempDataDir
		server = webserver.New(8080, "")
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
			server.AddZone(zone1)
			zone1.Scheduler.AddEvent(scheduler.Event{Hour: 7, Min: 30, Action: scheduler.TurnOn})
			zone1.Scheduler.AddEvent(scheduler.Event{Hour: 8, Min: 30, Action: scheduler.TurnOff})

			values = url.Values{}
			values.Set("hour", "10")
			values.Set("min", "24")
			values.Set("action", "on")
		})

		It("should add the event to the schedule and redirect to the schedule", func() {

			w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			events := zone1.Scheduler.ReadEvents()
			Expect(events).To(HaveLen(3))
			Expect(events).To(ContainElement(scheduler.Event{Hour: 10, Min: 24, Action: scheduler.TurnOn}))
		})

		It("should save the zone state", func() {
			doRequestWithValues(server, "POST", "/zones/one/schedule", values)

			data := readFile(controller.DataDir + "/one.json")
			expected, _ := json.Marshal(map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 7, "min": 30, "action": "On"},
					{"hour": 8, "min": 30, "action": "Off"},
					{"hour": 10, "min": 24, "action": "On"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})

		Context("with invalid input", func() {
			It("should return an error with a non-numeric hour", func() {
				values.Set("hour", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("hour must be a number"))
				Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with a non-numeric minute", func() {
				values.Set("min", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("minute must be a number"))
				Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(2))
			})

			It("should return an error with a well-formed, but invalid event", func() {
				values.Set("min", "64")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("invalid event"))
				Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(2))
			})
		})
	})

	Describe("removing an event", func() {
		var (
			zone1 *controller.Zone
		)

		BeforeEach(func() {
			zone1 = controller.NewZone("one", output.Virtual("one"))
			server.AddZone(zone1)
			zone1.Scheduler.AddEvent(scheduler.Event{Hour: 7, Min: 30, Action: scheduler.TurnOn})
			zone1.Scheduler.AddEvent(scheduler.Event{Hour: 8, Min: 30, Action: scheduler.TurnOff})
		})

		It("should remove the matching event and redirect to the schedule", func() {
			w := doFakeDeleteRequest(server, "/zones/one/schedule/7-30")

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			events := zone1.Scheduler.ReadEvents()
			Expect(events).To(HaveLen(1))
			Expect(events).NotTo(ContainElement(scheduler.Event{Hour: 7, Min: 30, Action: scheduler.TurnOn}))
		})

		It("should save the zone state", func() {
			doFakeDeleteRequest(server, "/zones/one/schedule/7-30")

			data := readFile(controller.DataDir + "/one.json")
			expected, _ := json.Marshal(map[string]interface{}{
				"events": []map[string]interface{}{
					{"hour": 8, "min": 30, "action": "Off"},
				},
			})
			Expect(data).To(MatchJSON(expected))
		})

		It("should do nothing for a non-existent event", func() {
			w := doFakeDeleteRequest(server, "/zones/one/schedule/7-40")

			Expect(w.Code).To(Equal(302))
			Expect(w.Header().Get("Location")).To(Equal("/zones/one/schedule"))

			Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(2))
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
