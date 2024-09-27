package types

import (
	"encoding/json"
	"time"
)

type SubmissionStatus string

const (
	SubmissionStatusFailed                  SubmissionStatus = "FAILED"
	SubmissionStatusOk                      SubmissionStatus = "OK"
	SubmissionStatusPartial                 SubmissionStatus = "PARTIAL"
	SubmissionStatusCompilationError        SubmissionStatus = "COMPILATION_ERROR"
	SubmissionStatusRuntimeError            SubmissionStatus = "RUNTIME_ERROR"
	SubmissionStatusWrongAnswer             SubmissionStatus = "WRONG_ANSWER"
	SubmissionStatusPresentationError       SubmissionStatus = "PRESENTATION_ERROR"
	SubmissionStatusTimeLimitExceeded       SubmissionStatus = "TIME_LIMIT_EXCEEDED"
	SubmissionStatusMemoryLimitExceeded     SubmissionStatus = "MEMORY_LIMIT_EXCEEDED"
	SubmissionStatusIdlenessLimitExceeded   SubmissionStatus = "IDLENESS_LIMIT_EXCEEDED"
	SubmissionStatusSecurityViolated        SubmissionStatus = "SECURITY_VIOLATED"
	SubmissionStatusCrashed                 SubmissionStatus = "CRASHED"
	SubmissionStatusInputPreparationCrashed SubmissionStatus = "INPUT_PREPARATION_CRASHED"
	SubmissionStatusChallenged              SubmissionStatus = "CHALLENGED"
	SubmissionStatusSkipped                 SubmissionStatus = "SKIPPED"
	SubmissionStatusTesting                 SubmissionStatus = "TESTING"
	SubmissionStatusRejected                SubmissionStatus = "REJECTED"
)

type User struct {
	Handle    string `gorm:"primaryKey;not null;type:varchar(255)" json:"handle"`
	Avatar    string `json:"avatar"`
	Rating    int    `json:"rating"`
	Solved    int
	FriendOf  int       `json:"friendOfCount"`
	CreatedAt time.Time `json:"-"`

	Submissions         []Submission `gorm:"serializer:json"`
	SubmissionUpdatedAt time.Time

	RatingChanges        []RatingChange `gorm:"serializer:json"`
	RatingChangeUpdateAt time.Time
}

type Problem struct {
	ContestID      int      `json:"contestId"`
	ProblemSetName string   `json:"problemsetName"`
	Index          string   `json:"index"`
	Rating         int      `json:"rating"`
	Tags           []string `json:"tags"`
}

type Submission struct {
	At      time.Time        `json:"-"`
	Status  SubmissionStatus `json:"verdict"`
	Problem Problem          `json:"problem"`
}

type RatingChange struct {
	At        time.Time `json:"-"`
	NewRating int       `json:"newRating"`
}

func (u *User) UnmarshalJSON(data []byte) error {
	type alias User
	aux := &struct {
		T int64 `json:"registrationTimeSeconds"`
		*alias
	}{
		alias: (*alias)(u),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	u.CreatedAt = time.Unix(aux.T, 0)
	return nil
}

func (s *Submission) UnmarshalJSON(data []byte) error {
	type alias Submission
	aux := &struct {
		T int64 `json:"creationTimeSeconds"`
		*alias
	}{
		alias: (*alias)(s),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	s.At = time.Unix(aux.T, 0)
	return nil
}

func (r *RatingChange) UnmarshalJSON(data []byte) error {
	type alias RatingChange
	aux := &struct {
		T int64 `json:"ratingUpdateTimeSeconds"`
		*alias
	}{
		alias: (*alias)(r),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	r.At = time.Unix(aux.T, 0)
	return nil
}

func (u *User) MarshalJSON() ([]byte, error) {
	type alias User
	return json.Marshal(&struct {
		T int64 `json:"registrationTimeSeconds"`
		*alias
	}{
		T:     u.CreatedAt.Unix(),
		alias: (*alias)(u),
	})
}

func (s *Submission) MarshalJSON() ([]byte, error) {
	type alias Submission
	return json.Marshal(&struct {
		T int64 `json:"creationTimeSeconds"`
		*alias
	}{
		T:     s.At.Unix(),
		alias: (*alias)(s),
	})
}

func (r *RatingChange) MarshalJSON() ([]byte, error) {
	type alias RatingChange
	return json.Marshal(&struct {
		T int64 `json:"ratingUpdateTimeSeconds"`
		*alias
	}{
		T:     r.At.Unix(),
		alias: (*alias)(r),
	})
}
