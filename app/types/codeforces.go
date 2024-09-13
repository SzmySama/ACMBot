package types

import (
	"encoding/json"
	"time"
)

type SubmissionStatus string

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
	Rating int      `json:"rating"`
	Tags   []string `json:"tags"`
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

func (u User) MarshalJSON() ([]byte, error) {
	type alias User
	return json.Marshal(&struct {
		T int64 `json:"registrationTimeSeconds"`
		*alias
	}{
		T:     u.CreatedAt.Unix(),
		alias: (*alias)(&u),
	})
}

func (s Submission) MarshalJSON() ([]byte, error) {
	type alias Submission
	return json.Marshal(&struct {
		T int64 `json:"creationTimeSeconds"`
		*alias
	}{
		T:     s.At.Unix(),
		alias: (*alias)(&s),
	})
}

func (r RatingChange) MarshalJSON() ([]byte, error) {
	type alias RatingChange
	return json.Marshal(&struct {
		T int64 `json:"ratingUpdateTimeSeconds"`
		*alias
	}{
		T:     r.At.Unix(),
		alias: (*alias)(&r),
	})
}
