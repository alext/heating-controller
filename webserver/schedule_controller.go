package webserver

import (
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/alext/heating-controller/controller"
	"github.com/alext/heating-controller/units"
	"github.com/gorilla/mux"
)

func (srv *WebServer) scheduleEdit(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	t, err := template.ParseFiles(
		filepath.Join(srv.templatesPath, "_base.tmpl"),
		filepath.Join(srv.templatesPath, "schedule.tmpl"),
	)
	if err != nil {
		log.Println("Error parsing template:", err)
		writeError(w, err)
		return
	}
	err = t.Execute(w, z)
	if err != nil {
		log.Println("Error executing template:", err)
	}
}

func (srv *WebServer) scheduleAddEvent(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	e := controller.Event{}

	hour, err := strconv.Atoi(req.FormValue("hour"))
	if err != nil {
		http.Error(w, "hour must be a number: "+err.Error(), http.StatusBadRequest)
		return
	}
	min, err := strconv.Atoi(req.FormValue("min"))
	if err != nil {
		http.Error(w, "minute must be a number: "+err.Error(), http.StatusBadRequest)
		return
	}
	e.Time = units.NewTimeOfDay(hour, min)

	err = e.Action.UnmarshalText([]byte(req.FormValue("action")))
	if err != nil {
		http.Error(w, "invalid action: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.FormValue("therm_action") != "" {
		e.ThermAction = &controller.ThermostatAction{}
		err = e.ThermAction.Action.UnmarshalText([]byte(req.FormValue("therm_action")))
		if err != nil {
			http.Error(w, "invalid thermostat action: "+err.Error(), http.StatusBadRequest)
			return
		}
		param, err := units.ParseTemperature(req.FormValue("therm_param"))
		if err != nil {
			http.Error(w, "thermostat param must be a number: "+err.Error(), http.StatusBadRequest)
			return
		}
		e.ThermAction.Param = param
	}

	err = z.AddEvent(e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = z.Save()
	if err != nil {
		writeError(w, err)
		return
	}
	http.Redirect(w, req, "/zones/"+z.ID+"/schedule", 302)
}

func (srv *WebServer) scheduleRemoveEvent(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	t, err := units.ParseTimeOfDay(mux.Vars(req)["time"])
	if err != nil {
		write404(w)
		return
	}
	for _, e := range z.ReadEvents() {
		if e.Time == t {
			z.RemoveEvent(e)
			break
		}
	}

	err = z.Save()
	if err != nil {
		writeError(w, err)
		return
	}

	http.Redirect(w, req, "/zones/"+z.ID+"/schedule", 302)
}
