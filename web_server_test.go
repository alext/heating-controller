package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"net/http"
	"net/http/httptest"
	"testing"
	"."
)

func TestWebServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Web Server Suite")
}

var _ = Describe("Web Server", func() {

	It("returns an OK response", func() {
		server := main.NewWebServer(8080)
		req, _ := http.NewRequest("GET", "http://example.com/", nil)

		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)

		Expect(w.Code).To(Equal(200))
		Expect(w.Body.String()).To(Equal("OK\n"))
	})

})
