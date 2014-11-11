package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/output"
)

func (srv *WebServer) outputsIndex(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles(
		srv.templatesPath+"/_base.html",
		srv.templatesPath+"/index.html",
	)
	if err != nil {
		logger.Warn("Error parsing template:", err)
		writeError(w, err)
		return
	}
	var b bytes.Buffer
	err = t.Execute(&b, srv.outputs)
	if err != nil {
		logger.Warn("Error executing template:", err)
		writeError(w, err)
		return
	}
	w.Write(b.Bytes())
}

func (srv *WebServer) outputActivate(w http.ResponseWriter, req *http.Request, out output.Output) {
	err := out.Activate()
	if err != nil {
		writeError(w, fmt.Errorf("Error activating output '%s': %s", out.Id(), err.Error()))
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (srv *WebServer) outputDeactivate(w http.ResponseWriter, req *http.Request, out output.Output) {
	err := out.Deactivate()
	if err != nil {
		writeError(w, fmt.Errorf("Error deactivating output '%s': %s", out.Id(), err.Error()))
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}
