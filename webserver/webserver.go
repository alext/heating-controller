package webserver

import (
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

func write404(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeError(w http.ResponseWriter, err error) {
	log.Printf("[webserver] Error : %s", err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
