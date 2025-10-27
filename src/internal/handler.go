package internal

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.GET("/req/rating", r.handler.getRating)
}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) getRating(ctx *gin.Context) {

	ctx.JSON(http.StatusOK, "works!")

}