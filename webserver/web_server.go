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
		jOut, err := newJsonOutput(out)
		if err != nil {
			writeError(w, err)
			return
		}
		data[id] = jOut
	}
	writeJson(w, data)
}

func (srv *WebServer) outputHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		parts := strings.SplitN(req.URL.Path, "/", 3)
		if out, ok := srv.outputs[parts[2]]; ok {
			jOut, err := newJsonOutput(out)
			if err != nil {
				writeError(w, err)
				return
			}
			writeJson(w, jOut)
			return
		}
	}
	w.WriteHeader(http.StatusNotFound)
}

func writeJson(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func writeError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

type jsonOutput struct {
	Id     string `json: id`
	Active bool   `json: active`
}

func newJsonOutput(out output.Output) (jOut *jsonOutput, err error) {
	jOut = &jsonOutput{
		Id: out.Id(),
	}
	jOut.Active, err = out.Active()
	if err != nil {
		return nil, fmt.Errorf("Error reading output '%s': %s", jOut.Id, err.Error())
	}
	return
}
