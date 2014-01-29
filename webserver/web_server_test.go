package webserver_test

import (
	"code.google.com/p/gomock/gomock"
	"encoding/json"
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
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(gomocktestreporter.New())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("returns an OK response", func() {
		server := webserver.New(8080)

		w := doGetRequest(server, "/")

		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("OK\n"))
	})

	Describe("Accessing outputs", func() {
		var (
			server *webserver.WebServer
		)

		BeforeEach(func() {
			server = webserver.New(8080)
		})

		Context("with no given outputs", func() {
			It("should return an empty list of outputs as json", func() {
				w := doGetRequest(server, "/outputs")

				Expect(w.Code).To(Equal(200))
				Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))
				Expect(w.Body.String()).To(Equal("{}"))
			})

			It("should return a 404 trying to access a non-existent output", func() {
				w := doGetRequest(server, "/outputs/foo")

				Expect(w.Code).To(Equal(404))
			})
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

			XIt("should return a list of outputs with their current state", func() {
				output1.EXPECT().Active().Return(true, nil)
				output2.EXPECT().Active().Return(false, nil)

				w := doGetRequest(server, "/outputs")

				Expect(w.Code).To(Equal(200))
				Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))
				Expect(w.Body.String()).To(Equal("Some json"))
			})

			It("should return details of an output when requested", func() {
				output1.EXPECT().Active().Return(true, nil)

				w := doGetRequest(server, "/outputs/one")

				Expect(w.Code).To(Equal(200))
				Expect(w.Header().Get("Content-Type")).To(Equal("application/json"))

				data, _ := decodeJsonOutput(w.Body.Bytes())
				Expect(data.Id).To(Equal("one"))
				Expect(data.Active).To(Equal(true))
			})

			It("should 404 for a non-existent output", func() {
				w := doGetRequest(server, "/outputs/foo")

				Expect(w.Code).To(Equal(404))
			})
		})
	})
})

func doGetRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	req, _ := http.NewRequest("GET", "http://example.com"+path, nil)

	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return
}

type jsonOutput struct {
	Id     string `json: id`
	Active bool   `json: active`
}

func decodeJsonOutput(data []byte) (jOut *jsonOutput, err error) {
	//var jOut jsonOutput
	err = json.Unmarshal(data, &jOut)
	if err != nil {
		return nil, err
	}
	return
}
