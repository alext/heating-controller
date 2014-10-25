package webserver

import (
	"fmt"
	"net/http"

	"github.com/go-martini/martini"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
)

type WebServer struct {
	listenUrl string
	mux       http.Handler
	outputs   map[string]output.Output
}

func New(port int) (srv *WebServer) {
	srv = &WebServer{
		listenUrl: fmt.Sprintf(":%d", port),
		outputs:   make(map[string]output.Output),
	}
	srv.buildMux()
	return
}

func (srv *WebServer) buildMux() {
	m := martini.Classic()
	m.Handlers() // Clear default handlers
	m.Use(martini.Recovery())
	m.Use(martini.Static("public", martini.StaticOptions{SkipLogging: true}))

	srv.buildRoutes(m)

	srv.mux = m
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

func (srv *WebServer) findOutput(w http.ResponseWriter, c martini.Context, params martini.Params) {
	if out, ok := srv.outputs[params["id"]]; ok {
		c.Map(out)
	} else {
		write404(w)
	}
}

func write404(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
