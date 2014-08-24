package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-martini/martini"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
	"github.com/alext/heating-controller/timer"
)

type WebServer struct {
	listenUrl string
	mux       http.Handler
	ctrl	  controller.Controller
}

func New(ctrl controller.Controller, port int) (srv *WebServer) {
	srv = &WebServer{
		listenUrl: fmt.Sprintf(":%d", port),
		ctrl:      ctrl,
	}
	srv.buildMux()
	return
}

func (srv *WebServer) buildMux() {
	m := martini.Classic()
	m.Handlers() // Clear default handlers
	m.Use(martini.Recovery())
	m.Use(martini.Static("public", martini.StaticOptions{SkipLogging: true}))

	m.Get("/", func() string {
		return "OK\n"
	})
	m.Get("/outputs", srv.outputIndex)
	m.Group("/outputs/:id", func(r martini.Router) {
		r.Get("", srv.outputShow)
		r.Put("/activate", srv.outputActivate)
		r.Put("/deactivate", srv.outputDeactivate)
	}, srv.findOutput)
	m.Get("/timers", srv.timerIndex)
	m.Group("/timers.:id", func(r martini.Router) {
		r.Get("", srv.timerShow)
	}, srv.findTimer)

	srv.mux = m
}

func (srv *WebServer) Run() error {
	logger.Info("Web server starting on", srv.listenUrl)
	return http.ListenAndServe(srv.listenUrl, srv)
}

func (srv *WebServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	srv.mux.ServeHTTP(w, req)
}

func (srv *WebServer) findOutput(w http.ResponseWriter, c martini.Context, params martini.Params) {
	out := srv.ctrl.Output(params["id"])
	if out != nil {
		c.Map(out)
	} else {
		write404(w)
	}
}

func (srv *WebServer) outputIndex(w http.ResponseWriter) {
	outputs := srv.ctrl.Outputs()
	data := make(map[string]*jsonOutput, len(outputs))
	for id, out := range outputs {
		jOut, err := newJsonOutput(out)
		if err != nil {
			writeError(w, err)
			return
		}
		data[id] = jOut
	}
	writeJson(w, data)
}

func (srv *WebServer) outputShow(w http.ResponseWriter, out output.Output) {
	writeOutputJson(w, out)
}

func (srv *WebServer) outputActivate(w http.ResponseWriter, out output.Output) {
	err := out.Activate()
	if err != nil {
		writeError(w, fmt.Errorf("Error activating output '%s': %s", out.Id(), err.Error()))
		return
	}
	writeOutputJson(w, out)
}

func (srv *WebServer) outputDeactivate(w http.ResponseWriter, out output.Output) {
	err := out.Deactivate()
	if err != nil {
		writeError(w, fmt.Errorf("Error deactivating output '%s': %s", out.Id(), err.Error()))
		return
	}
	writeOutputJson(w, out)
}

func writeOutputJson(w http.ResponseWriter, out output.Output) {
	jOut, err := newJsonOutput(out)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJson(w, jOut)
}

func (srv *WebServer) findTimer(w http.ResponseWriter, c martini.Context, params martini.Params) {
	tmr := srv.ctrl.Timer(params["id"])
	if tmr != nil {
		c.Map(tmr)
	} else {
		write404(w)
	}
}

func (srv *WebServer) timerIndex(w http.ResponseWriter) {
	writeJson(w, map[string]interface{}{})
}

func (srv *WebServer) timerShow(w http.ResponseWriter, tmr timer.Timer) {
	writeTimerJson(w, tmr)
}

func writeTimerJson(w http.ResponseWriter, timer timer.Timer) {
	data := make(map[string]interface{})
	data["id"] = timer.Id()
	data["output_active"] = timer.OutputActive()
	writeJson(w, data)
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
