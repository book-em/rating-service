package internal

import (
	"bookem-rating-service/util"
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"   
	"strings"

)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(g *gin.RouterGroup) {
	g.POST("/ratings/:id", r.handler.createRating)           // ?type=host|room
	g.DELETE("/ratings/:id", r.handler.deleteRating)         // ?type=host|room
	g.GET("/ratings/all/:id", r.handler.getRatingsWithAvg)   // ?type=host|room
}


type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createRating(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "create-rating-api")
	defer util.TEL.Pop()

	auth, ok := requireGuestAuth(ctx); if !ok { return }

	rt, ok := parseRatingType(ctx); if !ok { return }

	targetID, ok := parseUintParam(ctx, "id"); if !ok { return }

	var dto CreateRatingDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		AbortError(ctx, ErrBadRequest); return
	}
	dto.TargetID = targetID

	var (
		r   *Rating
		err error
	)
	switch rt {
	case Host:
		r, err = h.service.CreateHostRating(util.TEL.Ctx(), auth, dto)
	case Room:
		r, err = h.service.CreateRoomRating(util.TEL.Ctx(), auth, dto)
	}
	if err != nil { AbortError(ctx, err); return }

	ctx.JSON(http.StatusCreated, r)
}

func (h *Handler) deleteRating(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "delete-rating-api")
	defer util.TEL.Pop()

	auth, ok := requireGuestAuth(ctx); if !ok { return }

	rt, ok := parseRatingType(ctx); if !ok { return }

	targetID, ok := parseUintParam(ctx, "id"); if !ok { return }

	var err error
	switch rt {
	case Host:
		err = h.service.DeleteHostRating(util.TEL.Ctx(), auth, targetID)
	case Room:
		err = h.service.DeleteRoomRating(util.TEL.Ctx(), auth, targetID)
	}
	if err != nil { AbortError(ctx, err); return }

	ctx.Status(http.StatusNoContent)
}

func (h *Handler) getRatingsWithAvg(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "get-ratings-with-avg-api")
	defer util.TEL.Pop()

	rt, ok := parseRatingType(ctx); if !ok { return }

	targetID, ok := parseUintParam(ctx, "id"); if !ok { return }

	var (
		res *RatingsWithAverageDTO 
		err error
	)
	switch rt {
	case Host:
		res, err = h.service.GetHostRatings(util.TEL.Ctx(), AuthContext{}, targetID)
	case Room:
		res, err = h.service.GetRoomRatings(util.TEL.Ctx(), AuthContext{}, targetID)
	}
	if err != nil { AbortError(ctx, err); return }

	ctx.JSON(http.StatusOK, res) 
}


func parseUintParam(ctx *gin.Context, name string) (uint, bool) {
	raw := ctx.Param(name)
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || v == 0 {
		AbortError(ctx, ErrBadRequestCustom("invalid "+name))
		return 0, false
	}
	return uint(v), true
}

func requireGuestAuth(ctx *gin.Context) (AuthContext, bool) {
	jwtStr, err := util.GetJwtString(ctx)
	if err != nil { AbortError(ctx, ErrUnauthenticated); return AuthContext{}, false }
	jwt, err := util.GetJwt(ctx)
	if err != nil { AbortError(ctx, ErrUnauthenticated); return AuthContext{}, false }
	if jwt.Role != util.Guest { AbortError(ctx, ErrUnauthorized); return AuthContext{}, false }
	return AuthContext{CallerID: jwt.ID, JWT: jwtStr}, true
}

func parseRatingType(ctx *gin.Context) (RatingType, bool) {
	t := strings.ToLower(strings.TrimSpace(ctx.Query("type")))
	switch t {
	case string(Host):
		return Host, true
	case string(Room):
		return Room, true
	default:
		AbortError(ctx, ErrBadRequestCustom("query param 'type' must be 'host' or 'room'"))
		return "", false
	}
}
