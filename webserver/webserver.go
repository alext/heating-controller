package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/alext/heating-controller/controller"
)

type WebServer struct {
	controller    *controller.Controller
	listenUrl     string
	templatesPath string
	mux           http.Handler
}

func New(ctrl *controller.Controller, port int, templatesPath string) (srv *WebServer) {
	srv = &WebServer{
		controller:    ctrl,
		listenUrl:     fmt.Sprintf(":%d", port),
		templatesPath: templatesPath,
	}
	srv.mux = srv.buildRouter()
	return
}

func (srv *WebServer) Run() error {
	log.Print("[webserver] server starting on", srv.listenUrl)
	return http.ListenAndServe(srv.listenUrl, srv)
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	srv.mux.ServeHTTP(w, req)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	err := enc.Encode(data)
	if err != nil {
		log.Printf("[webserver] Error encoding JSON: %s", err.Error())
	}
}

func write404(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeError(w http.ResponseWriter, err error, optionalCode ...int) {
	code := http.StatusInternalServerError
	if len(optionalCode) > 0 {
		code = optionalCode[0]
	}
	log.Printf("[webserver] Error %d: %s", code, err.Error())
	http.Error(w, err.Error(), code)
}
