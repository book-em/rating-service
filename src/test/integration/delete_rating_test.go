package test

import (
	"net/http"
	"testing"

	"bookem-rating-service/util"
	"github.com/stretchr/testify/require"
)

func TestDeleteHostRating_HappyPath(t *testing.T) {
	_, hostID, room := SetupHostRoomAvailabilityPrice("host_del_hp_001")

	RegisterUser("guest_del_hp_001", "pass", util.Guest)
	guestJWT := LoginUser2("guest_del_hp_001", "pass")
	MakeGuestEligibleForRoom(guestJWT, room.ID)

	createResp, err := CreateHostRating(guestJWT, hostID, CreateRatingDTO{
		Score:   5,
		Comment: "Host was great.",
	})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	delResp, err := DeleteHostRating(guestJWT, hostID)
	require.NoError(t, err)
	defer delResp.Body.Close()
	require.Equal(t, http.StatusNoContent, delResp.StatusCode)
}

func TestDeleteRoomRating_HappyPath(t *testing.T) {
	_, _, room := SetupHostRoomAvailabilityPrice("host_del_hp_002")

	RegisterUser("guest_del_hp_002", "pass", util.Guest)
	guestJWT := LoginUser2("guest_del_hp_002", "pass")
	MakeGuestEligibleForRoom(guestJWT, room.ID)

	createResp, err := CreateRoomRating(guestJWT, room.ID, CreateRatingDTO{
		Score:   5,
		Comment: "Room was excellent.",
	})
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	delResp, err := DeleteRoomRating(guestJWT, room.ID)
	require.NoError(t, err)
	defer delResp.Body.Close()
	require.Equal(t, http.StatusNoContent, delResp.StatusCode)

}
