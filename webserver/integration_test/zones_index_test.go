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

var _ = Describe("viewing the index", func() {
	var (
		page       *agouti.Page
		server     *webserver.WebServer
		testServer *httptest.Server
	)

	BeforeEach(func() {
		var err error

		server = webserver.New(8080, "../templates")
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
			zone1   *zone.Zone
			zone2   *zone.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			output2 = output.Virtual("two")
			zone1 = zone.New("one", output1)
			zone2 = zone.New("two", output2)
			server.AddZone(zone1)
			server.AddZone(zone2)
		})

		It("should return a list of zones with their current state", func() {
			output1.Activate()

			Expect(page.Navigate(testServer.URL)).To(Succeed())
			Expect(page.All("table tr").At(1).All("td").At(0)).To(HaveText("one"))
			Expect(page.All("table tr").At(1).All("td").At(1)).To(HaveText("active"))
			Expect(page.All("table tr").At(2).All("td").At(0)).To(HaveText("two"))
			Expect(page.All("table tr").At(2).All("td").At(1)).To(HaveText("inactive"))
		})
	})
})
