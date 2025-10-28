package test

import (
	"bookem-rating-service/client/roomclient"
	"bookem-rating-service/client/userclient"
	"bookem-rating-service/internal"
	"context"
	"time"

	mock "github.com/stretchr/testify/mock"
)

func CreateTestRatingService() (
	internal.Service,
	*MockRatingRepo,
	*MockUserClient,
	*MockRoomClient,
	*MockReservationClient,
) {
	mockRepo := new(MockRatingRepo)
	mockUserClient := new(MockUserClient)
	mockRoomClient := new(MockRoomClient)
	mockReservationClient := new(MockReservationClient)

	svc := internal.NewService(mockRepo, mockUserClient, mockRoomClient, mockReservationClient)
	return svc, mockRepo, mockUserClient, mockRoomClient, mockReservationClient
}

// ------------------------------- Mock repo

type MockRatingRepo struct {
	mock.Mock
}

func (m *MockRatingRepo) CreateOrUpdateRating(rt internal.RatingType, targetID, raterID uint, score int, comment string) (bool, error) {
	args := m.Called(rt, targetID, raterID, score, comment)
	return args.Bool(0), args.Error(1)
}

func (m *MockRatingRepo) DeleteRating(rt internal.RatingType, targetID uint, raterID uint) error {
	args := m.Called(rt, targetID, raterID)
	return args.Error(0)
}

func (m *MockRatingRepo) GetRatingsWithAverage(rt internal.RatingType, targetID uint) (*internal.RatingsWithAverageDTO, error) {
	args := m.Called(rt, targetID)
	var res *internal.RatingsWithAverageDTO
	if v := args.Get(0); v != nil {
		res = v.(*internal.RatingsWithAverageDTO)
	}
	return res, args.Error(1)
}

func (m *MockRatingRepo) FindAllRatings(rt internal.RatingType, targetID uint) ([]internal.Rating, error) {
	args := m.Called(rt, targetID)
	if v := args.Get(0); v != nil {
		return v.([]internal.Rating), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRatingRepo) GetAverageRating(rt internal.RatingType, targetID uint) (float64, error) {
	args := m.Called(rt, targetID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockRatingRepo) FindRatingByRater(rt internal.RatingType, targetID uint, raterID uint) (*internal.Rating, error) {
	args := m.Called(rt, targetID, raterID)
	if v := args.Get(0); v != nil {
		return v.(*internal.Rating), args.Error(1)
	}
	return nil, args.Error(1)
}

// ------------------------------- Mock user client

type MockUserClient struct {
	mock.Mock
}

func (m *MockUserClient) FindById(ctx context.Context, id uint) (*userclient.UserDTO, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*userclient.UserDTO)
	return user, args.Error(1)
}

// ------------------------------- Mock room client (not used in delete but required in ctor)

type MockRoomClient struct {
	mock.Mock
}

func (r *MockRoomClient) FindById(context context.Context, id uint) (*roomclient.RoomDTO, error) {
	args := r.Called(context, id)
	room, _ := args.Get(0).(*roomclient.RoomDTO)
	return room, args.Error(1)
}

func (r *MockRoomClient) FindCurrentAvailabilityListOfRoom(context context.Context, roomId uint) (*roomclient.RoomAvailabilityListDTO, error) {
	args := r.Called(context, roomId)
	list, _ := args.Get(0).(*roomclient.RoomAvailabilityListDTO)
	return list, args.Error(1)
}

func (r *MockRoomClient) FindCurrentPricelistOfRoom(context context.Context, roomId uint) (*roomclient.RoomPriceListDTO, error) {
	args := r.Called(context, roomId)
	list, _ := args.Get(0).(*roomclient.RoomPriceListDTO)
	return list, args.Error(1)
}

func (r *MockRoomClient) QueryForReservation(context context.Context, jwt string, dto roomclient.RoomReservationQueryDTO) (*roomclient.RoomReservationQueryResponseDTO, error) {
	args := r.Called(context, jwt, dto)
	resp, _ := args.Get(0).(*roomclient.RoomReservationQueryResponseDTO)
	return resp, args.Error(1)
}

func (r *MockRoomClient) FindByHostId(context context.Context, id uint) ([]roomclient.RoomDTO, error) {
	args := r.Called(context, id)
	room, _ := args.Get(0).([]roomclient.RoomDTO)
	return room, args.Error(1)
}



// ------------------------------- Mock reservation client (not used in delete but required)

type MockReservationClient struct {
	mock.Mock
}

func (m *MockReservationClient) CanUserRateHost(ctx context.Context, guestID, hostID uint) (bool, error) {
	args := m.Called(ctx, guestID, hostID)
	return args.Bool(0), args.Error(1)
}

func (m *MockReservationClient) CanUserRateRoom(ctx context.Context, guestID, roomID uint) (bool, error) {
	args := m.Called(ctx, guestID, roomID)
	return args.Bool(0), args.Error(1)
}

// ------------------------------- Common mock data

var DefaultGuest = &userclient.UserDTO{
	Id:       10,
	Username: "guest1",
	Email:    "g1@mail.com",
	Name:     "G",
	Surname:  "U",
	Role:     "guest",
	Address:  "gAddress 1",
}

var NotGuest = &userclient.UserDTO{
	Id:       11,
	Username: "host1",
	Email:    "h1@mail.com",
	Name:     "H",
	Surname:  "O",
	Role:     "host",
	Address:  "hAddress 2",
}

var SampleRatings = []internal.Rating{
	{ID: 1, TargetType: internal.Host, TargetID: 22, RaterID: 10, Score: 5, Comment: "great",  CreatedAt: time.Now().Add(-48 * time.Hour)},
	{ID: 2, TargetType: internal.Host, TargetID: 22, RaterID: 99, Score: 3, Comment: "ok",     UpdatedAt: time.Now().Add(-24 * time.Hour)},
}

