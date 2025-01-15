package db

import (
	"errors"
	"time"

	"gorm.io/gorm"

)

type AtcoderSubmissionStatus string

const (
	AtcoderSubmissionStatusAccepted     AtcoderSubmissionStatus = "AC"
	AtcoderSubmissionStatusWrongAnswer  AtcoderSubmissionStatus = "WA"
	AtcoderSubmissionStatusRuntimeError AtcoderSubmissionStatus = "RE"
)

type AtcoderUser struct {
	gorm.Model

	Handle           string `gorm:"uniqueIndex;index:idx_handle"`
	Avatar           string
	Rating           uint
	MaxRating        uint
	Level            string
	PromotionMessage string
	Submissions      []AtcoderSubmission
	// RatingChanges []AtcoderRatingChange
}

type AtcoderSubmission struct {
	gorm.Model

	AtcoderUserID    uint   `gorm:"index:idx_user_id"`                     // 单独索引用户ID
	AtcoderProblemID string `gorm:"index:idx_problem_id;type:varchar(64)"` // 单独索引问题ID

	SubmissionTime time.Time `gorm:"index:idx_user_id_at,idx_problem_id_at"` // 用户ID和时间的联合索引

	Status string
}

type AtcoderProblem struct {
	gorm.Model

	ID     string `gorm:"primaryKey;type:varchar(64)"`
	Rating uint

	Submissions []AtcoderSubmission
}

func MigrateAtcoder() error {
	return db.AutoMigrate(
		&AtcoderUser{},
		&AtcoderProblem{},
		&AtcoderSubmission{},
	)
}

func CountAtcoderSolvedByUID(uid uint) (uint, error) {
	var result int64
	if err := db.Model(&AtcoderSubmission{}).Where("atcoder_user_id = ? AND status = ?", uid, AtcoderSubmissionStatusAccepted).Distinct("atcoder_problem_id").Count(&result).Error; err != nil {
		return 0, err
	}
	return uint(result), nil
}

func LoadAtcoderUserByHandle(handle string) (*AtcoderUser, error) {
	var result AtcoderUser
	if err := db.Where("handle = ?", handle).First(&result).Error; err != nil {
		return nil, err
	}
	return &result, nil
}

func LoadAtcoderSolvedProblemByUID(UID uint) ([]AtcoderProblem, error) {
	var submissions []AtcoderSubmission
	if err := db.Where("atcoder_user_id = ?", UID).
		Where("status = ?", AtcoderSubmissionStatusAccepted).Find(&submissions).Error; err != nil {
		return nil, err
	}

	m := make(map[string]byte) // ProblemID Set

	for _, submission := range submissions {
		m[submission.AtcoderProblemID] = 0
	}

	problemIDs := make([]string, 0, len(m))
	for k := range m {
		problemIDs = append(problemIDs, k)
	}
	var problems []AtcoderProblem
	if err := db.Where("id IN ?", problemIDs).Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

func LoadLastAtcoderSubmissionByUID(UID uint) (*AtcoderSubmission, error) {
	var result AtcoderSubmission
	if err := db.Where("atcoder_user_id = ?", UID).Order("submission_time DESC").First(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &result, nil
}

func SaveAtcoderProblems(problems []AtcoderProblem) error {
	return saveLoop(problems)
}

func SaveAtcoderUser(user *AtcoderUser) error {
	return db.Save(user).Error
}

func SaveAtcoderSubmissions(submissions []AtcoderSubmission) error {
	return saveLoop(submissions)
}

func GetAtcoderUserID(handle string) (uint, error) {
	var user AtcoderUser
	if err := db.Model(&AtcoderUser{}).Where("handle = ?", handle).Select("id").First(&user).Error; err != nil {
		return 0, err
	}
	return user.ID, nil
}
