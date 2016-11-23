package webserver_test

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestWebServer(t *testing.T) {
	RegisterFailHandler(Fail)
	log.SetOutput(ioutil.Discard)
	RunSpecs(t, "Web Server Suite")
}

func doGetRequest(server http.Handler, path string) *httptest.ResponseRecorder {
	return doRequest(server, "GET", path)
}

func doPutRequest(server http.Handler, path string) *httptest.ResponseRecorder {
	return doRequest(server, "PUT", path)
}

// POST request pretending to be a PUT request because browsers...
func doFakePutRequest(server http.Handler, path string) *httptest.ResponseRecorder {
	return doFakeRequestWithValues(server, "PUT", path, url.Values{})
}

func doFakeDeleteRequest(server http.Handler, path string) *httptest.ResponseRecorder {
	return doFakeRequestWithValues(server, "DELETE", path, url.Values{})
}

func doFakeRequestWithValues(server http.Handler, verb, path string, values url.Values) *httptest.ResponseRecorder {
	values.Set("_method", verb)
	return doRequestWithValues(server, "POST", path, values)
}

func doRequest(server http.Handler, method, path string) *httptest.ResponseRecorder {
	return doRequestWithValues(server, method, path)
}

func doRequestWithValues(server http.Handler, method, path string, values ...url.Values) *httptest.ResponseRecorder {
	var req *http.Request
	if len(values) > 0 {
		body := strings.NewReader(values[0].Encode())
		req, _ = http.NewRequest(method, "http://example.com"+path, body)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	} else {
		req, _ = http.NewRequest(method, "http://example.com"+path, nil)
	}
	w := httptest.NewRecorder()
	server.ServeHTTP(w, req)
	return w
}

func decodeJsonResponse(w *httptest.ResponseRecorder) map[string]interface{} {
	var data map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&data)
	ExpectWithOffset(1, err).To(BeNil())
	return data
}
