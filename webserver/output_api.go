package webserver

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alext/heating-controller/output"
)

func (srv *WebServer) apiOutputIndex(w http.ResponseWriter) {
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

func (srv *WebServer) apiOutputShow(w http.ResponseWriter, out output.Output) {
	writeOutputJson(w, out)
}

func (srv *WebServer) apiOutputActivate(w http.ResponseWriter, out output.Output) {
	err := out.Activate()
	if err != nil {
		writeError(w, fmt.Errorf("Error activating output '%s': %s", out.Id(), err.Error()))
		return
	}
	writeOutputJson(w, out)
}

func (srv *WebServer) apiOutputDeactivate(w http.ResponseWriter, out output.Output) {
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

func writeJson(w http.ResponseWriter, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		writeError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
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
