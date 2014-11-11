package webserver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/alext/heating-controller/logger"
)

func TestWebServer(t *testing.T) {
	RegisterFailHandler(Fail)
	logger.SetDestination("/dev/null")
	RunSpecs(t, "Web Server Suite")
}

func doGetRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	return doRequest(server, "GET", path)
}

func doPutRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	return doRequest(server, "PUT", path)
}

// POST request pretending to be a PUT request because browsers...
func doFakePutRequest(server http.Handler, path string) (w *httptest.ResponseRecorder) {
	//return doRequest(server, "POST", path, strings.NewReader(url.Values{"_method": {"PUT"}}.Encode()))
	body := strings.NewReader(url.Values{"_method": {"PUT"}}.Encode())
	req, _ := http.NewRequest("POST", "http://example.com"+path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return
}

func doRequest(server http.Handler, method, path string) (w *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(method, "http://example.com"+path, nil)
	w = httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return
}

func decodeJsonResponse(w *httptest.ResponseRecorder) (map[string]interface{}) {
	var data map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &data)
	Expect(err).To(BeNil())
	return data
}
