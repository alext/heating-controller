package integration_test

import (
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/sclevine/agouti/core"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var _ = Describe("toggling a zone's state", func() {
	var (
		page       Page
		server     *webserver.WebServer
		testServer *httptest.Server
	)

	BeforeEach(func() {
		server = webserver.New(8080)
		testServer = httptest.NewServer(server)

		var err error
		page, err = agoutiDriver.Page()
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		page.Destroy()
		testServer.Close()
	})

	Describe("changing an output state", func() {
		var (
			output1 output.Output
			zone1   *zone.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = zone.New("one", output1)
			server.AddZone(zone1)
		})

		It("activates the output and redirects back to the index", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			button := page.All("table tr").At(1).Find("form input[value=Activate]")
			Expect(button).To(BeFound())

			Expect(button.Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/"))

			button = page.All("table tr").At(1).Find("form input[value=Deactivate]")
			Expect(button).To(BeFound())

			Expect(output1.Active()).To(Equal(true))
		})

		It("deactivates the output and redirects back to the index", func() {
			output1.Activate()

			Expect(page.Navigate(testServer.URL)).To(Succeed())

			button := page.All("table tr").At(1).Find("form input[value=Deactivate]")
			Expect(button).To(BeFound())

			Expect(button.Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/"))

			button = page.All("table tr").At(1).Find("form input[value=Activate]")
			Expect(button).To(BeFound())

			Expect(output1.Active()).To(Equal(false))
		})
	})
})
