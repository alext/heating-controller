package webserver

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/alext/heating-controller/zone"
)

type WebServer struct {
	listenUrl     string
	templatesPath string
	mux           http.Handler
	zones         map[string]*zone.Zone
}

func New(port int, templatesPath string) (srv *WebServer) {
	srv = &WebServer{
		listenUrl:     fmt.Sprintf(":%d", port),
		templatesPath: templatesPath,
		zones:         make(map[string]*zone.Zone),
	}
	srv.mux = srv.buildRouter()
	return
}

func (srv *WebServer) AddZone(z *zone.Zone) {
	srv.zones[z.ID] = z
}

func (srv *WebServer) Run() error {
	log.Print("[webserver] server starting on", srv.listenUrl)
	return http.ListenAndServe(srv.listenUrl, srv)
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	srv.mux.ServeHTTP(w, req)
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(output)
}

func write404(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeError(w http.ResponseWriter, err error) {
	log.Printf("[webserver] Error : %s", err.Error())
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
