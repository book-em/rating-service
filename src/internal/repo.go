package internal

import (
	"gorm.io/gorm"
)

type Repository interface {
	CreateOrUpdateRating(rt RatingType, targetID uint, raterID uint, score int, comment string) (bool, error)
	FindRatingByRater(rt RatingType, targetID uint, raterID uint) (*Rating, error)
	DeleteRating(rt RatingType, targetID uint, raterID uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db}
}

func (r *repository) CreateOrUpdateRating(rt RatingType, targetID uint, raterID uint, score int, comment string) (bool, error) {
	var existing Rating
	tx := r.db.Where("target_type = ? AND target_id = ? AND rater_id = ?", rt, targetID, raterID).First(&existing)
	if tx.Error == gorm.ErrRecordNotFound {
		rating := Rating{
			TargetType: rt,
			TargetID:   targetID,
			RaterID:    raterID,
			Score:      score,
			Comment:    comment,
		}
		return true, r.db.Create(&rating).Error
	}
	if tx.Error != nil {
		return false, tx.Error
	}
	existing.Score = score
	existing.Comment = comment
	return false, r.db.Save(&existing).Error
}

func (r *repository) DeleteRating(rt RatingType, targetID uint, raterID uint) error {
	return r.db.Where("target_type = ? AND target_id = ? AND rater_id = ?", rt, targetID, raterID).
		Delete(&Rating{}).Error
}

func (r *repository) FindRatingByRater(rt RatingType, targetID uint, raterID uint) (*Rating, error) {
    var rating Rating
    if err := r.db.Where("target_type = ? AND target_id = ? AND rater_id = ?", rt, targetID, raterID).
        First(&rating).Error; err != nil {
        return nil, err
    }
    return &rating, nil
}
