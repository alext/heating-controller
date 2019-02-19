package webserver

import (
	"errors"
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
	err := populateEventFromRequest(&e, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
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

func (srv *WebServer) scheduleEditEvent(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	t, err := units.ParseTimeOfDay(mux.Vars(req)["time"])
	if err != nil {
		write404(w)
		return
	}
	e, ok := z.FindEvent(t)
	if !ok {
		write404(w)
		return
	}

	ed := eventData{
		Zone:        z,
		HourValue:   strconv.Itoa(e.Time.Hour()),
		MinuteValue: strconv.Itoa(e.Time.Minute()),
		Action:      e.Action.String(),
	}
	if e.ThermAction != nil {
		ed.ThermAction = e.ThermAction.Action.String()
		ed.ThermParam = strconv.FormatFloat(float64(e.ThermAction.Param)/1000, 'f', -1, 64) //e.ThermAction.Param.String()
	}
	srv.renderEventEdit(w, ed)
}

func (srv *WebServer) scheduleUpdateEvent(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	t, err := units.ParseTimeOfDay(mux.Vars(req)["time"])
	if err != nil {
		write404(w)
		return
	}

	e, ok := z.FindEvent(t)
	if !ok {
		write404(w)
		return
	}

	err = populateEventFromRequest(&e, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = z.ReplaceEvent(t, e)
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
	err = z.RemoveEvent(t)
	if err != nil {
		writeError(w, err)
		return
	}

	err = z.Save()
	if err != nil {
		writeError(w, err)
		return
	}

	http.Redirect(w, req, "/zones/"+z.ID+"/schedule", 302)
}

type eventData struct {
	Zone        *controller.Zone
	HourValue   string
	MinuteValue string
	Action      string
	ThermAction string
	ThermParam  string
}

func (srv *WebServer) renderEventEdit(w http.ResponseWriter, ed eventData) {
	t, err := template.ParseFiles(
		filepath.Join(srv.templatesPath, "_base.tmpl"),
		filepath.Join(srv.templatesPath, "event_edit.tmpl"),
	)
	if err != nil {
		log.Println("Error parsing template:", err)
		writeError(w, err)
		return
	}
	err = t.Execute(w, ed)
	if err != nil {
		log.Println("Error executing template:", err)
	}
}

func populateEventFromRequest(e *controller.Event, req *http.Request) error {
	hour, err := strconv.Atoi(req.FormValue("hour"))
	if err != nil {
		return errors.New("hour must be a number: " + err.Error())
	}
	min, err := strconv.Atoi(req.FormValue("min"))
	if err != nil {
		return errors.New("minute must be a number: " + err.Error())
	}
	e.Time = units.NewTimeOfDay(hour, min)

	err = e.Action.UnmarshalText([]byte(req.FormValue("action")))
	if err != nil {
		return errors.New("invalid action: " + err.Error())
	}
	if req.FormValue("therm_action") != "" {
		e.ThermAction = &controller.ThermostatAction{}
		err = e.ThermAction.Action.UnmarshalText([]byte(req.FormValue("therm_action")))
		if err != nil {
			return errors.New("invalid thermostat action: " + err.Error())
		}
		param, err := units.ParseTemperature(req.FormValue("therm_param"))
		if err != nil {
			return errors.New("thermostat param must be a number: " + err.Error())
		}
		e.ThermAction.Param = param
	} else {
		e.ThermAction = nil
	}
	return nil
}
