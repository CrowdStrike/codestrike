package api

import (
	"github.com/emicklei/go-restful/v3"
)

func RegisterRoutes(container *restful.Container, handler *Handler) {
	healthWs := new(restful.WebService)
	healthWs.Path("/healthz").Produces(restful.MIME_JSON)
	healthWs.Route(healthWs.GET("").To(handler.Healthz))
	container.Add(healthWs)

	ws := new(restful.WebService)
	ws.Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)
	ws.Route(ws.POST("review").To(handler.Review))
	container.Add(ws)
}
