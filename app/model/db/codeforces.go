package db

import (
	"gorm.io/gorm"
	"time"
)

type CodeforcesSubmissionStatus string

const (
	CodeforcesSubmissionStatusFailed                  CodeforcesSubmissionStatus = "FAILED"
	CodeforcesSubmissionStatusOk                      CodeforcesSubmissionStatus = "OK"
	CodeforcesSubmissionStatusPartial                 CodeforcesSubmissionStatus = "PARTIAL"
	CodeforcesSubmissionStatusCompilationError        CodeforcesSubmissionStatus = "COMPILATION_ERROR"
	CodeforcesSubmissionStatusRuntimeError            CodeforcesSubmissionStatus = "RUNTIME_ERROR"
	CodeforcesSubmissionStatusWrongAnswer             CodeforcesSubmissionStatus = "WRONG_ANSWER"
	CodeforcesSubmissionStatusPresentationError       CodeforcesSubmissionStatus = "PRESENTATION_ERROR"
	CodeforcesSubmissionStatusTimeLimitExceeded       CodeforcesSubmissionStatus = "TIME_LIMIT_EXCEEDED"
	CodeforcesSubmissionStatusMemoryLimitExceeded     CodeforcesSubmissionStatus = "MEMORY_LIMIT_EXCEEDED"
	CodeforcesSubmissionStatusIdlenessLimitExceeded   CodeforcesSubmissionStatus = "IDLENESS_LIMIT_EXCEEDED"
	CodeforcesSubmissionStatusSecurityViolated        CodeforcesSubmissionStatus = "SECURITY_VIOLATED"
	CodeforcesSubmissionStatusCrashed                 CodeforcesSubmissionStatus = "CRASHED"
	CodeforcesSubmissionStatusInputPreparationCrashed CodeforcesSubmissionStatus = "INPUT_PREPARATION_CRASHED"
	CodeforcesSubmissionStatusChallenged              CodeforcesSubmissionStatus = "CHALLENGED"
	CodeforcesSubmissionStatusSkipped                 CodeforcesSubmissionStatus = "SKIPPED"
	CodeforcesSubmissionStatusTesting                 CodeforcesSubmissionStatus = "TESTING"
	CodeforcesSubmissionStatusRejected                CodeforcesSubmissionStatus = "REJECTED"
)

type CodeforcesUser struct {
	gorm.Model

	Handle      string `gorm:"uniqueIndex;index:idx_handle"`
	Avatar      string
	Rating      uint
	FriendCount uint
	Solved      uint

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

func MigrateCodeforces(db *gorm.DB) error {
	return db.AutoMigrate(
		&CodeforcesUser{},
		&CodeforcesProblem{},
		&CodeforcesSubmission{},
		&CodeforcesRatingChange{},
	)
}
