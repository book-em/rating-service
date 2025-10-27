package internal

import "time"

type RatingType string

const (
	Host  RatingType = "host"
	Room RatingType = "room"
)

type Rating struct {
	ID          	uint           `gorm:"primaryKey"`
	TargetType  	RatingType     `gorm:"index:idx_target_rater,priority:1;index"`
	TargetID    	uint		   `gorm:"index:idx_target_rater,priority:2;index"`
	RaterID			uint           `gorm:"index:idx_target_rater,priority:3;index"`
	Score       	int            `gorm:"not null;check:score>=1 AND score<=5"`
	Comment     	string         `gorm:"type:text"`
	CreatedAt   	time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   	time.Time      `gorm:"autoUpdateTime"`
}