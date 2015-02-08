package webserver

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/scheduler"
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

func (srv *WebServer) scheduleAddEvent(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	e := scheduler.Event{}
	e.Hour, _ = strconv.Atoi(req.FormValue("hour"))
	e.Min, _ = strconv.Atoi(req.FormValue("min"))
	if req.FormValue("action") == "on" {
		e.Action = scheduler.TurnOn
	}
	z.Scheduler.AddEvent(e)
	srv.scheduleEdit(w, req, z)
}
