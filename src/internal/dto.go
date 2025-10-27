package internal

import "time"

type CreateRatingDTO struct {
	TargetID    	uint         `json:"targetId"`
	Score       	int          `json:"score"`
	Comment     	string       `json:"comment"`
}

type RatingDTO struct {
	Username string    `json:"username"`
	Score    int       `json:"score"`
	Comment  string    `json:"comment"`
	Time     time.Time `json:"time"` 
}

type RatingsWithAverageDTO struct {
	Average float64         `json:"average"`
	Ratings []RatingDTO 	`json:"ratings"`
}
