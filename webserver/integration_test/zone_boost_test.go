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

var _ = Describe("boosting a zone", func() {
	var (
		page       Page
		server     *webserver.WebServer
		testServer *httptest.Server
	)

	BeforeEach(func() {
		server = webserver.New(8080, "../templates")
		testServer = httptest.NewServer(server)

		var err error
		page, err = agoutiDriver.Page()
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		page.Destroy()
		testServer.Close()
	})

	Describe("using the boost function", func() {
		var (
			output1 output.Output
			zone1   *zone.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = zone.New("one", output1)
			server.AddZone(zone1)
			zone1.Scheduler.Start()
		})

		AfterEach(func() {
			zone1.Scheduler.Stop()
		})

		It("applies the boost and redirects back to the index", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			cell := boostCell(page, 2)

			Expect(cell.Find("select").Select("30 mins")).To(Succeed())
			cell.Find("input[value=Boost]").Click()

			Expect(page).To(HaveURL(testServer.URL + "/"))

			Expect(zone1.Active()).To(Equal(true))

			//cell = boostCell(page, 2)
			//Expect(cell).To(MatchText("Boosted"))
		})
	})

	Describe("cancelling a boost", func() {
	})
})

func boostCell(page Page, row int) Selection {
	cell := page.All("table tr").At(row - 1).All("td").At(3)
	ExpectWithOffset(1, cell).To(BeFound())
	return cell
}
