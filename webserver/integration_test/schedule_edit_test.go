package integration_test

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/core"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var _ = Describe("Editing the schedule for a zone", func() {
	var (
		page       Page
		server     *webserver.WebServer
		testServer *httptest.Server
		zone1      *zone.Zone
		zone2      *zone.Zone
	)

	BeforeEach(func() {
		server = webserver.New(8080, "../templates")

		zone1 = zone.New("one", output.Virtual("one"))
		zone2 = zone.New("two", output.Virtual("two"))
		server.AddZone(zone1)
		server.AddZone(zone2)

		testServer = httptest.NewServer(server)

		var err error
		page, err = agoutiDriver.Page()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		page.Destroy()
		testServer.Close()
	})

	Describe("Inspecting the current schedule", func() {
		It("should show an empty schedule with no events", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			link := page.All("table tr").At(1).FindByLink("edit schedule")
			Expect(link).To(BeFound())
			Expect(link.Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
			Expect(page.Find("h1")).To(HaveText("one schedule"))
		})

		Context("with some events", func() {
			BeforeEach(func() {
				zone1.Scheduler.AddEvent(scheduler.Event{Hour: 7, Min: 30, Action: scheduler.TurnOn})
				zone1.Scheduler.AddEvent(scheduler.Event{Hour: 8, Min: 30, Action: scheduler.TurnOff})
				zone1.Scheduler.AddEvent(scheduler.Event{Hour: 17, Min: 0, Action: scheduler.TurnOn})
				zone1.Scheduler.AddEvent(scheduler.Event{Hour: 21, Min: 45, Action: scheduler.TurnOff})
			})

			It("should show the schedule", func() {
				Expect(page.Navigate(testServer.URL)).To(Succeed())

				link := page.All("table tr").At(1).FindByLink("edit schedule")
				Expect(link).To(BeFound())
				Expect(link.Click()).To(Succeed())

				Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
				Expect(page.Find("h1")).To(HaveText("one schedule"))

				rows := page.All("table tr")

				Expect(rows.At(1).All("td").At(0)).To(HaveText("7:30 On"))
				Expect(rows.At(2).All("td").At(0)).To(HaveText("8:30 Off"))
				Expect(rows.At(3).All("td").At(0)).To(HaveText("17:00 On"))
				Expect(rows.At(4).All("td").At(0)).To(HaveText("21:45 Off"))
			})
		})
	})
})
