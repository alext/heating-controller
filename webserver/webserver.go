package webserver

import (
	"fmt"
	"net/http"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
)

type WebServer struct {
	listenUrl     string
	templatesPath string
	mux           http.Handler
	outputs       map[string]output.Output
}

func New(port int, templatesPath string) (srv *WebServer) {
	srv = &WebServer{
		listenUrl:     fmt.Sprintf(":%d", port),
		templatesPath: templatesPath,
		outputs:       make(map[string]output.Output),
	}
	srv.mux = srv.buildRouter()
	return
}

func (srv *WebServer) AddOutput(out output.Output) {
	srv.outputs[out.Id()] = out
}

func (srv *WebServer) Run() error {
	logger.Info("Web server starting on", srv.listenUrl)
	return http.ListenAndServe(srv.listenUrl, srv)
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	srv.mux.ServeHTTP(w, req)
}

func write404(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
