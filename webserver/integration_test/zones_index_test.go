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

var _ = Describe("viewing the index", func() {
	var (
		page       *agouti.Page
		ctrl       *controller.Controller
		testServer *httptest.Server
	)

	BeforeEach(func() {
		var err error

		ctrl = controller.New()
		server := webserver.New(ctrl, 8080, "../templates")
		testServer = httptest.NewServer(server)

		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		page.Destroy()
		testServer.Close()
	})

	Context("with no zones", func() {
		It("should show a message indicating there are no zones", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())
			Expect(page.Find("body p")).To(HaveText("No zones"))
		})
	})

	Context("with some zones", func() {
		var (
			output1 output.Output
			output2 output.Output
			zone1   *controller.Zone
			zone2   *controller.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			output2 = output.Virtual("two")
			zone1 = controller.NewZone("one", output1)
			zone2 = controller.NewZone("two", output2)
			ctrl.AddZone(zone1)
			ctrl.AddZone(zone2)
		})

		It("should return a list of zones with their current state", func() {
			output1.Activate()

			Expect(page.Navigate(testServer.URL)).To(Succeed())
			zoneContent := page.FindByID("zone-one")
			Expect(zoneContent).To(BeFound())
			Expect(zoneContent.Find("th")).To(HaveText("one"))
			Expect(zoneContent.All("tr").At(0).Find("td")).To(HaveText("active"))

			zoneContent = page.FindByID("zone-two")
			Expect(zoneContent).To(BeFound())
			Expect(zoneContent.Find("th")).To(HaveText("two"))
			Expect(zoneContent.All("tr").At(0).Find("td")).To(HaveText("inactive"))
		})

		Context("with a thermostat configured", func() {
			var (
				sensor *mockSensor
			)

			BeforeEach(func() {
				sensor = &mockSensor{temp: 18253}
				sensor.Start()
				zone1.SetupThermostat(sensor.URL, 19500)
			})
			AfterEach(func() {
				if sensor != nil {
					sensor.Close()
				}
			})

			It("should include details from the thermostat", func() {
				Expect(page.Navigate(testServer.URL)).To(Succeed())
				zoneContent := page.FindByID("zone-one")
				Expect(zoneContent).To(BeFound())
				Expect(zoneContent).To(MatchText("Current temp\\s+18.253"))
				Expect(zoneContent).To(MatchText("Target temp\\s+19.5"))
			})
		})
	})
})
