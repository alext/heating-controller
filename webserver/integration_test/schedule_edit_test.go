package integration_test

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/thermostat/thermostatfakes"
	"github.com/alext/heating-controller/units"
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
		zone1.Thermostat = &thermostatfakes.FakeThermostat{}
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
				zone1.AddEvent(controller.Event{
					Time: units.NewTimeOfDay(7, 30), Action: controller.On,
					ThermAction: &controller.ThermostatAction{Action: controller.DecreaseTarget, Param: 19000},
				})
				zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(8, 30), Action: controller.Off})
				zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(17, 0), Action: controller.On})
				zone1.AddEvent(controller.Event{Time: units.NewTimeOfDay(21, 45), Action: controller.Off})
			})

			It("should show the schedule", func() {
				Expect(page.Navigate(testServer.URL)).To(Succeed())

				link := page.FindByID("zone-one").FindByLink("edit schedule")
				Expect(link).To(BeFound())
				Expect(link.Click()).To(Succeed())

				Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
				Expect(page.Find("h1")).To(HaveText("one schedule"))

				rows := page.All("table tr")

				Expect(rows.At(1).All("td").At(0)).To(HaveText("7:30 On DecreaseTarget(19°C)"))
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
				Expect(events).To(ContainElement(controller.Event{Time: units.NewTimeOfDay(14, 42), Action: controller.On}))
			})

			It("should allow adding an event with a thermostat action", func() {
				Expect(page.Navigate(testServer.URL + "/zones/one/schedule")).To(Succeed())

				form := page.All("table tr").At(5).Find("form")
				Expect(form).To(BeFound())

				Expect(form.Find("input[name=hour]").Fill("14")).To(Succeed())
				Expect(form.Find("input[name=min]").Fill("42")).To(Succeed())
				Expect(form.Find("select[name=action]").Select("On")).To(Succeed())
				Expect(form.Find("select[name=therm_action]").Select("Set Target")).To(Succeed())
				Expect(form.Find("input[name=therm_param]").Fill("19.5")).To(Succeed())
				Expect(form.Find("input[value='Add Event']").Click()).To(Succeed())

				Expect(page).To(HaveURL(testServer.URL + "/zones/one/schedule"))
				Expect(page.Find("h1")).To(HaveText("one schedule"))

				events := zone1.ReadEvents()
				Expect(events).To(HaveLen(5))
				Expect(events).To(ContainElement(controller.Event{
					Time: units.NewTimeOfDay(14, 42), Action: controller.On,
					ThermAction: &controller.ThermostatAction{Action: controller.SetTarget, Param: 19500},
				}))
			})

			It("does not show the thermostat action fields for a zone without a thermostat", func() {
				Expect(page.Navigate(testServer.URL + "/zones/two/schedule")).To(Succeed())

				form := page.All("table tr").At(1).Find("form")
				Expect(form).To(BeFound())

				Expect(form.Find("select[name=therm_action]")).ToNot(BeFound())
				Expect(form.Find("input[name=therm_param]")).ToNot(BeFound())
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
				Expect(events).NotTo(ContainElement(controller.Event{Time: units.NewTimeOfDay(8, 30), Action: controller.Off}))
			})
		})
	})
})
