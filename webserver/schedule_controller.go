package webserver

import (
	"net/http"
	"strconv"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/scheduler"
	"github.com/alext/heating-controller/zone"
	"github.com/gorilla/mux"
)

func (srv *WebServer) scheduleEdit(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	template := srv.templates["schedule"]
	err := template.Execute(w, z)
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

func (srv *WebServer) scheduleRemoveEvent(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	hour, _ := strconv.Atoi(mux.Vars(req)["hour"])
	min, _ := strconv.Atoi(mux.Vars(req)["min"])
	for _, e := range z.Scheduler.ReadEvents() {
		if e.Hour == hour && e.Min == min {
			z.Scheduler.RemoveEvent(e)
			break
		}
	}

	http.Redirect(w, req, "/zones/"+z.ID+"/schedule", 302)
}
