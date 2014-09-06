package webserver

import (
	"encoding/json"
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
	m.Handlers(martini.Recovery())

	m.Get("/", func() string {
		return "OK\n"
	})
	m.Get("/outputs", srv.outputIndex)
	m.Get("/outputs/:id", srv.outputShow)
	m.Put("/outputs/:id/activate", srv.outputActivate)
	m.Put("/outputs/:id/deactivate", srv.outputDeactivate)

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

func (srv *WebServer) outputIndex(w http.ResponseWriter) {
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

func (srv *WebServer) outputShow(w http.ResponseWriter, params martini.Params) {
	if out, ok := srv.outputs[params["id"]]; ok {
		writeOutputJson(w, out)
	} else {
		write404(w)
	}
}

func (srv *WebServer) outputActivate(w http.ResponseWriter, params martini.Params) {
	if out, ok := srv.outputs[params["id"]]; ok {
		err := out.Activate()
		if err != nil {
			writeError(w, fmt.Errorf("Error activating output '%s': %s", out.Id(), err.Error()))
			return
		}
		writeOutputJson(w, out)
	} else {
		write404(w)
	}
}

func (srv *WebServer) outputDeactivate(w http.ResponseWriter, params martini.Params) {
	if out, ok := srv.outputs[params["id"]]; ok {
		err := out.Deactivate()
		if err != nil {
			writeError(w, fmt.Errorf("Error deactivating output '%s': %s", out.Id(), err.Error()))
			return
		}
		writeOutputJson(w, out)
	} else {
		write404(w)
	}
}

func writeOutputJson(w http.ResponseWriter, out output.Output) {
	jOut, err := newJsonOutput(out)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJson(w, jOut)
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

func write404(w http.ResponseWriter) {
	http.Error(w, "Not found", http.StatusNotFound)
}

func writeError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type jsonOutput struct {
	Id     string `json:"id"`
	Active bool   `json:"active"`
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
