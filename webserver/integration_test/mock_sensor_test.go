package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"
)

type mockSensor struct {
	*httptest.Server
	temp int
}

func (s *mockSensor) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	data := map[string]interface{}{
		"temperature": s.temp,
		"updated_at":  time.Now(),
	}
	json.NewEncoder(w).Encode(&data)
}

func (s *mockSensor) Start() {
	s.Server = httptest.NewServer(s)
}
