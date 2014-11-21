package webserver

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/alext/heating-controller/output"
)

func (srv *WebServer) buildRouter() http.Handler {
	r := mux.NewRouter()
	r.Methods("GET").Path("/").HandlerFunc(srv.outputsIndex)
	r.Methods("PUT").Path("/outputs/{output_id}/activate").HandlerFunc(srv.withOutput(srv.outputActivate))
	r.Methods("PUT").Path("/outputs/{output_id}/deactivate").HandlerFunc(srv.withOutput(srv.outputDeactivate))

	return httpMethodOverrideHandler(r)
}

type outputHandlerFunc func(http.ResponseWriter, *http.Request, output.Output)

func (srv *WebServer) withOutput(hf outputHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if out, ok := srv.outputs[mux.Vars(req)["output_id"]]; ok {
			hf(w, req, out)
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
