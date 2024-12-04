package db

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type CodeforcesSubmissionStatus string

const (
	CodeforcesSubmissionStatusOk CodeforcesSubmissionStatus = "OK"

	//CodeforcesSubmissionStatusFailed                  CodeforcesSubmissionStatus = "FAILED"
	//CodeforcesSubmissionStatusPartial                 CodeforcesSubmissionStatus = "PARTIAL"
	//CodeforcesSubmissionStatusCompilationError        CodeforcesSubmissionStatus = "COMPILATION_ERROR"
	//CodeforcesSubmissionStatusRuntimeError            CodeforcesSubmissionStatus = "RUNTIME_ERROR"
	//CodeforcesSubmissionStatusWrongAnswer             CodeforcesSubmissionStatus = "WRONG_ANSWER"
	//CodeforcesSubmissionStatusPresentationError       CodeforcesSubmissionStatus = "PRESENTATION_ERROR"
	//CodeforcesSubmissionStatusTimeLimitExceeded       CodeforcesSubmissionStatus = "TIME_LIMIT_EXCEEDED"
	//CodeforcesSubmissionStatusMemoryLimitExceeded     CodeforcesSubmissionStatus = "MEMORY_LIMIT_EXCEEDED"
	//CodeforcesSubmissionStatusIdlenessLimitExceeded   CodeforcesSubmissionStatus = "IDLENESS_LIMIT_EXCEEDED"
	//CodeforcesSubmissionStatusSecurityViolated        CodeforcesSubmissionStatus = "SECURITY_VIOLATED"
	//CodeforcesSubmissionStatusCrashed                 CodeforcesSubmissionStatus = "CRASHED"
	//CodeforcesSubmissionStatusInputPreparationCrashed CodeforcesSubmissionStatus = "INPUT_PREPARATION_CRASHED"
	//CodeforcesSubmissionStatusChallenged              CodeforcesSubmissionStatus = "CHALLENGED"
	//CodeforcesSubmissionStatusSkipped                 CodeforcesSubmissionStatus = "SKIPPED"
	//CodeforcesSubmissionStatusTesting                 CodeforcesSubmissionStatus = "TESTING"
	//CodeforcesSubmissionStatusRejected                CodeforcesSubmissionStatus = "REJECTED"
)

type CodeforcesUser struct {
	gorm.Model

	Handle      string `gorm:"uniqueIndex;index:idx_handle"`
	Avatar      string
	Rating      int
	MaxRating   int
	FriendCount int

	Submissions   []CodeforcesSubmission
	RatingChanges []CodeforcesRatingChange
}

type CodeforcesSubmission struct {
	gorm.Model

	CodeforcesUserID    uint   `gorm:"index:idx_user_id"`                     // 单独索引用户ID
	CodeforcesProblemID string `gorm:"index:idx_problem_id;type:varchar(64)"` // 单独索引问题ID

	At time.Time `gorm:"index:idx_user_id_at,idx_problem_id_at"` // 用户ID和时间的联合索引

	Status string
}

type CodeforcesProblem struct {
	gorm.Model

	ID     string `gorm:"primaryKey;type:varchar(64)"`
	Rating int

	Submissions []CodeforcesSubmission
}

type CodeforcesRatingChange struct {
	gorm.Model

	CodeforcesUserID uint `gorm:"index:idx_codeforces_user_id"`

	At        time.Time `gorm:"index:idx_codeforces_user_id_at"`
	NewRating int
}

func MigrateCodeforces() error {
	return db.AutoMigrate(
		&CodeforcesUser{},
		&CodeforcesProblem{},
		&CodeforcesSubmission{},
		&CodeforcesRatingChange{},
	)
}

func CountCodeforcesSolvedByUID(uid uint) (int, error) {
	result := 0
	if err := db.Raw(`
		SELECT COUNT(DISTINCT codeforces_problem_id) 
		FROM codeforces_submissions 
		WHERE codeforces_user_id = ? AND status = ?`,
		uid, CodeforcesSubmissionStatusOk).Scan(&result).Error; err != nil {
		return 0, err
	}
	return result, nil
}

func LoadCodeforcesUserByHandle(handle string) (*CodeforcesUser, error) {
	result := &CodeforcesUser{}
	if err := db.Where("handle = ?", handle).First(result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func LoadCodeforcesSolvedProblemByUID(UID uint) ([]CodeforcesProblem, error) {
	var submissions []CodeforcesSubmission
	if err := db.Where("codeforces_user_id = ?", UID).
		Where("status = ?", CodeforcesSubmissionStatusOk).Find(&submissions).Error; err != nil {
		return nil, err
	}

	m := make(map[string]byte) // ProblemID Set

	for _, submission := range submissions {
		m[submission.CodeforcesProblemID] = 0
	}

	problemIDs := make([]string, 0, len(m))
	for k := range m {
		problemIDs = append(problemIDs, k)
	}
	var problems []CodeforcesProblem
	if err := db.Where("id IN ?", problemIDs).Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

func LoadLastCodeforcesSubmissionByUID(UID uint) (*CodeforcesSubmission, error) {
	result := &CodeforcesSubmission{}
	if err := db.Where("codeforces_user_id = ?", UID).Order("at DESC").First(result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func LoadLastCodeforcesRatingChangeByUID(UID uint) (*CodeforcesRatingChange, error) {
	result := &CodeforcesRatingChange{}
	if err := db.Where("codeforces_user_id = ?", UID).Order("at DESC").First(result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return result, nil
}

func LoadCodeforcesRatingChangesByUID(UID uint) ([]CodeforcesRatingChange, error) {
	result := make([]CodeforcesRatingChange, 0)
	if err := db.Where("codeforces_user_id = ?", UID).Find(&result).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return result, nil
		}
		return nil, err
	}
	return result, nil
}

func SaveCodeforcesProblems(problems []CodeforcesProblem) error {
	return saveLoop[CodeforcesProblem](problems)
}

func SaveCodeforcesUser(user *CodeforcesUser) error {
	return db.Save(user).Error
}

func SaveCodeforcesSubmissions(submissions []CodeforcesSubmission) error {
	return saveLoop[CodeforcesSubmission](submissions)
}

func SaveCodeforcesRatingChanges(changes []CodeforcesRatingChange) error {
	return saveLoop[CodeforcesRatingChange](changes)
}

func saveLoop[T any](data []T) error {
	for i := 0; i < len(data); i += signalInsertLimit {
		if err := db.Save(data[i:min(len(data), i+signalInsertLimit)]).Error; err != nil {
			return err
		}
	}
	return nil
}
