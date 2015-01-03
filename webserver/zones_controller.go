package webserver

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/alext/heating-controller/logger"
	"github.com/alext/heating-controller/zone"
)

func (srv *WebServer) zonesIndex(w http.ResponseWriter, req *http.Request) {
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
	err = t.Execute(&b, srv.zones)
	if err != nil {
		logger.Warn("Error executing template:", err)
		writeError(w, err)
		return
	}
	w.Write(b.Bytes())
}

func (srv *WebServer) zoneBoost(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	durationString := req.FormValue("duration")
	d, err := time.ParseDuration(durationString)
	if err == nil {
		z.Scheduler.Boost(d)
		http.Redirect(w, req, "/", http.StatusFound)
	} else {
		http.Error(w, fmt.Sprintf("Invalid boost duration '%s'", durationString), http.StatusBadRequest)
	}
}

func (srv *WebServer) zoneActivate(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	err := z.Out.Activate()
	if err != nil {
		writeError(w, fmt.Errorf("Error activating output '%s': %s", z.ID, err.Error()))
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}

func (srv *WebServer) zoneDeactivate(w http.ResponseWriter, req *http.Request, z *zone.Zone) {
	err := z.Out.Deactivate()
	if err != nil {
		writeError(w, fmt.Errorf("Error deactivating output '%s': %s", z.ID, err.Error()))
		return
	}
	http.Redirect(w, req, "/", http.StatusFound)
}
