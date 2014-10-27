package webserver

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/alext/heating-controller/logger"
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
