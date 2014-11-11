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

	api := r.PathPrefix("/api").Subrouter()
	api.Methods("GET").Path("/outputs").HandlerFunc(srv.apiOutputIndex)
	api.Methods("GET").Path("/outputs/{output_id}").HandlerFunc(srv.withOutput(srv.apiOutputShow))
	api.Methods("PUT").Path("/outputs/{output_id}/activate").HandlerFunc(srv.withOutput(srv.apiOutputActivate))
	api.Methods("PUT").Path("/outputs/{output_id}/deactivate").HandlerFunc(srv.withOutput(srv.apiOutputDeactivate))

	return r
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
