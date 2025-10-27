package internal

type CreateRatingDTO struct {
	TargetID    	uint         `json:"targetId"`
	Score       	int          `json:"score"`
	Comment     	string       `json:"comment"`
}