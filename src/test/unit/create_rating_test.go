package test

import (
	"bookem-rating-service/internal"
	"bookem-rating-service/client/userclient"
	"bookem-rating-service/client/roomclient"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateHostRating_Success(t *testing.T) {
	svc, repo, users, _, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 5, Comment: " great "}

	// caller must be guest
	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	// target must exist and be host
	users.On("FindById", mock.Anything, dto.TargetID).
		Return(&userclient.UserDTO{Id: dto.TargetID, Username: "hostUser", Role: "host"}, nil)
	// eligibility ok
	resCli.On("CanUserRateHost", mock.Anything, auth.CallerID, dto.TargetID).Return(true, nil)
	// upsert ok (created=true)
	repo.On("CreateOrUpdateRating", internal.Host, dto.TargetID, auth.CallerID, dto.Score, "great").Return(true, nil)
	// return stored entity
	repo.On("FindRatingByRater", internal.Host, dto.TargetID, auth.CallerID).
		Return(&internal.Rating{ID: 1, TargetType: internal.Host, TargetID: dto.TargetID, RaterID: auth.CallerID, Score: dto.Score, Comment: "great"}, nil)

	r, err := svc.CreateHostRating(context.Background(), auth, dto)

	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, uint(1), r.ID)
	repo.AssertCalled(t, "CreateOrUpdateRating", internal.Host, dto.TargetID, auth.CallerID, dto.Score, "great")
	repo.AssertCalled(t, "FindRatingByRater", internal.Host, dto.TargetID, auth.CallerID)
}

func Test_CreateRoomRating_Success(t *testing.T) {
	svc, repo, users, rooms, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 33, Score: 4, Comment: "nice"}

	// caller guest
	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	// room exists
	rooms.On("FindById", mock.Anything, dto.TargetID).
		Return(&roomclient.RoomDTO{ID: dto.TargetID, HostID: 2, Name: "r"}, nil)
	// eligibility ok
	resCli.On("CanUserRateRoom", mock.Anything, auth.CallerID, dto.TargetID).Return(true, nil)
	// upsert ok (updated=false case also fine)
	repo.On("CreateOrUpdateRating", internal.Room, dto.TargetID, auth.CallerID, dto.Score, "nice").Return(false, nil)
	// fetch stored
	repo.On("FindRatingByRater", internal.Room, dto.TargetID, auth.CallerID).
		Return(&internal.Rating{ID: 2, TargetType: internal.Room, TargetID: dto.TargetID, RaterID: auth.CallerID, Score: dto.Score, Comment: "nice"}, nil)

	r, err := svc.CreateRoomRating(context.Background(), auth, dto)

	assert.NoError(t, err)
	assert.NotNil(t, r)
	assert.Equal(t, uint(2), r.ID)
}

func Test_CreateRating_Unauthenticated_CallerLookupFails(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 5}

	users.On("FindById", mock.Anything, auth.CallerID).Return(nil, errors.New("user svc down"))

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err)
	assert.Nil(t, r)
}

func Test_CreateRating_Unauthorized_CallerNotGuest(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: NotGuest.Id, JWT: "jwt"} // role=host
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 5}

	users.On("FindById", mock.Anything, auth.CallerID).Return(NotGuest, nil)

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateRating_InvalidDTO_TargetMissing(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 0, Score: 5}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateRating_InvalidDTO_ScoreOutOfRange(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 6}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateHostRating_TargetHostNotFound(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 4}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	users.On("FindById", mock.Anything, dto.TargetID).Return(nil, errors.New("not found"))

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateHostRating_TargetNotHostRole(t *testing.T) {
	svc, _, users, _, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 4}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	users.On("FindById", mock.Anything, dto.TargetID).
		Return(&userclient.UserDTO{Id: dto.TargetID, Username: "x", Role: "guest"}, nil)

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateHostRating_EligibilityError(t *testing.T) {
	svc, _, users, _, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 4}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	users.On("FindById", mock.Anything, dto.TargetID).
		Return(&userclient.UserDTO{Id: dto.TargetID, Username: "host", Role: "host"}, nil)
	resCli.On("CanUserRateHost", mock.Anything, auth.CallerID, dto.TargetID).
		Return(false, errors.New("elig svc fail"))

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err)
	assert.Nil(t, r)
}

func Test_CreateHostRating_NotEligible(t *testing.T) {
	svc, _, users, _, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 22, Score: 4}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	users.On("FindById", mock.Anything, dto.TargetID).
		Return(&userclient.UserDTO{Id: dto.TargetID, Username: "host", Role: "host"}, nil)
	resCli.On("CanUserRateHost", mock.Anything, auth.CallerID, dto.TargetID).
		Return(false, nil)

	r, err := svc.CreateHostRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateRoomRating_RoomNotFound(t *testing.T) {
	svc, _, users, rooms, _ := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 33, Score: 4}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	rooms.On("FindById", mock.Anything, dto.TargetID).Return(nil, errors.New("no room"))

	r, err := svc.CreateRoomRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateRoomRating_EligibilityError(t *testing.T) {
	svc, _, users, rooms, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 33, Score: 3}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	rooms.On("FindById", mock.Anything, dto.TargetID).
		Return(&roomclient.RoomDTO{ID: dto.TargetID, HostID: 2, Name: "r"}, nil)
	resCli.On("CanUserRateRoom", mock.Anything, auth.CallerID, dto.TargetID).
		Return(false, errors.New("elig fail"))

	r, err := svc.CreateRoomRating(context.Background(), auth, dto)
	assert.Error(t, err)
	assert.Nil(t, r)
}

func Test_CreateRoomRating_NotEligible(t *testing.T) {
	svc, _, users, rooms, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 33, Score: 3}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	rooms.On("FindById", mock.Anything, dto.TargetID).
		Return(&roomclient.RoomDTO{ID: dto.TargetID, HostID: 2, Name: "r"}, nil)
	resCli.On("CanUserRateRoom", mock.Anything, auth.CallerID, dto.TargetID).
		Return(false, nil)

	r, err := svc.CreateRoomRating(context.Background(), auth, dto)
	assert.Error(t, err) 
	assert.Nil(t, r)
}

func Test_CreateRating_RepoUpsertError(t *testing.T) {
	svc, repo, users, rooms, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 33, Score: 5, Comment: "x"}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	rooms.On("FindById", mock.Anything, dto.TargetID).
		Return(&roomclient.RoomDTO{ID: dto.TargetID, HostID: 2, Name: "r"}, nil)
	resCli.On("CanUserRateRoom", mock.Anything, auth.CallerID, dto.TargetID).Return(true, nil)

	repo.On("CreateOrUpdateRating", internal.Room, dto.TargetID, auth.CallerID, dto.Score, "x").
		Return(false, errors.New("db upsert fail"))

	r, err := svc.CreateRoomRating(context.Background(), auth, dto)
	assert.Error(t, err)
	assert.Nil(t, r)
}

func Test_CreateRating_RepoFindAfterUpsertError(t *testing.T) {
	svc, repo, users, rooms, resCli := CreateTestRatingService()

	auth := internal.AuthContext{CallerID: DefaultGuest.Id, JWT: "jwt"}
	dto := internal.CreateRatingDTO{TargetID: 33, Score: 5, Comment: "x"}

	users.On("FindById", mock.Anything, auth.CallerID).Return(DefaultGuest, nil)
	rooms.On("FindById", mock.Anything, dto.TargetID).
		Return(&roomclient.RoomDTO{ID: dto.TargetID, HostID: 2, Name: "r"}, nil)
	resCli.On("CanUserRateRoom", mock.Anything, auth.CallerID, dto.TargetID).Return(true, nil)

	repo.On("CreateOrUpdateRating", internal.Room, dto.TargetID, auth.CallerID, dto.Score, "x").Return(true, nil)
	repo.On("FindRatingByRater", internal.Room, dto.TargetID, auth.CallerID).
		Return(nil, errors.New("fetch fail"))

	r, err := svc.CreateRoomRating(context.Background(), auth, dto)
	assert.Error(t, err)
	assert.Nil(t, r)
}