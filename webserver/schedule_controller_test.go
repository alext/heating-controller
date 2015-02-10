package webserver_test

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var _ = Describe("schedule controller", func() {
	var (
		server *webserver.WebServer
	)

	BeforeEach(func() {
		server = webserver.New(8080)
	})

	Describe("adding an event", func() {
		var (
			zone1 *zone.Zone
		)

		BeforeEach(func() {
			zone1 = zone.New("one", output.Virtual("one"))
			server.AddZone(zone1)
		})

		Context("with invalid input", func() {
			var values url.Values

			BeforeEach(func() {
				values = url.Values{}
				values.Set("hour", "10")
				values.Set("min", "24")
				values.Set("action", "on")
			})

			It("should return an error with a non-numeric hour", func() {
				values.Set("hour", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("hour must be a number"))
				Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(0))
			})

			It("should return an error with a non-numeric minute", func() {
				values.Set("min", "fooey")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("minute must be a number"))
				Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(0))
			})

			It("should return an error with a well-formed, but invalid event", func() {
				values.Set("min", "64")
				w := doRequestWithValues(server, "POST", "/zones/one/schedule", values)
				Expect(w.Code).To(Equal(400))
				Expect(w.Body.String()).To(ContainSubstring("invalid event"))
				Expect(zone1.Scheduler.ReadEvents()).To(HaveLen(0))
			})
		})
	})

	Describe("removing an event", func() {
		var (
			zone1 *zone.Zone
		)

		BeforeEach(func() {
			zone1 = zone.New("one", output.Virtual("one"))
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
