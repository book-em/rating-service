package internal

import (
	"bookem-rating-service/util"
	"net/http"
	"github.com/gin-gonic/gin"
	"strconv"   
)

type Route struct{ handler Handler }

func NewRoute(handler Handler) *Route { return &Route{handler} }

func (r *Route) Route(rg *gin.RouterGroup) {
	rg.POST("/ratings/:id/host", r.handler.createHostRating)
	rg.POST("/ratings/:id/room", r.handler.createRoomRating)
	rg.DELETE("/ratings/:id/host", r.handler.deleteHostRating)
	rg.DELETE("/ratings/:id/room", r.handler.deleteRoomRating)

}

type Handler struct{ service Service }

func NewHandler(s Service) Handler { return Handler{s} }

func (h *Handler) createHostRating(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "create-host-rating-api")
	defer util.TEL.Pop()

	jwtString, err := util.GetJwtString(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT string", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}
	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed parsing JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}
	if jwt.Role != util.Guest {
		util.TEL.Error("user is not guest", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto CreateRatingDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		util.TEL.Error("failed binding JSON", err)
		AbortError(ctx, ErrBadRequest)
		return
	}
	idParam := ctx.Param("id")
	targetID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		util.TEL.Error("invalid target ID", err, "param", idParam)
		AbortError(ctx, ErrBadRequestCustom("invalid ID"))
		return
	}
dto.TargetID = uint(targetID)
	rating, err := h.service.CreateHostRating(util.TEL.Ctx(), AuthContext{CallerID: jwt.ID, JWT: jwtString}, dto)
	if err != nil {
		util.TEL.Error("failed creating/updating host rating", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, rating)
}

func (h *Handler) createRoomRating(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "create-room-rating-api")
	defer util.TEL.Pop()

	jwtString, err := util.GetJwtString(ctx)
	if err != nil {
		util.TEL.Error("failed fetching JWT string", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}
	jwt, err := util.GetJwt(ctx)
	if err != nil {
		util.TEL.Error("failed parsing JWT", err)
		AbortError(ctx, ErrUnauthenticated)
		return
	}
	if jwt.Role != util.Guest {
		util.TEL.Error("user is not guest", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	var dto CreateRatingDTO
	if err := ctx.ShouldBindJSON(&dto); err != nil {
		util.TEL.Error("failed binding JSON", err)
		AbortError(ctx, ErrBadRequest)
		return
	}
	
	idParam := ctx.Param("id")
	targetID, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		util.TEL.Error("invalid target ID", err, "param", idParam)
		AbortError(ctx, ErrBadRequestCustom("invalid ID"))
		return
	}
	dto.TargetID = uint(targetID)

	rating, err := h.service.CreateRoomRating(util.TEL.Ctx(), AuthContext{CallerID: jwt.ID, JWT: jwtString}, dto)
	if err != nil {
		util.TEL.Error("failed creating/updating room rating", err)
		AbortError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, rating)
}

func (h *Handler) deleteHostRating(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "delete-host-rating-api")
	defer util.TEL.Pop()

	jwtString, err := util.GetJwtString(ctx)
	if err != nil { util.TEL.Error("failed fetching JWT", err); AbortError(ctx, ErrUnauthenticated); return }
	jwt, err := util.GetJwt(ctx)
	if err != nil { util.TEL.Error("failed parsing JWT", err); AbortError(ctx, ErrUnauthenticated); return }
	if jwt.Role != util.Guest {
		util.TEL.Error("user is not guest", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	idParam := ctx.Param("id")
	targetID64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || targetID64 == 0 {
		util.TEL.Error("invalid target id", err, "param", idParam)
		AbortError(ctx, ErrBadRequestCustom("invalid id"))
		return
	}
	if err := h.service.DeleteHostRating(util.TEL.Ctx(), AuthContext{CallerID: jwt.ID, JWT: jwtString}, uint(targetID64)); err != nil {
		AbortError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *Handler) deleteRoomRating(ctx *gin.Context) {
	util.TEL.Push(ctx.Request.Context(), "delete-room-rating-api")
	defer util.TEL.Pop()

	jwtString, err := util.GetJwtString(ctx)
	if err != nil { util.TEL.Error("failed fetching JWT", err); AbortError(ctx, ErrUnauthenticated); return }
	jwt, err := util.GetJwt(ctx)
	if err != nil { util.TEL.Error("failed parsing JWT", err); AbortError(ctx, ErrUnauthenticated); return }
	if jwt.Role != util.Guest {
		util.TEL.Error("user is not guest", nil, "role", jwt.Role)
		AbortError(ctx, ErrUnauthorized)
		return
	}

	idParam := ctx.Param("id")
	targetID64, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil || targetID64 == 0 {
		util.TEL.Error("invalid target id", err, "param", idParam)
		AbortError(ctx, ErrBadRequestCustom("invalid id"))
		return
	}
	if err := h.service.DeleteRoomRating(util.TEL.Ctx(), AuthContext{CallerID: jwt.ID, JWT: jwtString}, uint(targetID64)); err != nil {
		AbortError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
