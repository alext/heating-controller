package webserver

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/alext/heating-controller/controller"
)

func (srv *WebServer) buildRouter() http.Handler {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").HandlerFunc(srv.zonesIndex)

	r.Methods("GET").Path("/sensors").HandlerFunc(srv.sensorIndex)
	r.Methods("PUT").Path("/sensors").HandlerFunc(srv.sensorBulkPut)
	r.Methods("GET").Path("/sensors/{sensor_id}").HandlerFunc(srv.sensorGet)
	r.Methods("PUT").Path("/sensors/{sensor_id}").HandlerFunc(srv.sensorPut)

	r.Methods("GET").Path("/zones").HandlerFunc(srv.zonesAPIIndex)

	r.Methods("PUT").Path("/zones/{zone_id}/boost").HandlerFunc(srv.withZone(srv.zoneBoost))
	r.Methods("DELETE").Path("/zones/{zone_id}/boost").HandlerFunc(srv.withZone(srv.zoneCancelBoost))

	r.Methods("GET").Path("/zones/{zone_id}/schedule").HandlerFunc(srv.withZone(srv.scheduleEdit))
	r.Methods("GET").Path("/zones/{zone_id}/schedule/new").HandlerFunc(srv.withZone(srv.scheduleNewEvent))
	r.Methods("POST").Path("/zones/{zone_id}/schedule").HandlerFunc(srv.withZone(srv.scheduleAddEvent))
	r.Methods("GET").Path("/zones/{zone_id}/schedule/{time:\\d+:\\d+}").HandlerFunc(srv.withZone(srv.scheduleEditEvent))
	r.Methods("PUT").Path("/zones/{zone_id}/schedule/{time:\\d+:\\d+}").HandlerFunc(srv.withZone(srv.scheduleUpdateEvent))
	r.Methods("DELETE").Path("/zones/{zone_id}/schedule/{time:\\d+:\\d+}").HandlerFunc(srv.withZone(srv.scheduleRemoveEvent))

	r.Methods("POST").Path("/zones/{zone_id}/thermostat/increment").HandlerFunc(srv.withZone(srv.thermostatInc))
	r.Methods("POST").Path("/zones/{zone_id}/thermostat/decrement").HandlerFunc(srv.withZone(srv.thermostatDec))

	r.Methods("GET").Path("/metrics").Handler(srv.controller.Metrics.Handler())

	return httpMethodOverrideHandler(r)
}

type zoneHandlerFunc func(http.ResponseWriter, *http.Request, *controller.Zone)

func (srv *WebServer) withZone(hf zoneHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if z, ok := srv.controller.Zones[mux.Vars(req)["zone_id"]]; ok {
			hf(w, req, z)
		} else {
			write404(w)
		}
	}
}

// Adapted from https://github.com/gorilla/handlers/blob/master/handlers.go#L343
func httpMethodOverrideHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			om := r.FormValue("_method")
			if om == "PUT" || om == "PATCH" || om == "DELETE" {
				r.Method = om
			}
		}
		h.ServeHTTP(w, r)
	})
}
