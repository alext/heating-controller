package webserver

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/zone"
)

type WebServer struct {
	listenUrl string
	mux       http.Handler
	zones     map[string]*zone.Zone
	templates map[string]*template.Template
}

func New(port int) (srv *WebServer) {
	srv = &WebServer{
		listenUrl: fmt.Sprintf(":%d", port),
		zones:     make(map[string]*zone.Zone),
	}
	srv.mux = srv.buildRouter()
	srv.templates = parseTemplates()
	return
}

func (srv *WebServer) AddZone(z *zone.Zone) {
	srv.zones[z.ID] = z
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
