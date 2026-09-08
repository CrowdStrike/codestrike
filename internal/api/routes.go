package api

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
)

func RegisterRoutes(container *restful.Container, handler *Handler) {
	ws := new(restful.WebService)

	ws.
		Path("/api/v1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	ws.
		Route(ws.GET("healthz").
			To(handler.Healthz).
			Doc("Health check").
			Metadata(restfulspec.KeyOpenAPITags, []string{"health"}).
			Writes(HealthResponse{}).
			Returns(200, "OK", HealthResponse{}))

	ws.
		Route(ws.POST("review").
			To(handler.Review).
			Doc("Run an AI review on a pull request").
			Metadata(restfulspec.KeyOpenAPITags, []string{"review"}).
			Reads(ReviewRequest{}).
			Writes(ReviewResponse{}).
			Returns(200, "Review completed", ReviewResponse{}).
			Returns(400, "Invalid request", ErrorResponse{}).
			Returns(500, "Review failed", ErrorResponse{}))

	container.Add(ws)
}
