package webserver

import (
	"github.com/go-martini/martini"
)

func (srv *WebServer) buildRoutes(r martini.Router) {
	r.Get("/", srv.outputsIndex)

	r.Group("/api", func(r martini.Router) {
		r.Get("/outputs", srv.apiOutputIndex)
		r.Group("/outputs/:id", func(r martini.Router) {
			r.Get("", srv.apiOutputShow)
			r.Put("/activate", srv.apiOutputActivate)
			r.Put("/deactivate", srv.apiOutputDeactivate)
		}, srv.findOutput)
	})
}
