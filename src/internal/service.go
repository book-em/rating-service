package internal

import (
	"bookem-rating-service/client/roomclient"
	"bookem-rating-service/client/userclient"
	"bookem-rating-service/client/reservationclient"
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
	DeleteHostRating(ctx context.Context, authctx AuthContext, targetID uint) error 
	DeleteRoomRating(ctx context.Context, authctx AuthContext, targetID uint) error
	GetRoomRatings(ctx context.Context, _ AuthContext, roomID uint) (*RatingsWithAverageDTO, error) 
	GetHostRatings(ctx context.Context, _ AuthContext, hostID uint) (*RatingsWithAverageDTO, error) 
}

type service struct {
	repo       Repository
	userClient userclient.UserClient
	roomClient roomclient.RoomClient
	reservationClient reservationclient.ReservationClient
}

func NewService(
	roomRepo Repository,
	userClient userclient.UserClient,
	roomClient roomclient.RoomClient,
	reservationClient reservationclient.ReservationClient,
	) Service {
	return &service{roomRepo, userClient, roomClient, reservationClient}
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
		util.TEL.Debug("check if guest is eligible to rate host", nil, "guest_id", callerID, "host_id", dto.TargetID)
		ok, err := s.reservationClient.CanUserRateHost(util.TEL.Ctx(), callerID, dto.TargetID)
		if err != nil {
			util.TEL.Error("eligibility check failed (host)", err, "guest_id", callerID, "host_id", dto.TargetID)
			return nil, ErrBadRequestCustom("eligibility check failed by reservation-service")
		}
		if !ok {
			util.TEL.Error("guest not eligible to rate host", nil, "guest_id", callerID, "host_id", dto.TargetID)
			return nil, ErrBadRequestCustom("guest not eligible to rate host")
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
		util.TEL.Debug("check if guest is eligible to rate room", nil, "guest_id", callerID, "room_id", dto.TargetID)
		ok, err := s.reservationClient.CanUserRateRoom(util.TEL.Ctx(), callerID, dto.TargetID)
		if err != nil {
			util.TEL.Error("eligibility check failed (room)", err, "guest_id", callerID, "room_id", dto.TargetID)
			return nil, ErrBadRequestCustom("eligibility check failed by reservation-service")
		}
		if !ok {
			util.TEL.Error("guest not eligible to rate room", nil, "guest_id", callerID, "room_id", dto.TargetID)
			return nil, ErrBadRequestCustom("guest not eligible to rate room")
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

func (s *service) DeleteHostRating(ctx context.Context, authctx AuthContext, targetID uint) error {
	return s.deleteRating(ctx, authctx, Host, targetID)
}

func (s *service) DeleteRoomRating(ctx context.Context, authctx AuthContext, targetID uint) error {
	return s.deleteRating(ctx, authctx, Room, targetID)
}

func (s *service) deleteRating(ctx context.Context, authctx AuthContext, rt RatingType, targetID uint) error {
	callerID := authctx.CallerID
	util.TEL.Info("guest wants to delete a rating", nil, "caller_id", callerID, "target_type", string(rt), "target_id", targetID)

	util.TEL.Push(ctx, "validate-user-role")
	defer util.TEL.Pop()

	util.TEL.Debug("fetch user by id", nil, "id", callerID)
	user, err := s.userClient.FindById(util.TEL.Ctx(), callerID)
	if err != nil {
		util.TEL.Error("user not found", err, "id", callerID)
		return ErrUnauthenticated
	}
	if user.Role != string(util.Guest) {
		util.TEL.Error("user role not allowed to delete rating", nil, "role", user.Role)
		return ErrUnauthorized
	}

	util.TEL.Push(ctx, "validate-input")
	defer util.TEL.Pop()
	if targetID == 0 {
		util.TEL.Error("target id is empty", nil)
		return ErrBadRequestCustom("targetId is required")
	}

	// Ensure rating exists & belongs to caller
	util.TEL.Push(ctx, "find-rating")
	defer util.TEL.Pop()

	r, err := s.repo.FindRatingByRater(rt, targetID, callerID)
	if err != nil {
		util.TEL.Error("rating not found for caller", err, "target_id", targetID, "rater_id", callerID)
		return ErrNotFound("rating", targetID)
	}
	if r.RaterID != callerID {
		util.TEL.Error("caller is not the author of rating", nil, "caller_id", callerID, "rater_id", r.RaterID)
		return ErrUnauthorized
	}

	util.TEL.Push(ctx, "delete-rating-in-db")
	defer util.TEL.Pop()

	if err := s.repo.DeleteRating(rt, targetID, callerID); err != nil {
		util.TEL.Error("failed to delete rating", err)
		return err
	}

	util.TEL.Info("rating deleted", "target_type", string(rt), "target_id", targetID, "rater_id", callerID)
	return nil
}

func (s *service) GetHostRatings(ctx context.Context, _ AuthContext, hostID uint) (*RatingsWithAverageDTO, error) {
	return s.getRatings(ctx, Host, hostID)
}

func (s *service) GetRoomRatings(ctx context.Context, _ AuthContext, roomID uint) (*RatingsWithAverageDTO, error) {
	return s.getRatings(ctx, Room, roomID)
}

func (s *service) getRatings(ctx context.Context, rt RatingType, targetID uint) (*RatingsWithAverageDTO, error) {
	util.TEL.Push(ctx, "get-ratings")
	defer util.TEL.Pop()

	switch rt {
	case Host:
		u, err := s.userClient.FindById(util.TEL.Ctx(), targetID)
		if err != nil {
			util.TEL.Error("host not found", err, "id", targetID)
			return nil, ErrNotFound("host", targetID)
		}
		if u.Role != string(util.Host) {
			return nil, ErrBadRequestCustom("target user is not a host")
		}
	case Room:
		room, err := s.roomClient.FindById(util.TEL.Ctx(), targetID)
		if err != nil || room.ID == 0 {
			util.TEL.Error("room not found", err, "id", targetID)
			return nil, ErrNotFound("room", targetID)
		}
	default:
		return nil, ErrBadRequestCustom("invalid rating type")
	}

	ratings, err := s.repo.FindAllRatings(rt, targetID)
	if err != nil {
		util.TEL.Error("failed fetching ratings", err)
		return nil, err
	}
	avg, err := s.repo.GetAverageRating(rt, targetID)
	if err != nil {
		util.TEL.Error("failed computing average", err)
		return nil, err
	}

	out := make([]RatingDTO, 0, len(ratings))
	for _, r := range ratings {
		user, err := s.userClient.FindById(util.TEL.Ctx(), r.RaterID)
		username := ""
		if err != nil {
			util.TEL.Error("failed fetching rater user", err, "rater_id", r.RaterID)
		} else {
			username = user.Username
		}

		t := r.UpdatedAt
		if t.IsZero() {
			t = r.CreatedAt
		}
		out = append(out, RatingDTO{
			Username: username,
			Score:    r.Score,
			Comment:  r.Comment,
			Time:     t,
		})
	}

	return &RatingsWithAverageDTO{
		Average: avg,
		Ratings: out,
	}, nil
}

