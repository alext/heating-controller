package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/alext/heating-controller/controller"
)

func (srv *WebServer) zonesIndex(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFiles(
		filepath.Join(srv.templatesPath, "_base.tmpl"),
		filepath.Join(srv.templatesPath, "index.tmpl"),
	)
	if err != nil {
		log.Print("Error parsing template:", err)
		writeError(w, err)
		return
	}
	var b bytes.Buffer
	err = t.Execute(&b, srv.controller.Zones)
	if err != nil {
		log.Println("Error executing template:", err)
		writeError(w, err)
		return
	}
	w.Write(b.Bytes())
}

type jsonZone struct {
	Active bool `json:"active"`
}

func newJSONZone(z *controller.Zone) *jsonZone {
	return &jsonZone{
		Active: z.Active(),
	}
}

func (srv *WebServer) zonesAPIIndex(w http.ResponseWriter, req *http.Request) {
	data := make(map[string]*jsonZone)
	for name, z := range srv.controller.Zones {
		data[name] = newJSONZone(z)
	}
	writeJSON(w, data)
}

func (srv *WebServer) zoneBoost(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	durationString := req.FormValue("duration")
	d, err := time.ParseDuration(durationString)
	if err == nil {
		log.Printf("[webserver] Zone %s, boosting for %s", z.ID, d)
		z.Boost(d)
		http.Redirect(w, req, "/", http.StatusFound)
	} else {
		http.Error(w, fmt.Sprintf("Invalid boost duration '%s'", durationString), http.StatusBadRequest)
	}
}

func (srv *WebServer) zoneCancelBoost(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	log.Printf("[webserver] Zone %s, cancelling boost", z.ID)
	z.CancelBoost()
	http.Redirect(w, req, "/", http.StatusFound)
}

const thermostatIncrement = 500

func (srv *WebServer) thermostatInc(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	if z.Thermostat == nil {
		write404(w)
		return
	}
	target := z.Thermostat.Target()
	z.Thermostat.Set(target + thermostatIncrement)
	err := z.Save()
	if err != nil {
		writeError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}
func (srv *WebServer) thermostatDec(w http.ResponseWriter, req *http.Request, z *controller.Zone) {
	if z.Thermostat == nil {
		write404(w)
		return
	}
	target := z.Thermostat.Target()
	z.Thermostat.Set(target - thermostatIncrement)
	err := z.Save()
	if err != nil {
		writeError(w, err)
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}
