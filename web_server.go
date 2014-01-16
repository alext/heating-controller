package main

import (
	"fmt"
	"net/http"
)

type WebServer struct {
	listenUrl string
}

func NewWebServer(port int) *WebServer {
	return &WebServer{
		listenUrl: fmt.Sprintf(":%d", port),
	}
}

func (srv *WebServer) Run() error {
	logInfo("Web server starting on", srv.listenUrl)
	return http.ListenAndServe(srv.listenUrl, srv)
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("OK\n"))
}
