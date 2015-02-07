package webserver

import (
	"html/template"
	"net/http"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/zone"
)

func (srv *WebServer) scheduleEdit(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	t, err := template.ParseFiles(
		srv.templatesPath+"/_base.html",
		srv.templatesPath+"/schedule.html",
	)
	if err != nil {
		logger.Warn("Error parsing template:", err)
		writeError(w, err)
		return
	}
	err = t.Execute(w, z)
	if err != nil {
		logger.Warn("Error executing template:", err)
	}
}
