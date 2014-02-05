package webserver_test

import (
	"code.google.com/p/gomock/gomock"
	"encoding/json"
	"errors"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/ginkgo/thirdparty/gomocktestreporter"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/output/mock_output"
	"github.com/alext/heating-controller/webserver"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Web Server Suite")
}

var _ = Describe("Web Server", func() {
	var (
		mockCtrl *gomock.Controller
		server *webserver.WebServer
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(gomocktestreporter.New())
		server = webserver.New(8080)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("returns an OK response", func() {
		w := doGetRequest(server, "/")

		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("OK\n"))
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

				var data map[string]jsonOutput
				json.Unmarshal(w.Body.Bytes(), &data)
				Expect(data["one"].Id).To(Equal("one"))
				Expect(data["one"].Active).To(Equal(true))
				Expect(data["two"].Id).To(Equal("two"))
				Expect(data["two"].Active).To(Equal(false))
			})

			It("should return a 500 and error string on error reading output state", func() {
				err := errors.New("Computer says no!")
				output1.EXPECT().Active().Return(false, err)

				w := doGetRequest(server, "/outputs")

				Expect(w.Code).To(Equal(500))
				Expect(w.Body.String()).To(Equal("Error reading output 'one': Computer says no!"))
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

			var data jsonOutput
			json.Unmarshal(w.Body.Bytes(), &data)
			Expect(data.Id).To(Equal("one"))
			Expect(data.Active).To(Equal(true))
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
			Expect(w.Body.String()).To(Equal("Error reading output 'one': Computer says no!"))
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

			var data jsonOutput
			json.Unmarshal(w.Body.Bytes(), &data)
			Expect(data.Id).To(Equal("one"))
			Expect(data.Active).To(Equal(true))
		})

		It("should deactivate the output and return the state", func() {
			gomock.InOrder(
				output1.EXPECT().Deactivate().Return(nil),
				output1.EXPECT().Active().Return(false, nil),
			)

			w := doPutRequest(server, "/outputs/one/deactivate")

			Expect(w.Code).To(Equal(200))

			var data jsonOutput
			json.Unmarshal(w.Body.Bytes(), &data)
			Expect(data.Id).To(Equal("one"))
			Expect(data.Active).To(Equal(false))
		})

		It("should return a 500 and error string if activating fails", func() {
			err := errors.New("Computer says no!")
			output1.EXPECT().Activate().Return(err)

			w := doPutRequest(server, "/outputs/one/activate")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(Equal("Error activating output 'one': Computer says no!"))
		})

		It("should return a 500 and error string if deactivating fails", func() {
			err := errors.New("Computer says no!")
			output1.EXPECT().Deactivate().Return(err)

			w := doPutRequest(server, "/outputs/one/deactivate")

			Expect(w.Code).To(Equal(500))
			Expect(w.Body.String()).To(Equal("Error deactivating output 'one': Computer says no!"))
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

func doGetRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	return doRequest(server, "GET", path)
}

func doPutRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	return doRequest(server, "PUT", path)
}

func doRequest(server http.Handler, method, path string) (w *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(method, "http://example.com"+path, nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return
}

type jsonOutput struct {
	Id     string `json: id`
	Active bool   `json: active`
}
