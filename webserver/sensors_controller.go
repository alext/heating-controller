package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/alext/heating-controller/sensor"
	"github.com/alext/heating-controller/units"
	"github.com/gorilla/mux"
)

func (srv *WebServer) sensorIndex(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]*jsonSensor)
	for name, s := range srv.controller.SensorsByName {
		data[name] = newJSONSensor(s)
	}
	writeJSON(w, data)
}

func (srv *WebServer) sensorBulkPut(w http.ResponseWriter, req *http.Request) {
	var reqData struct {
		Temperatures map[string]units.Temperature `json:"temperatures"`
	}
	err := json.NewDecoder(req.Body).Decode(&reqData)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}

	for id, temp := range reqData.Temperatures {
		s, ok := srv.controller.SensorsByDeviceID[id]
		if !ok {
			log.Printf("[webserver] sensor bulk update ignoring unknown sensor '%s'", id)
			continue
		}
		ss, ok := s.(sensor.SettableSensor)
		if !ok {
			log.Printf("[webserver] sensor bulk update ignoring non-settable sensor '%s'", id)
			continue
		}
		ss.Set(temp, time.Now())
	}

	fmt.Fprintln(w, "OK")
}

func (srv *WebServer) sensorGet(w http.ResponseWriter, req *http.Request) {
	s, ok := srv.controller.SensorsByName[mux.Vars(req)["sensor_id"]]
	if !ok {
		write404(w)
		return
	}

	writeJSON(w, newJSONSensor(s))
}

func (srv *WebServer) sensorPut(w http.ResponseWriter, req *http.Request) {
	sensorID := mux.Vars(req)["sensor_id"]
	s, ok := srv.controller.SensorsByName[sensorID]
	if !ok {
		write404(w)
		return
	}
	ss, ok := s.(sensor.SettableSensor)
	if !ok {
		w.Header().Set("Allow", "GET")
		writeError(w, fmt.Errorf("Non-writable sensor %s", sensorID), http.StatusMethodNotAllowed)
		return
	}

	var data struct {
		Temp *units.Temperature `json:"temperature"`
	}
	err := json.NewDecoder(req.Body).Decode(&data)
	if err != nil {
		writeError(w, err, http.StatusBadRequest)
		return
	}
	if data.Temp == nil {
		writeError(w, fmt.Errorf("Missing temperature data in request"), http.StatusBadRequest)
		return
	}

	ss.Set(*data.Temp, time.Now())

	writeJSON(w, newJSONSensor(ss))
}

type jsonSensor struct {
	Temperature units.Temperature `json:"temperature"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

func newJSONSensor(s sensor.Sensor) *jsonSensor {
	temperature, updatedAt := s.Read()
	return &jsonSensor{
		Temperature: temperature,
		UpdatedAt:   updatedAt,
	}
}
