package webserver_test

import (
	"errors"

	"code.google.com/p/gomock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output/mock_output"
	"github.com/alext/heating-controller/webserver"
)

var _ = Describe("Output API", func() {
	var (
		mockCtrl *gomock.Controller
		server   *webserver.WebServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		server = webserver.New(8080)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("outputs index", func() {
		It("should return an empty list of outputs as json", func() {
			w := doGetRequest(server, "/outputs")

			Expect(w.Code).To(Equal(200))
			Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))
			Expect(w.Body.String()).To(Equal("{}"))
		})

		Context("with some outputs", func() {
			var (
				output1 *mock_output.MockOutput
				output2 *mock_output.MockOutput
			)

			BeforeEach(func() {
				output1 = mock_output.NewMockOutput(mockCtrl)
				output1.EXPECT().Id().AnyTimes().Return("one")
				output2 = mock_output.NewMockOutput(mockCtrl)
				output2.EXPECT().Id().AnyTimes().Return("two")
				server.AddOutput(output1)
				server.AddOutput(output2)
			})

			It("should return a list of outputs with their current state", func() {
				output1.EXPECT().Active().Return(true, nil)
				output2.EXPECT().Active().Return(false, nil)

				w := doGetRequest(server, "/outputs")

				Expect(w.Code).To(Equal(200))
				Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))

				data := decodeJsonResponse(w)
				data1, ok := data["one"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(data1["id"]).To(Equal("one"))
				Expect(data1["active"]).To(Equal(true))
				data2, ok := data["two"].(map[string]interface{})
				Expect(ok).To(BeTrue())
				Expect(data2["id"]).To(Equal("two"))
				Expect(data2["active"]).To(Equal(false))
			})

			It("should return a 500 and error string on error reading output state", func() {
				err := errors.New("Computer says no!")
				output1.EXPECT().Active().Return(false, err)

				w := doGetRequest(server, "/outputs")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error reading output 'one': Computer says no!\n"))
			})
		})
	})

	Describe("reading an output", func() {
		var (
			output1 *mock_output.MockOutput
			output2 *mock_output.MockOutput
		)

		BeforeEach(func() {
			output1 = mock_output.NewMockOutput(mockCtrl)
			output1.EXPECT().Id().AnyTimes().Return("one")
			output2 = mock_output.NewMockOutput(mockCtrl)
			output2.EXPECT().Id().AnyTimes().Return("two")
			server.AddOutput(output1)
			server.AddOutput(output2)
		})

		It("should return details of an output when requested", func() {
			output1.EXPECT().Active().Return(true, nil)

			w := doGetRequest(server, "/outputs/one")

			Expect(w.Code).To(Equal(200))
			Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))

			data := decodeJsonResponse(w)
			Expect(data["id"]).To(Equal("one"))
			Expect(data["active"]).To(Equal(true))
		})

		It("should 404 for a non-existent output", func() {
			w := doGetRequest(server, "/outputs/foo")

			Expect(w.Code).To(Equal(404))
		})

		It("should return a 500 and error string on error reading output state", func() {
			err := errors.New("Computer says no!")
			output1.EXPECT().Active().Return(false, err)

			w := doGetRequest(server, "/outputs/one")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(Equal("Error reading output 'one': Computer says no!\n"))
		})

		It("should 404 trying to get a subpath of an output", func() {
			w := doGetRequest(server, "/outputs/one/foo")
			Expect(w.Code).To(Equal(404))
		})

		It("should 405 trying to PUT to an output", func() {
			w := doPutRequest(server, "/outputs/one")
			Expect(w.Code).To(Equal(405))
			Expect(w.Header().Get("Allow")).To(Equal("GET"))
		})
	})

	Describe("changing an output state", func() {
		var (
			output1 *mock_output.MockOutput
		)

		BeforeEach(func() {
			output1 = mock_output.NewMockOutput(mockCtrl)
			output1.EXPECT().Id().AnyTimes().Return("one")
			server.AddOutput(output1)
		})

		It("should activate the output and return the state", func() {
			gomock.InOrder(
				output1.EXPECT().Activate().Return(nil),
				output1.EXPECT().Active().Return(true, nil),
			)

			w := doPutRequest(server, "/outputs/one/activate")

			Expect(w.Code).To(Equal(200))

			data := decodeJsonResponse(w)
			Expect(data["id"]).To(Equal("one"))
			Expect(data["active"]).To(Equal(true))
		})

		It("should deactivate the output and return the state", func() {
			gomock.InOrder(
				output1.EXPECT().Deactivate().Return(nil),
				output1.EXPECT().Active().Return(false, nil),
			)

			w := doPutRequest(server, "/outputs/one/deactivate")

			Expect(w.Code).To(Equal(200))

			data := decodeJsonResponse(w)
			Expect(data["id"]).To(Equal("one"))
			Expect(data["active"]).To(Equal(false))
		})

		It("should return a 500 and error string if activating fails", func() {
			err := errors.New("Computer says no!")
			output1.EXPECT().Activate().Return(err)

			w := doPutRequest(server, "/outputs/one/activate")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(Equal("Error activating output 'one': Computer says no!\n"))
		})

		It("should return a 500 and error string if deactivating fails", func() {
			err := errors.New("Computer says no!")
			output1.EXPECT().Deactivate().Return(err)

			w := doPutRequest(server, "/outputs/one/deactivate")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(Equal("Error deactivating output 'one': Computer says no!\n"))
		})

		It("should 404 for a non-existent subpath of output", func() {
			w := doPutRequest(server, "/outputs/one/foo")
			Expect(w.Code).To(Equal(404))
		})

		It("should 405 for a get request", func() {
			w := doGetRequest(server, "/outputs/one/activate")
			Expect(w.Code).To(Equal(405))
			Expect(w.Header().Get("Allow")).To(Equal("PUT"))

			w = doGetRequest(server, "/outputs/one/deactivate")
			Expect(w.Code).To(Equal(405))
			Expect(w.Header().Get("Allow")).To(Equal("PUT"))
		})
	})
})
