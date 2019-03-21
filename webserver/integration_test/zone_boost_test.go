package integration_test

import (
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/metrics"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("boosting a zone", func() {
	var (
		page       *agouti.Page
		ctrl       *controller.Controller
		testServer *httptest.Server
	)

	BeforeEach(func() {
		ctrl = controller.New(metrics.New())
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

	Describe("using the boost function", func() {
		var (
			output1 output.Output
			zone1   *controller.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = controller.NewZone("one", output1)
			ctrl.AddZone(zone1)
			zone1.Scheduler.Start()
		})

		AfterEach(func() {
			zone1.Scheduler.Stop()
		})

		It("applies the boost and redirects back to the index", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			cell := boostCell(page, "one")

			Expect(cell.Find("select").Select("30 mins")).To(Succeed())
			cell.Find("input[value=Boost]").Click()

			Expect(page).To(HaveURL(testServer.URL + "/"))

			Expect(zone1.Active()).To(Equal(true))

			nextEvent := zone1.NextEvent()
			Expect(nextEvent).NotTo(BeNil())

			Expect(nextEvent.Action).To(Equal(controller.Off))

			eventTime := nextEvent.NextOccurance()
			expected := time.Now().Local().Add(30 * time.Minute)
			Expect(eventTime).To(BeTemporally("~", expected, 65*time.Second)) // allow for minute tickover.

			cell = boostCell(page, "one")
			Expect(cell).To(MatchText("Boosted"))
		})
	})

	Describe("cancelling a boost", func() {
		var (
			output1 output.Output
			zone1   *controller.Zone
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			zone1 = controller.NewZone("one", output1)
			ctrl.AddZone(zone1)
			zone1.Scheduler.Start()
			zone1.Boost(23 * time.Minute)
		})

		AfterEach(func() {
			zone1.Scheduler.Stop()
		})

		It("cancels the boost and redirects back to the index", func() {
			Expect(page.Navigate(testServer.URL)).To(Succeed())

			cell := boostCell(page, "one")
			Expect(cell).To(MatchText("Boosted"))
			button := cell.Find("input[value=\"Cancel boost\"]")
			Expect(button).To(BeFound())
			button.Click()

			Expect(page).To(HaveURL(testServer.URL + "/"))

			Expect(zone1.Active()).To(Equal(false))

			nextEvent := zone1.NextEvent()
			Expect(nextEvent).To(BeNil())

			cell = boostCell(page, "one")
			Expect(cell.Find("input[value=Boost]")).To(BeFound())
		})
	})
})

func boostCell(page *agouti.Page, zoneName string) *agouti.Selection {
	cell := page.FindByID("zone-" + zoneName).All("tr").At(2).All("td").At(1)
	ExpectWithOffset(1, cell).To(BeFound())
	return cell
}
