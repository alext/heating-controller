package integration_test

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("Editing the schedule for a zone", func() {
	var (
		page       *agouti.Page
		testServer *httptest.Server
		zone1      *controller.Zone
		zone2      *controller.Zone
	)

	BeforeEach(func() {
		ctrl := controller.New()

		zone1 = controller.NewZone("one", output.Virtual("one"))
		zone2 = controller.NewZone("two", output.Virtual("two"))
		ctrl.AddZone(zone1)
		ctrl.AddZone(zone2)

		server := webserver.New(ctrl, 8080, "../templates")
		testServer = httptest.NewServer(server)

		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		page.Destroy()
		testServer.Close()
	})

	Describe("Inspecting the current schedule", func() {
		It("should show an empty schedule with no events", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			link := page.FindByID("zone-one").FindByLink("edit schedule")
			Expect(link).To(BeFound())
			Expect(link.Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
			Expect(page.Find("h1")).To(HaveText("one schedule"))
		})

		Context("with some events", func() {
			BeforeEach(func() {
				zone1.AddEvent(controller.Event{Hour: 7, Min: 30, Action: controller.TurnOn})
				zone1.AddEvent(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff})
				zone1.AddEvent(controller.Event{Hour: 17, Min: 0, Action: controller.TurnOn})
				zone1.AddEvent(controller.Event{Hour: 21, Min: 45, Action: controller.TurnOff})
			})

			It("should show the schedule", func() {
				Expect(page.Navigate(testServer.URL)).To(Succeed())

				link := page.FindByID("zone-one").FindByLink("edit schedule")
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

			It("should allow adding an event", func() {
				Expect(page.Navigate(testServer.URL + "/zones/one/schedule")).To(Succeed())

				form := page.All("table tr").At(5).Find("form")
				Expect(form).To(BeFound())

				Expect(form.Find("input[name=hour]").Fill("14")).To(Succeed())
				Expect(form.Find("input[name=min]").Fill("42")).To(Succeed())
				Expect(form.Find("select[name=action]").Select("On")).To(Succeed())
				Expect(form.Find("input[value='Add Event']").Click()).To(Succeed())

				Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
				Expect(page.Find("h1")).To(HaveText("one schedule"))

				events := zone1.ReadEvents()
				Expect(events).To(HaveLen(5))
				Expect(events).To(ContainElement(controller.Event{Hour: 14, Min: 42, Action: controller.TurnOn}))
			})

			It("should allow removing an event", func() {
				Expect(page.Navigate(testServer.URL + "/zones/one/schedule")).To(Succeed())

				deleteButton := page.All("table tr").At(2).Find("input[value='Delete Event']")
				Expect(deleteButton).To(BeFound())
				Expect(deleteButton.Click()).To(Succeed())

				Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
				Expect(page.Find("h1")).To(HaveText("one schedule"))

				events := zone1.ReadEvents()
				Expect(events).To(HaveLen(3))
				Expect(events).NotTo(ContainElement(controller.Event{Hour: 8, Min: 30, Action: controller.TurnOff}))
			})
		})
	})
})
