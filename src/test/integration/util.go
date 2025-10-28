package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"bookem-rating-service/client/roomclient"
	"bookem-rating-service/client/userclient"
	"bookem-rating-service/util"
)

// ---- Base URLs (compose service names) ----
const URL_user = "http://user-service:8080/api/"
const URL_room = "http://room-service:8080/api/"
const URL_reservation = "http://reservation-service:8080/api/"
const URL_rating = "http://rating-service:8080/api/"

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func GenName(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

// ---------- USERS ----------

func RegisterUser(username_or_email string, password string, role util.UserRole) (*http.Response, error) {
	username := username_or_email
	email := username + "@gmail.com"

	if strings.HasSuffix(username_or_email, "@gmail.com") {
		username = strings.Split(username_or_email, "@")[0]
		email = username_or_email
	}

	dto := userclient.UserCreateDTO{
		Username: username,
		Password: password,
		Email:    email,
		Role:     string(role),
		Name:     GenName(6),
		Surname:  GenName(6),
		Address:  GenName(10),
	}

	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	return http.Post(URL_user+"register", "application/json", bytes.NewBuffer(jsonBytes))
}

func LoginUser(username_or_email string, password string) (*http.Response, error) {
	dto := userclient.LoginDTO{
		UsernameOrEmail: username_or_email,
		Password:        password,
	}
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	return http.Post(URL_user+"login", "application/json", bytes.NewBuffer(jsonBytes))
}

func LoginUser2(username_or_email string, password string) string {
	resp, _ := LoginUser(username_or_email, password)
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}
	var token userclient.JWTDTO
	if err := json.Unmarshal(bodyBytes, &token); err != nil {
		panic(fmt.Sprintf("failed to unmarshal jwt: %v", err))
	}
	return token.Jwt
}

func MustGetUserIDFromJWT(jwt string) uint {
	j, err := util.GetJwtFromString(jwt)
	if err != nil {
		panic(err)
	}
	return j.ID
}

// ---------- ROOMS ----------

func CreateRoom(jwt string, dto roomclient.CreateRoomDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, URL_room+"new", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func ResponseToRoom(resp *http.Response) roomclient.RoomDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}
	var obj roomclient.RoomDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		fmt.Print(string(bodyBytes))
		panic(fmt.Sprintf("failed to unmarshal room: %v", err))
	}
	return obj
}

func CreateRoomAvailability(jwt string, dto roomclient.CreateRoomAvailabilityListDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, URL_room+"available", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func CreateRoomPrice(jwt string, dto roomclient.CreateRoomPriceListDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, URL_room+"price", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

const SMALL_IMG = "data:image/jpeg;base64,/9j/4AAQSkZJRgABAQEASABIAAD/2wBDAAMCAgMCAgMDAwMEAwMEBQgFBQQEBQoHBwYIDAoMDAsKCwsNDhIQDQ4RDgsLEBYQERMUFRUVDA8XGBYUGBIUFRT/wAALCAABAAEBAREA/8QAFAABAAAAAAAAAAAAAAAAAAAACf/EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAD8AKp//2Q=="

func SetupHostRoomAvailabilityPrice(hostUsername string) (string, uint, roomclient.RoomDTO) {
	// host
	RegisterUser(hostUsername, "pass", util.Host)
	hostJWT := LoginUser2(hostUsername, "pass")
	hostID := MustGetUserIDFromJWT(hostJWT)

	// room
	roomDTO := roomclient.CreateRoomDTO{
		HostID:        hostID,
		Name:          "Room_" + GenName(6),
		Description:   "Test room",
		Address:       "Test address",
		MinGuests:     1,
		MaxGuests:     4,
		PhotosPayload: []string{SMALL_IMG},
		Commodities:   []string{"WiFi"},
		AutoApprove:   true,
	}
	roomResp, err := CreateRoom(hostJWT, roomDTO)
	if err != nil {
		panic(err)
	}
	defer roomResp.Body.Close()
	room := ResponseToRoom(roomResp)

	avail := roomclient.CreateRoomAvailabilityListDTO{
		RoomID: room.ID,
		Items: []roomclient.CreateRoomAvailabilityItemDTO{
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
				Available:  true,
			},
		},
	}
	if resp, err := CreateRoomAvailability(hostJWT, avail); err != nil {
		panic(err)
	} else {
		resp.Body.Close()
	}

	price := roomclient.CreateRoomPriceListDTO{
		RoomID:    room.ID,
		BasePrice: 80,
		PerGuest:  false,
		Items: []roomclient.CreateRoomPriceItemDTO{
			{
				ExistingID: 0,
				DateFrom:   time.Date(2025, 9, 1, 0, 0, 0, 0, time.UTC),
				DateTo:     time.Date(2025, 9, 30, 0, 0, 0, 0, time.UTC),
				Price:      100,
			},
		},
	}
	if resp, err := CreateRoomPrice(hostJWT, price); err != nil {
		panic(err)
	} else {
		resp.Body.Close()
	}

	return hostJWT, hostID, room
}

// ---------- RESERVATIONS (for eligibility) ----------

type ReservationDTO struct {
	RoomID     uint      `json:"roomId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GuestCount int       `json:"guestCount"`
}

func CreateReservation(jwt string, dto ReservationDTO) (*http.Response, error) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, URL_reservation+"new", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+jwt)
	req.Header.Add("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func MakeGuestEligibleForRoom(guestJWT string, roomID uint) {
	dto := CreateReservationRequestDTO{
		RoomID:     roomID,
		DateFrom:   time.Date(2025, 9, 10, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 9, 12, 0, 0, 0, 0, time.UTC),
		GuestCount: 2,
	}
	resp, err := CreateReservationRequest(guestJWT, dto)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		panic(fmt.Sprintf("reservation request failed: status=%d body=%s", resp.StatusCode, string(body)))
	}
}


// ---------- RATING ----------

type CreateRatingDTO struct {
	TargetID uint   `json:"targetId,omitempty"`
	Score    int    `json:"score"`
	Comment  string `json:"comment,omitempty"`
}

type RatingDTO struct {
	ID         uint   `json:"id"`
	TargetType string `json:"targetType"` // "room" | "host"
	TargetID   uint   `json:"targetId"`
	RaterID    uint   `json:"raterId"`    // the guest (caller)
	Score      int    `json:"score"`
	Comment    string `json:"comment"`
}

func CreateHostRating(jwt string, hostID uint, body CreateRatingDTO) (*http.Response, error) {
	b, _ := json.Marshal(body)
	url := fmt.Sprintf(URL_rating+"ratings/%d?type=host", hostID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func CreateRoomRating(jwt string, roomID uint, body CreateRatingDTO) (*http.Response, error) {
	b, _ := json.Marshal(body)
	url := fmt.Sprintf(URL_rating+"ratings/%d?type=room", roomID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}


func ResponseToRating(resp *http.Response) RatingDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}
	var obj RatingDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		fmt.Print(string(bodyBytes))
		panic(fmt.Sprintf("failed to unmarshal rating: %v", err))
	}
	return obj
}

type CreateReservationRequestDTO struct {
	RoomID     uint      `json:"roomId"`
	DateFrom   time.Time `json:"dateFrom"`
	DateTo     time.Time `json:"dateTo"`
	GuestCount int       `json:"guestCount"`
}

func CreateReservationRequest(jwt string, dto CreateReservationRequestDTO) (*http.Response, error) {
	b, _ := json.Marshal(dto)
	req, err := http.NewRequest(http.MethodPost, URL_reservation+"req", bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+jwt)
	req.Header.Set("Content-Type", "application/json")
	return http.DefaultClient.Do(req)
}

func DeleteHostRating(jwt string, hostID uint) (*http.Response, error) {
	url := fmt.Sprintf(URL_rating+"ratings/%d?type=host", hostID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil { return nil, err }
	req.Header.Set("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

func DeleteRoomRating(jwt string, roomID uint) (*http.Response, error) {
	url := fmt.Sprintf(URL_rating+"ratings/%d?type=room", roomID)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil { return nil, err }
	req.Header.Set("Authorization", "Bearer "+jwt)
	return http.DefaultClient.Do(req)
}

type PublicRatingDTO struct {
	Username string    `json:"username"`
	Score    int       `json:"score"`
	Comment  string    `json:"comment"`
	Time     time.Time `json:"time"`
}

type RatingsWithAverageListDTO struct {
	Average float64           `json:"average"`
	Ratings []PublicRatingDTO `json:"ratings"`
}

func GetRatingsWithAvg(rt string, id uint) (*http.Response, error) {
	url := fmt.Sprintf(URL_rating+"ratings/all/%d?type=%s", id, rt)
	return http.Get(url)
}

func ResponseToRatingsWithAvg(resp *http.Response) RatingsWithAverageListDTO {
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Sprintf("failed to read response body: %v", err))
	}
	var obj RatingsWithAverageListDTO
	if err := json.Unmarshal(bodyBytes, &obj); err != nil {
		fmt.Print(string(bodyBytes))
		panic(fmt.Sprintf("failed to unmarshal ratings-with-avg: %v", err))
	}
	return obj
}
