package webserver_test

import (
	"errors"

	"code.google.com/p/gomock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/output/mock_output"
	"github.com/alext/heating-controller/webserver"
	"github.com/alext/heating-controller/zone"
)

var _ = Describe("The index page", func() {
	var (
		mockCtrl *gomock.Controller
		server   *webserver.WebServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		server = webserver.New(8080, "./templates")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("with no zones", func() {
		It("should show a message indicating there are no zones", func() {
			response := doGetRequest(server, "/")

			Expect(response.Code).To(Equal(200))
			Expect(response.Body.String()).To(ContainSubstring("No zones"))
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

			response := doGetRequest(server, "/")

			Expect(response.Code).To(Equal(200))

			body := response.Body.String()
			Expect(body).To(MatchRegexp(`one</td>\s*<td>active`))
			Expect(body).To(MatchRegexp(`two</td>\s*<td>inactive`))

			Expect(response.Body.String()).NotTo(ContainSubstring("No zones"))
		})

		It("should return a 500 and error string on error reading output state", func() {
			mockOutput := mock_output.NewMockOutput(mockCtrl)
			server.AddZone(zone.New("mock", mockOutput))

			err := errors.New("Computer says no!")
			mockOutput.EXPECT().Active().Return(false, err)

			w := doGetRequest(server, "/")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(ContainSubstring("Computer says no!"))
		})
	})
})
