package test

import (
	"net/http"
	"testing"

	"bookem-rating-service/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateRoomRating_HappyPath_WithEligibility(t *testing.T) {
	_, _, room := SetupHostRoomAvailabilityPrice("host_rate_hp_002")

	RegisterUser("guest_rate_hp_002", "pass", util.Guest)
	guestJWT := LoginUser2("guest_rate_hp_002", "pass")
	MakeGuestEligibleForRoom(guestJWT, room.ID)

	body := CreateRatingDTO{
		Score:   4,
		Comment: "Clean, quiet, would book again.",
	}
	resp, err := CreateRoomRating(guestJWT, room.ID, body)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	r := ResponseToRating(resp)
	assert.Equal(t, room.ID, r.TargetID)
	assert.Equal(t, 4, r.Score)
	assert.Equal(t, "room", r.TargetType)
	assert.Equal(t, "Clean, quiet, would book again.", r.Comment)
	assert.NotZero(t, r.ID)
	assert.NotZero(t, r.RaterID)     

}
