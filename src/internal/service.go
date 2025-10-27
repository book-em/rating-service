package internal

import (
	"bookem-rating-service/client/roomclient"
	"bookem-rating-service/client/userclient"
	"bookem-rating-service/util"

	"strings"
	"context"
)

// AuthContext is used in cases where callerID is not enoug
type AuthContext struct {
	CallerID uint
	JWT      string
}

type Service interface {
    CreateHostRating(ctx context.Context, authctx AuthContext, dto CreateRatingDTO) (*Rating, error)
    CreateRoomRating(ctx context.Context, authctx AuthContext, dto CreateRatingDTO) (*Rating, error)
}

type service struct {
	repo       Repository
	userClient userclient.UserClient
	roomClient roomclient.RoomClient
}

func NewService(
	roomRepo Repository,
	userClient userclient.UserClient,
	roomClient roomclient.RoomClient) Service {
	return &service{roomRepo, userClient, roomClient}
}

func (s *service) CreateHostRating(ctx context.Context, authctx AuthContext, dto CreateRatingDTO) (*Rating, error) {
	return s.createRating(ctx, authctx, Host, dto)
}

func (s *service) CreateRoomRating(ctx context.Context, authctx AuthContext, dto CreateRatingDTO) (*Rating, error) {
	return s.createRating(ctx, authctx, Room, dto)
}

func (s *service) createRating(ctx context.Context, authctx AuthContext, rt RatingType, dto CreateRatingDTO) (*Rating, error) {
	callerID := authctx.CallerID

	util.TEL.Info("guest wants to create/update a rating", nil, "caller_id", callerID, "target_type", string(rt), "target_id", dto.TargetID)

	util.TEL.Push(ctx, "validate-user-role")
	defer util.TEL.Pop()

	util.TEL.Debug("fetch user by id", nil, "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user not found", err, "id", callerID)
		return nil, ErrUnauthenticated
	}
	if user.Role != string(util.Guest) {
		util.TEL.Error("user role not allowed to rate", nil, "role", user.Role)
		return nil, ErrUnauthorized
	}

	// DTO validation 
	util.TEL.Push(ctx, "validate-dto")
	defer util.TEL.Pop()

	if dto.TargetID == 0 {
		util.TEL.Error("target id is empty", nil)
		return nil, ErrBadRequestCustom("targetId is required")
	}

	if dto.Score < 1 || dto.Score > 5 {
		util.TEL.Error("score out of range", nil, "score", dto.Score)
		return nil, ErrBadRequestCustom("score must be between 1 and 5")
	}

	util.TEL.Push(ctx, "validate-target-exists")
	defer util.TEL.Pop()

	switch rt {
	case Host:
		util.TEL.Debug("check if host exists", nil, "id", dto.TargetID)
		host, err := s.userClient.FindById(util.TEL.Ctx(), dto.TargetID)
		if err != nil {
			util.TEL.Error("host not found", err, "id", dto.TargetID)
			return nil, ErrNotFound("host", dto.TargetID)
		}
		if host.Role != string(util.Host) {
			util.TEL.Error("target user is not a host", nil, "role", host.Role)
			return nil, ErrBadRequestCustom("target user is not a host")
		}
	case Room:
		util.TEL.Debug("check if room exists", nil, "id", dto.TargetID)
		room, err := s.roomClient.FindById(util.TEL.Ctx(), dto.TargetID)
		if err != nil {
			util.TEL.Error("room not found", err, "id", dto.TargetID)
			return nil, ErrNotFound("room", dto.TargetID)
		}
		if room.ID == 0 {
			util.TEL.Error("room id missing in response", nil)
			return nil, ErrNotFound("room", dto.TargetID)
		}
	default:
		return nil, ErrBadRequestCustom("invalid rating type")
	}

	// Upsert rating in DB 
	util.TEL.Push(ctx, "upsert-rating-in-db")
	defer util.TEL.Pop()

	created, err := s.repo.CreateOrUpdateRating(rt, dto.TargetID, callerID, dto.Score, strings.TrimSpace(dto.Comment))
	if err != nil {
		util.TEL.Error("failed to create/update rating", err)
		return nil, err
	}

	r, err := s.repo.FindRatingByRater(rt, dto.TargetID, callerID)
	if err != nil {
		util.TEL.Error("failed to fetch rating after upsert", err)
		return nil, err
	}

	if created {
		util.TEL.Info("rating created", "rating_id", r.ID, "target_type", string(rt), "target_id", dto.TargetID, "rater_id", callerID)
	} else {
		util.TEL.Info("rating updated", "rating_id", r.ID, "target_type", string(rt), "target_id", dto.TargetID, "rater_id", callerID)
	}

	return r, nil
}
