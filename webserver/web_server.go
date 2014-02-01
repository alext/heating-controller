package webserver

import (
	"encoding/json"
	"fmt"
	"github.com/alext/heating-controller/output"
	"net/http"
	"strings"
)

type WebServer struct {
	listenUrl string
	mux       *http.ServeMux
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
	srv.mux = http.NewServeMux()
	srv.mux.HandleFunc("/", srv.rootHandler)
	srv.mux.HandleFunc("/outputs", srv.outputIndexHandler)
	srv.mux.HandleFunc("/outputs/", srv.outputHandler)
}

func (srv *WebServer) AddOutput(out output.Output) {
	srv.outputs[out.Id()] = out
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
	data := make(map[string]*jsonOutput, len(srv.outputs))
	for id, out := range srv.outputs {
		data[id], _ = newJsonOutput(out)
	}
	jsonData, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (srv *WebServer) outputHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		parts := strings.SplitN(req.URL.Path, "/", 3)
		if out, ok := srv.outputs[parts[2]]; ok {
			jOut, _ := newJsonOutput(out)
			jsonData, _ := json.Marshal(jOut)
			w.Header().Set("Content-Type", "application/json")
			w.Write(jsonData)
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

type jsonOutput struct {
	Id     string `json: id`
	Active bool   `json: active`
}

func newJsonOutput(out output.Output) (*jsonOutput, error) {
	jOut := &jsonOutput{
		Id: out.Id(),
	}
	jOut.Active, _ = out.Active()
	return jOut, nil
}
