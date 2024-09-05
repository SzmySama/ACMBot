package db

import (
	"errors"
	"time"

	"github.com/SzmySama/ACMBot/app/types"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func SaveUser(u *types.User) error {
	return GetDBConnection().Save(u).Error
}

func InsertSubmissions(handle string, submission2insert []types.Submission) error {
	// Submission 是倒着排的
	if len(submission2insert) < 1 {
		return nil
	}

	db := GetDBConnection()
	var lastSubmission types.Submission
	var lastSubmissionAt time.Time
	if err := db.Where("user_handle = ?", handle).Order("at Desc").First(&lastSubmission).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		lastSubmissionAt = time.Unix(0, 0)
	} else {
		lastSubmissionAt = lastSubmission.At
	}

	for _, thisSubmission := range submission2insert {
		if thisSubmission.At.Unix() <= lastSubmissionAt.Unix() {
			break
		}

		if err := db.Save(&thisSubmission).Error; err != nil {
			log.Errorf("Failed to save submission for user %s: %v", handle, err)
			return err
		}
	}
	return nil
}

func InsertRatingChanges(handle string, ratingChanges2insert []types.RatingChange) error {
	// Rating Change 是顺着排序的
	if len(ratingChanges2insert) < 1 {
		return nil
	}

	db := GetDBConnection()
	var lastRatingChangeAt time.Time
	var lastRatingChange types.RatingChange
	if err := db.Where("user_handle = ?", handle).Order("at Desc").First(&lastRatingChange).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		lastRatingChangeAt = time.Unix(0, 0)
	} else {
		lastRatingChangeAt = lastRatingChange.At
	}

	for i := len(ratingChanges2insert) - 1; i >= 0; i-- {
		thisRatingChange := ratingChanges2insert[i]
		if thisRatingChange.At.Unix() <= lastRatingChangeAt.Unix() {
			break
		}

		if err := db.Save(&thisRatingChange).Error; err != nil {
			log.Errorf("Failed to save rating change for user %s: %v", handle, err)
			return err
		}
	}

	return nil
}

func SaveProblem(problem *types.Problem) error {
	return GetDBConnection().Save(problem).Error
}

func CountUserSolved(handle string) (result int64, err error) {
	if err = GetDBConnection().Model(&types.Submission{}).Where("handle = ? AND status = ?", handle, types.SUBMISSION_STATUS_OK).Count(&result).Error; err != nil {
		result = 0
		return
	}
	return
}
