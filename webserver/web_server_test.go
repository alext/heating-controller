package webserver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

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
	})
})

func doGetRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	req, _ := http.NewRequest("GET", "http://example.com"+path, nil)

	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return
}
