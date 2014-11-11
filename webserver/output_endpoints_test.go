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

var _ = Describe("Endpoints for an output", func() {
	var (
		mockCtrl *gomock.Controller
		server   *webserver.WebServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		server = webserver.New(8080, "")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("viewing details for an output", func() {
	})

	Describe("changing an output state", func() {
		var (
			output1 output.Output
		)

		BeforeEach(func() {
			output1 = output.Virtual("one")
			server.AddOutput(output1)
		})

		Describe("activating", func() {

			It("should activate the output", func() {
				doFakePutRequest(server, "/outputs/one/activate")

				Expect(output1.Active()).To(Equal(true))
			})

			It("should redirect to the index", func() {
				w := doFakePutRequest(server, "/outputs/one/activate")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})

			It("should show an error if activating fails", func() {
				mock_output := mock_output.NewMockOutput(mockCtrl)
				mock_output.EXPECT().Id().AnyTimes().Return("mock")
				server.AddOutput(mock_output)

				err := errors.New("Computer says no!")
				mock_output.EXPECT().Activate().Return(err)

				w := doFakePutRequest(server, "/outputs/mock/activate")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error activating output 'mock': Computer says no!\n"))
			})
		})

		Describe("deactivating", func() {
			BeforeEach(func() {
				output1.Activate()
			})

			It("should deactivate the output", func() {
				doFakePutRequest(server, "/outputs/one/deactivate")

				Expect(output1.Active()).To(Equal(false))
			})

			It("should redirect to the index", func() {
				w := doFakePutRequest(server, "/outputs/one/deactivate")

				Expect(w.Code).To(Equal(302))
				Expect(w.Header().Get("Location")).To(Equal("/"))
			})

			It("should show an error if activating fails", func() {
				mock_output := mock_output.NewMockOutput(mockCtrl)
				mock_output.EXPECT().Id().AnyTimes().Return("mock")
				server.AddOutput(mock_output)

				err := errors.New("Computer says no!")
				mock_output.EXPECT().Deactivate().Return(err)

				w := doFakePutRequest(server, "/outputs/mock/deactivate")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error deactivating output 'mock': Computer says no!\n"))
			})
		})
	})

})
