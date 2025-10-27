package internal

import (
	"bookem-rating-service/client/roomclient"
	"bookem-rating-service/client/userclient"
)

// AuthContext is used in cases where callerID is not enoug
type AuthContext struct {
	CallerID uint
	JWT      string
}

type Service interface {

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
