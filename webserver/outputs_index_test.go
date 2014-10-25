package webserver_test

import (
	"errors"

	"code.google.com/p/gomock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/output/mock_output"
	"github.com/alext/heating-controller/webserver"
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

	Context("with no outputs", func() {
		It("should show a message indicating there are no outputs", func() {
			response := doGetRequest(server, "/")

			Expect(response.Code).To(Equal(200))
			Expect(response.Body.String()).To(ContainSubstring("No outputs"))
		})
	})

	Context("with some outputs", func() {
		var (
			output1 output.Output
			output2 output.Output
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			output2 = output.Virtual("two")
			server.AddOutput(output1)
			server.AddOutput(output2)
		})

		It("should return a list of outputs with their current state", func() {
			output1.Activate()

			response := doGetRequest(server, "/")

			Expect(response.Code).To(Equal(200))

			body := response.Body.String()
			Expect(body).To(MatchRegexp(`one</td>\s*<td>active`))
			Expect(body).To(MatchRegexp(`two</td>\s*<td>inactive`))

			Expect(response.Body.String()).NotTo(ContainSubstring("No outputs"))
		})

		It("should return a 500 and error string on error reading output state", func() {
			mock_output := mock_output.NewMockOutput(mockCtrl)
			mock_output.EXPECT().Id().AnyTimes().Return("mock")
			server.AddOutput(mock_output)

			err := errors.New("Computer says no!")
			mock_output.EXPECT().Active().Return(false, err)

			w := doGetRequest(server, "/")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(ContainSubstring("Computer says no!"))
		})
	})
})
