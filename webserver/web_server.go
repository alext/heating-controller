package webserver

import (
	"fmt"
	"net/http"
)

type WebServer struct {
	listenUrl string
	mux       *http.ServeMux
}

func New(port int) (srv *WebServer) {
	srv = &WebServer{
		listenUrl: fmt.Sprintf(":%d", port),
	}
	srv.buildMux()
	return
}

func (srv *WebServer) buildMux() {
	srv.mux = http.NewServeMux()
	srv.mux.HandleFunc("/", srv.rootHandler)
	srv.mux.HandleFunc("/outputs", srv.outputIndexHandler)
	srv.mux.HandleFunc("/outputs/", srv.outputHandler)
}

func (srv *WebServer) Run() error {
	logInfo("Web server starting on", srv.listenUrl)
	return http.ListenAndServe(srv.listenUrl, srv)
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	srv.mux.ServeHTTP(w, req)
}

func (srv *WebServer) rootHandler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK\n"))
}

func (srv *WebServer) outputIndexHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{}"))
}

func (srv *WebServer) outputHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusNotFound)
}
