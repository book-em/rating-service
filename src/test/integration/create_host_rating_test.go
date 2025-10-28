package test

import (
	"net/http"
	"testing"

	"bookem-rating-service/util"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateHostRating_HappyPath_WithEligibility(t *testing.T) {
	_, hostID, room := SetupHostRoomAvailabilityPrice("host_rate_hp_001")

	RegisterUser("guest_rate_hp_001", "pass", util.Guest)
	guestJWT := LoginUser2("guest_rate_hp_001", "pass")
	MakeGuestEligibleForRoom(guestJWT, room.ID)

	// Guest rates the host
	body := CreateRatingDTO{
		Score:   5,
		Comment: "Great host — smooth stay.",
	}
	resp, err := CreateHostRating(guestJWT, hostID, body)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	r := ResponseToRating(resp)
	assert.Equal(t, hostID, r.TargetID)
	assert.Equal(t, 5, r.Score)
	assert.Equal(t, "host", r.TargetType)
	assert.Equal(t, "Great host — smooth stay.", r.Comment)
	assert.NotZero(t, r.ID)
	assert.NotZero(t, r.RaterID)        

}
