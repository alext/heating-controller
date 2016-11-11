package integration_test

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var _ = Describe("controlling the thermostat", func() {
	var (
		page       *agouti.Page
		server     *webserver.WebServer
		testServer *httptest.Server
	)

	BeforeEach(func() {
		server = webserver.New(8080, "../templates")
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
			sensor *mockSensor
			zone1  *zone.Zone
		)

		BeforeEach(func() {
			sensor = &mockSensor{temp: 18253}
			sensor.Start()
			zone1 = zone.New("one", output.Virtual("one"))
			zone1.SetupThermostat(sensor.URL, 19500)
			server.AddZone(zone1)
		})

		AfterEach(func() {
			if sensor != nil {
				sensor.Close()
			}
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
