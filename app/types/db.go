package types

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type SubmissionStatus string
type ProblemType string

const (
	SUBMISSION_STATUS_FAILED                    SubmissionStatus = "FAILED"
	SUBMISSION_STATUS_OK                        SubmissionStatus = "OK"
	SUBMISSION_STATUS_PARTIAL                   SubmissionStatus = "PARTIAL"
	SUBMISSION_STATUS_COMPILATION_ERROR         SubmissionStatus = "COMPILATION_ERROR"
	SUBMISSION_STATUS_RUNTIME_ERROR             SubmissionStatus = "RUNTIME_ERROR"
	SUBMISSION_STATUS_WRONG_ANSWER              SubmissionStatus = "WRONG_ANSWER"
	SUBMISSION_STATUS_PRESENTATION_ERROR        SubmissionStatus = "PRESENTATION_ERROR"
	SUBMISSION_STATUS_TIME_LIMIT_EXCEEDED       SubmissionStatus = "TIME_LIMIT_EXCEEDED"
	SUBMISSION_STATUS_MEMORY_LIMIT_EXCEEDED     SubmissionStatus = "MEMORY_LIMIT_EXCEEDED"
	SUBMISSION_STATUS_IDLENESS_LIMIT_EXCEEDED   SubmissionStatus = "IDLENESS_LIMIT_EXCEEDED"
	SUBMISSION_STATUS_SECURITY_VIOLATED         SubmissionStatus = "SECURITY_VIOLATED"
	SUBMISSION_STATUS_CRASHED                   SubmissionStatus = "CRASHED"
	SUBMISSION_STATUS_INPUT_PREPARATION_CRASHED SubmissionStatus = "INPUT_PREPARATION_CRASHED"
	SUBMISSION_STATUS_CHALLENGED                SubmissionStatus = "CHALLENGED"
	SUBMISSION_STATUS_SKIPPED                   SubmissionStatus = "SKIPPED"
	SUBMISSION_STATUS_TESTING                   SubmissionStatus = "TESTING"
	SUBMISSION_STATUS_REJECTED                  SubmissionStatus = "REJECTED"
)

const (
	Contest    ProblemType = "contest"
	ProblemSet ProblemType = "problemSet"
	Unknow     ProblemType = "unknow"
)

const (
	Failed                  SubmissionStatus = "FAILED"
	OK                      SubmissionStatus = "OK"
	Partial                 SubmissionStatus = "PARTIAL"
	CompilationError        SubmissionStatus = "COMPILATION_ERROR"
	RuntimeError            SubmissionStatus = "RUNTIME_ERROR"
	WrongAnswer             SubmissionStatus = "WRONG_ANSWER"
	PresentationError       SubmissionStatus = "PRESENTATION_ERROR"
	TimeLimitExceeded       SubmissionStatus = "TIME_LIMIT_EXCEEDED"
	MemoryLimitExceeded     SubmissionStatus = "MEMORY_LIMIT_EXCEEDED"
	IdlenessLimitExceeded   SubmissionStatus = "IDLENESS_LIMIT_EXCEEDED"
	SecurityViolated        SubmissionStatus = "SECURITY_VIOLATED"
	Crashed                 SubmissionStatus = "CRASHED"
	InputPreparationCrashed SubmissionStatus = "INPUT_PREPARATION_CRASHED"
	Challenged              SubmissionStatus = "CHALLENGED"
	Skipped                 SubmissionStatus = "SKIPPED"
	Testing                 SubmissionStatus = "TESTING"
	Rejected                SubmissionStatus = "REJECTED"
)

type User struct {
	Handle    string    `gorm:"primaryKey;not null;type:varchar(255)" json:"handle"`
	Rating    int       `json:"rating"`
	Avatar    string    `json:"avatar"`
	CreatedAt time.Time `json:"-"`
	FriendOf  int       `json:"friendOfCount"`
	UpdatedAt time.Time
	Solved    int

	Submissions   []Submission `gorm:"many2many:user_submissions"`
	RatingChanges []RatingChange
}

type Problem struct {
	ID             string `gorm:"primaryKey;type:varchar(255)"`
	ContestID      *string
	ProblemsetName *string
	Index          string `gorm:"not null"`
	Rating         int

	Submissions []Submission
}

type Submission struct {
	ID        uint             `gorm:"primaryKey"`
	Status    SubmissionStatus `gorm:"not null;type:varchar(20)"`
	At        time.Time        `gorm:"not null"`
	ProblemID string           `gorm:"not null"`

	Problem Problem
	Users   []User `gorm:"many2many:user_submissions"`
}

type RatingChange struct {
	UserHandle string    `gorm:"index:idx_handle_at;not null;type:varchar(255)"`
	At         time.Time `gorm:"index:idx_handle_at;primaryKey;not null"`
	NewRating  int       `gorm:"not null"`

	User User
}

func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User
	aux := &struct {
		RegistrationTimeSeconds int64 `json:"registrationTimeSeconds"`
		*Alias
	}{
		Alias: (*Alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.CreatedAt = time.Unix(aux.RegistrationTimeSeconds, 0)
	return nil
}

func (p *Problem) BeforeCreate(tx *gorm.DB) error {
	if p.ContestID != nil {
		p.ID = ProblemID(Contest, *p.ContestID, p.Index)
	} else if p.ProblemsetName != nil {
		p.ID = ProblemID(ProblemSet, *p.ProblemsetName, p.Index)
	} else {
		p.ID = ProblemID(Unknow, "", p.Index)
	}
	return nil
}

func ProblemID(problemType ProblemType, id, index string) string {
	if problemType == Contest {
		return fmt.Sprintf("contest-%s-%s", id, index)
	} else if problemType == ProblemSet {
		return fmt.Sprintf("problemset-%s-%s", id, index)
	} else {
		return fmt.Sprintf("unknow-problem-%s", index)
	}
}
