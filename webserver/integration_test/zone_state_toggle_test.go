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

var _ = Describe("toggling a zone's state", func() {
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

	Describe("changing an output state", func() {
		var (
			output1 output.Output
			zone1   *controller.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = controller.NewZone("one", output1)
			server.AddZone(zone1)
		})

		It("activates the output and redirects back to the index", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			button := page.Find("#zone-one form input[value=Activate]")
			Expect(button).To(BeFound())

			Expect(button.Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/"))

			button = page.Find("#zone-one form input[value=Deactivate]")
			Expect(button).To(BeFound())

			Expect(output1.Active()).To(Equal(true))
		})

		It("deactivates the output and redirects back to the index", func() {
			output1.Activate()

			Expect(page.Navigate(testServer.URL)).To(Succeed())

			button := page.Find("#zone-one form input[value=Deactivate]")
			Expect(button).To(BeFound())

			Expect(button.Click()).To(Succeed())

			Expect(page).To(HaveURL(testServer.URL + "/"))

			button = page.Find("#zone-one form input[value=Activate]")
			Expect(button).To(BeFound())

			Expect(output1.Active()).To(Equal(false))
		})
	})
})
