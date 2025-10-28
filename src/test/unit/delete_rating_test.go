package test

import (
	"bookem-rating-service/internal"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteHostRating_Success(t *testing.T) {
	svc, repo, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(22)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(DefaultGuest, nil)

	repo.
		On("FindRatingByRater", internal.Host, targetID, auth.CallerID).
		Return(&internal.Rating{ID: 1, TargetType: internal.Host, TargetID: targetID, RaterID: auth.CallerID}, nil)

	repo.
		On("DeleteRating", internal.Host, targetID, auth.CallerID).
		Return(nil)

	err := svc.DeleteHostRating(context.Background(), auth, targetID)
	assert.NoError(t, err)

	users.AssertCalled(t, "FindById", mock.Anything, auth.CallerID)
	repo.AssertCalled(t, "FindRatingByRater", internal.Host, targetID, auth.CallerID)
	repo.AssertCalled(t, "DeleteRating", internal.Host, targetID, auth.CallerID)
}

func TestDeleteRoomRating_Success(t *testing.T) {
	svc, repo, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(33)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(DefaultGuest, nil)

	repo.
		On("FindRatingByRater", internal.Room, targetID, auth.CallerID).
		Return(&internal.Rating{ID: 2, TargetType: internal.Room, TargetID: targetID, RaterID: auth.CallerID}, nil)

	repo.
		On("DeleteRating", internal.Room, targetID, auth.CallerID).
		Return(nil)

	err := svc.DeleteRoomRating(context.Background(), auth, targetID)
	assert.NoError(t, err)
}

func TestDeleteRating_Unauthenticated_UserLookupFails(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(44)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(nil, errors.New("user svc down"))

	err := svc.DeleteHostRating(context.Background(), auth, targetID)
	assert.Error(t, err) // service returns ErrUnauthenticated
}

func TestDeleteRating_Unauthorized_NotGuest(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: NotGuest.Id, JWT: "jwt"} // NotGuest has role "host"
	targetID := uint(55)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(NotGuest, nil)

	err := svc.DeleteRoomRating(context.Background(), auth, targetID)
	assert.Error(t, err) // service returns ErrUnauthorized
}

func TestDeleteRating_InvalidTargetID(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(0)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(DefaultGuest, nil)

	err := svc.DeleteHostRating(context.Background(), auth, targetID)
	assert.Error(t, err) // ErrBadRequestCustom("targetId is required")
}

func TestDeleteRating_NotFound_ForCaller(t *testing.T) {
	svc, repo, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(66)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(DefaultGuest, nil)

	repo.
		On("FindRatingByRater", internal.Host, targetID, auth.CallerID).
		Return(nil, errors.New("record not found"))

	err := svc.DeleteHostRating(context.Background(), auth, targetID)
	assert.Error(t, err) // ErrNotFound("rating", targetID)
}

func TestDeleteRating_NotOwner(t *testing.T) {
	svc, repo, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(77)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(DefaultGuest, nil)

	repo.
		On("FindRatingByRater", internal.Room, targetID, auth.CallerID).
		Return(&internal.Rating{ID: 3, TargetType: internal.Room, TargetID: targetID, RaterID: 999}, nil)

	err := svc.DeleteRoomRating(context.Background(), auth, targetID)
	assert.Error(t, err) // ErrUnauthorized
}

func TestDeleteRating_RepoDeleteError(t *testing.T) {
	svc, repo, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	targetID := uint(88)

	users.
		On("FindById", mock.Anything, auth.CallerID).
		Return(DefaultGuest, nil)

	repo.
		On("FindRatingByRater", internal.Host, targetID, auth.CallerID).
		Return(&internal.Rating{ID: 4, TargetType: internal.Host, TargetID: targetID, RaterID: auth.CallerID}, nil)

	repo.
		On("DeleteRating", internal.Host, targetID, auth.CallerID).
		Return(errors.New("db write failed"))

	err := svc.DeleteHostRating(context.Background(), auth, targetID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db write failed")
}
