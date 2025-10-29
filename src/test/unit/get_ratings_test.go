package test

import (
	"bookem-rating-service/client/roomclient"
	"bookem-rating-service/internal"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetHostRatings_Success(t *testing.T) {
	svc, repo, users, rooms, _, _ := CreateTestRatingService()

	hostID := uint(22)

	users.
		On("FindById", mock.Anything, hostID).
		Return(NotGuest, nil) // NotGuest.Role == "host"

	repo.
		On("FindAllRatings", internal.Host, hostID).
		Return(SampleRatings, nil)
	repo.
		On("GetAverageRating", internal.Host, hostID).
		Return(4.0, nil)

	users.
		On("FindById", mock.Anything, uint(10)).
		Return(DefaultGuest, nil)
	users.
		On("FindById", mock.Anything, uint(99)).
		Return(NotGuest, nil)

	out, err := svc.GetHostRatings(context.Background(), internal.AuthContext{}, hostID)
	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.InDelta(t, 4.0, out.Average, 1e-9)
		assert.Len(t, out.Ratings, 2)
		assert.Equal(t, "guest1", out.Ratings[0].Username)
		assert.Equal(t, 5, out.Ratings[0].Score)
		assert.Equal(t, "great", out.Ratings[0].Comment)

		assert.Equal(t, "host1", out.Ratings[1].Username)
		assert.Equal(t, 3, out.Ratings[1].Score)
		assert.Equal(t, "ok", out.Ratings[1].Comment)
	}

	users.AssertCalled(t, "FindById", mock.Anything, hostID)
	repo.AssertCalled(t, "FindAllRatings", internal.Host, hostID)
	repo.AssertCalled(t, "GetAverageRating", internal.Host, hostID)
	rooms.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
}

func TestGetHostRatings_RaterLookupFails_UsernameEmpty(t *testing.T) {
	svc, repo, users, _, _, _ := CreateTestRatingService()

	hostID := uint(22)

	users.
		On("FindById", mock.Anything, hostID).
		Return(NotGuest, nil)

	repo.
		On("FindAllRatings", internal.Host, hostID).
		Return(SampleRatings, nil)
	repo.
		On("GetAverageRating", internal.Host, hostID).
		Return(4.0, nil)

	users.
		On("FindById", mock.Anything, uint(10)).
		Return(DefaultGuest, nil)
	users.
		On("FindById", mock.Anything, uint(99)).
		Return((*DefaultGuest), errors.New("user svc down"))

	out, err := svc.GetHostRatings(context.Background(), internal.AuthContext{}, hostID)
	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.Len(t, out.Ratings, 2)
		assert.Equal(t, "guest1", out.Ratings[0].Username)
		assert.Equal(t, "", out.Ratings[1].Username)
	}
}

func TestGetHostRatings_TargetNotFound(t *testing.T) {
	svc, repo, users, rooms, _, _ := CreateTestRatingService()

	hostID := uint(999)

	users.
		On("FindById", mock.Anything, hostID).
		Return(nil, errors.New("not found"))

	out, err := svc.GetHostRatings(context.Background(), internal.AuthContext{}, hostID)
	assert.Error(t, err)
	assert.Nil(t, out)

	users.AssertCalled(t, "FindById", mock.Anything, hostID)
	repo.AssertNotCalled(t, "FindAllRatings", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "GetAverageRating", mock.Anything, mock.Anything)
	rooms.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
}

func TestGetHostRatings_TargetNotHostRole(t *testing.T) {
	svc, repo, users, rooms, _, _ := CreateTestRatingService()

	hostID := uint(77)

	users.
		On("FindById", mock.Anything, hostID).
		Return(DefaultGuest, nil)

	out, err := svc.GetHostRatings(context.Background(), internal.AuthContext{}, hostID)
	assert.Error(t, err)
	assert.Nil(t, out)

	users.AssertCalled(t, "FindById", mock.Anything, hostID)
	repo.AssertNotCalled(t, "FindAllRatings", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "GetAverageRating", mock.Anything, mock.Anything)
	rooms.AssertNotCalled(t, "FindById", mock.Anything, mock.Anything)
}

func TestGetHostRatings_RepoErrors(t *testing.T) {
	// FindAllRatings error
	{
		svc, repo, users, _, _, _ := CreateTestRatingService() // rooms unused here
		hostID := uint(22)

		users.
			On("FindById", mock.Anything, hostID).
			Return(NotGuest, nil)

		repo.
			On("FindAllRatings", internal.Host, hostID).
			Return(nil, errors.New("db read error"))

		out, err := svc.GetHostRatings(context.Background(), internal.AuthContext{}, hostID)
		assert.Error(t, err)
		assert.Nil(t, out)

		repo.AssertCalled(t, "FindAllRatings", internal.Host, hostID)
		repo.AssertNotCalled(t, "GetAverageRating", mock.Anything, mock.Anything)
	}

	// GetAverageRating error
	{
		svc, repo, users, _, _, _ := CreateTestRatingService() // rooms unused here
		hostID := uint(22)

		users.
			On("FindById", mock.Anything, hostID).
			Return(NotGuest, nil)

		repo.
			On("FindAllRatings", internal.Host, hostID).
			Return([]internal.Rating{}, nil)
		repo.
			On("GetAverageRating", internal.Host, hostID).
			Return(0.0, errors.New("avg error"))

		out, err := svc.GetHostRatings(context.Background(), internal.AuthContext{}, hostID)
		assert.Error(t, err)
		assert.Nil(t, out)
	}
}

func TestGetRoomRatings_Success(t *testing.T) {
	svc, repo, users, rooms, _, _ := CreateTestRatingService()

	roomID := uint(55)

	rooms.
		On("FindById", mock.Anything, roomID).
		Return(&roomclient.RoomDTO{ID: roomID}, nil)

	rate := []internal.Rating{
		{ID: 1, TargetType: internal.Room, TargetID: roomID, RaterID: 10, Score: 3, Comment: "meh", CreatedAt: time.Now()},
	}
	repo.
		On("FindAllRatings", internal.Room, roomID).
		Return(rate, nil)
	repo.
		On("GetAverageRating", internal.Room, roomID).
		Return(3.0, nil)

	users.
		On("FindById", mock.Anything, uint(10)).
		Return(DefaultGuest, nil)

	out, err := svc.GetRoomRatings(context.Background(), internal.AuthContext{}, roomID)
	assert.NoError(t, err)
	if assert.NotNil(t, out) {
		assert.InDelta(t, 3.0, out.Average, 1e-9)
		assert.Len(t, out.Ratings, 1)
		assert.Equal(t, "guest1", out.Ratings[0].Username)
		assert.Equal(t, 3, out.Ratings[0].Score)
		assert.Equal(t, "meh", out.Ratings[0].Comment)
	}

	rooms.AssertCalled(t, "FindById", mock.Anything, roomID)
	repo.AssertCalled(t, "FindAllRatings", internal.Room, roomID)
	repo.AssertCalled(t, "GetAverageRating", internal.Room, roomID)
	users.AssertCalled(t, "FindById", mock.Anything, uint(10))
}

func TestGetRoomRatings_RoomNotFound(t *testing.T) {
	svc, repo, _, rooms, _, _ := CreateTestRatingService()

	roomID := uint(404)

	rooms.
		On("FindById", mock.Anything, roomID).
		Return((*roomclient.RoomDTO)(nil), errors.New("not found"))

	out, err := svc.GetRoomRatings(context.Background(), internal.AuthContext{}, roomID)
	assert.Error(t, err)
	assert.Nil(t, out)

	rooms.AssertCalled(t, "FindById", mock.Anything, roomID)
	repo.AssertNotCalled(t, "FindAllRatings", mock.Anything, mock.Anything)
	repo.AssertNotCalled(t, "GetAverageRating", mock.Anything, mock.Anything)
}
