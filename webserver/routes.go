package webserver

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/alext/heating-controller/zone"
)

func (srv *WebServer) buildRouter() http.Handler {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").HandlerFunc(srv.zonesIndex)

	r.Methods("PUT").Path("/zones/{zone_id}/boost").HandlerFunc(srv.withZone(srv.zoneBoost))
	r.Methods("DELETE").Path("/zones/{zone_id}/boost").HandlerFunc(srv.withZone(srv.zoneCancelBoost))

	r.Methods("GET").Path("/zones/{zone_id}/schedule").HandlerFunc(srv.withZone(srv.scheduleEdit))
	r.Methods("POST").Path("/zones/{zone_id}/schedule").HandlerFunc(srv.withZone(srv.scheduleAddEvent))
	r.Methods("DELETE").Path("/zones/{zone_id}/schedule/{hour:\\d+}-{min:\\d+}").HandlerFunc(srv.withZone(srv.scheduleRemoveEvent))

	r.Methods("POST").Path("/zones/{zone_id}/thermostat/increment").HandlerFunc(srv.withZone(srv.thermostatInc))
	r.Methods("POST").Path("/zones/{zone_id}/thermostat/decrement").HandlerFunc(srv.withZone(srv.thermostatDec))

	r.Methods("PUT").Path("/zones/{zone_id}/activate").HandlerFunc(srv.withZone(srv.zoneActivate))
	r.Methods("PUT").Path("/zones/{zone_id}/deactivate").HandlerFunc(srv.withZone(srv.zoneDeactivate))

	return httpMethodOverrideHandler(r)
}

type zoneHandlerFunc func(http.ResponseWriter, *http.Request, *zone.Zone)

func (srv *WebServer) withZone(hf zoneHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if z, ok := srv.zones[mux.Vars(req)["zone_id"]]; ok {
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
