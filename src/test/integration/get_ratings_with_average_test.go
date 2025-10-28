package test

import (
	"net/http"
	"testing"
	"time"

	"bookem-rating-service/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

)

func TestGetRoomRatingsWithAvg_HappyPath(t *testing.T) {
	_, _, room := SetupHostRoomAvailabilityPrice("host_avg_room_001")

	RegisterUser("guest_avg_room_A", "pass", util.Guest)
	jwtA := LoginUser2("guest_avg_room_A", "pass")
	_, _ = CreateReservationRequest(jwtA, CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 12, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	})
	respA, err := CreateRoomRating(jwtA, room.ID, CreateRatingDTO{Score: 4, Comment: "good"})
	require.NoError(t, err)
	defer respA.Body.Close()
	require.Equal(t, http.StatusCreated, respA.StatusCode)

	RegisterUser("guest_avg_room_B", "pass", util.Guest)
	jwtB := LoginUser2("guest_avg_room_B", "pass")
	_, _ = CreateReservationRequest(jwtB, CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 13, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	})
	respB, err := CreateRoomRating(jwtB, room.ID, CreateRatingDTO{Score: 5, Comment: "great"})
	require.NoError(t, err)
	defer respB.Body.Close()
	require.Equal(t, http.StatusCreated, respB.StatusCode)

	getResp, err := GetRatingsWithAvg("room", room.ID)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	payload := ResponseToRatingsWithAvg(getResp)

	assert.Len(t, payload.Ratings, 2)
	assert.InDelta(t, 4.5, payload.Average, 0.001)

	found4, found5 := false, false
	for _, r := range payload.Ratings {
		assert.NotEmpty(t, r.Username)
		assert.True(t, r.Time.IsZero() == false)
		if r.Score == 4 && r.Comment == "good" {
			found4 = true
		}
		if r.Score == 5 && r.Comment == "great" {
			found5 = true
		}
	}
	assert.True(t, found4, "expected 4-star rating present")
	assert.True(t, found5, "expected 5-star rating present")
}

func TestGetHostRatingsWithAvg_HappyPath(t *testing.T) {
	_, hostID, room := SetupHostRoomAvailabilityPrice("host_avg_host_001")

	RegisterUser("guest_avg_host_A", "pass", util.Guest)
	jwtA := LoginUser2("guest_avg_host_A", "pass")
	_, _ = CreateReservationRequest(jwtA, CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 12, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	})
	respA, err := CreateHostRating(jwtA, hostID, CreateRatingDTO{Score: 4, Comment: "good host"})
	require.NoError(t, err)
	defer respA.Body.Close()
	require.Equal(t, http.StatusCreated, respA.StatusCode)

	RegisterUser("guest_avg_host_B", "pass", util.Guest)
	jwtB := LoginUser2("guest_avg_host_B", "pass")
	_, _ = CreateReservationRequest(jwtB, CreateReservationRequestDTO{
		RoomID:     room.ID,
		DateFrom:   time.Date(2025, 9, 13, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 15, 0, 0, 0, 0, time.UTC),
		GuestCount: 1,
	})
	respB, err := CreateHostRating(jwtB, hostID, CreateRatingDTO{Score: 5, Comment: "excellent host"})
	require.NoError(t, err)
	defer respB.Body.Close()
	require.Equal(t, http.StatusCreated, respB.StatusCode)

	getResp, err := GetRatingsWithAvg("host", hostID)
	require.NoError(t, err)
	defer getResp.Body.Close()
	require.Equal(t, http.StatusOK, getResp.StatusCode)

	payload := ResponseToRatingsWithAvg(getResp)

	assert.Len(t, payload.Ratings, 2)
	assert.InDelta(t, 4.5, payload.Average, 0.001)

	found4, found5 := false, false
	for _, r := range payload.Ratings {
		assert.NotEmpty(t, r.Username)
		assert.True(t, r.Time.IsZero() == false)
		if r.Score == 4 && r.Comment == "good host" {
			found4 = true
		}
		if r.Score == 5 && r.Comment == "excellent host" {
			found5 = true
		}
	}
	assert.True(t, found4, "expected 4-star host rating present")
	assert.True(t, found5, "expected 5-star host rating present")
}
