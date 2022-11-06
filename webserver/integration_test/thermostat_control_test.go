package integration_test

import (
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("controlling the thermostat", func() {
	var (
		page       *agouti.Page
		ctrl       *controller.Controller
		testServer *httptest.Server
	)

	BeforeEach(func() {
		ctrl = controller.New(nil)
		server := webserver.New(ctrl, 8080, "../templates", nil)
		testServer = httptest.NewServer(server)

		var err error
		page, err = agoutiDriver.NewPage()
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		page.Destroy()
		testServer.Close()
	})

	Describe("adjusting the thermostat", func() {
		var (
			zone1 *controller.Zone
		)

		BeforeEach(func() {
			sens := sensor.NewPushSensor("sens", "something")
			sens.Set(18253, time.Now())
			zone1 = controller.NewZone("one", output.Virtual("one"))
			zone1.SetupThermostat(sens, 19500)
			ctrl.AddZone(zone1)
		})

		It("increments the target temperature and redirects back", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			zoneContent := page.FindByID("zone-one")
			Expect(zoneContent).To(MatchText("Target temp\\s+19.5°C"))

			Expect(zoneContent.Find("input[value=\\+]").Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/"))
			zoneContent = page.FindByID("zone-one")
			Expect(zoneContent).To(MatchText("Target temp\\s+20°C"))
		})

		It("decrements the target temperature and redirects back", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			zoneContent := page.FindByID("zone-one")
			Expect(zoneContent).To(MatchText("Target temp\\s+19.5°C"))

			Expect(zoneContent.Find("input[value=\\-]").Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/"))
			zoneContent = page.FindByID("zone-one")
			Expect(zoneContent).To(MatchText("Target temp\\s+19°C"))
		})
	})
})
