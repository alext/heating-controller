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
	var err error
	e := scheduler.Event{}
	e.Hour, err = strconv.Atoi(req.FormValue("hour"))
	if err != nil {
		http.Error(w, "hour must be a number: "+err.Error(), http.StatusBadRequest)
		return
	}
	e.Min, err = strconv.Atoi(req.FormValue("min"))
	if err != nil {
		http.Error(w, "minute must be a number: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.FormValue("action") == "on" {
		e.Action = scheduler.TurnOn
	}
	err = z.Scheduler.AddEvent(e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	srv.scheduleEdit(w, req, z)
}
