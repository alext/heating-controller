package metrics_test

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestSensor(t *testing.T) {
	RegisterFailHandler(Fail)

	log.SetOutput(ioutil.Discard)

	RunSpecs(t, "Metrics")
}

func timeMS(t time.Time) int64 {
	return t.Unix()*1000 + int64(t.Nanosecond()/1000000)
}

func getMetricsBody(h http.Handler) string {
	r := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	Expect(w.Code).To(Equal(200))
	body, err := ioutil.ReadAll(w.Body)
	Expect(err).NotTo(HaveOccurred())
	return string(body)
}

func getMetricsLines(h http.Handler) []string {
	return strings.Split(getMetricsBody(h), "\n")
}
