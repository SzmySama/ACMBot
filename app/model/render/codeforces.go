package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/errs"
	"github.com/playwright-community/playwright-go"
	"time"
)

type CodeforcesRatingLevel string

func ConvertRatingToLevel(rating uint) CodeforcesRatingLevel {
	const (
		CodeforcesRatingLevelNewbie                   CodeforcesRatingLevel = "newbie"
		CodeforcesRatingLevelPupil                    CodeforcesRatingLevel = "pupil"
		CodeforcesRatingLevelSpecialist               CodeforcesRatingLevel = "specialist"
		CodeforcesRatingLevelExpert                   CodeforcesRatingLevel = "expert"
		CodeforcesRatingLevelCandidateMaster          CodeforcesRatingLevel = "candidate-master"
		CodeforcesRatingLevelMaster                   CodeforcesRatingLevel = "master"
		CodeforcesRatingLevelInternationalMaster      CodeforcesRatingLevel = "international-master"
		CodeforcesRatingLevelGrandmaster              CodeforcesRatingLevel = "grandmaster"
		CodeforcesRatingLevelInternationalGrandmaster CodeforcesRatingLevel = "international-grandmaster"
		CodeforcesRatingLevelLegendaryGrandmaster     CodeforcesRatingLevel = "legendary-grandmaster"
	)
	switch {
	case rating < 1200:
		return CodeforcesRatingLevelNewbie
	case rating < 1400:
		return CodeforcesRatingLevelPupil
	case rating < 1600:
		return CodeforcesRatingLevelSpecialist
	case rating < 1900:
		return CodeforcesRatingLevelExpert
	case rating < 2100:
		return CodeforcesRatingLevelCandidateMaster
	case rating < 2300:
		return CodeforcesRatingLevelMaster
	case rating < 2400:
		return CodeforcesRatingLevelInternationalMaster
	case rating < 2600:
		return CodeforcesRatingLevelGrandmaster
	case rating < 3000:
		return CodeforcesRatingLevelInternationalGrandmaster
	default:
		return CodeforcesRatingLevelLegendaryGrandmaster
	}
}

type CodeforcesUser struct {
	Handle    string `gorm:"primaryKey;not null;type:varchar(255)" json:"handle"`
	Avatar    string `json:"avatar"`
	Rating    uint   `json:"rating"`
	Solved    uint
	FriendOf  uint      `json:"friendOfCount"`
	CreatedAt time.Time `json:"-"`
	Level     CodeforcesRatingLevel
}

func (u *CodeforcesUser) MarshalJSON() ([]byte, error) {
	type alias CodeforcesUser
	return json.Marshal(&struct {
		T int64 `json:"registrationTimeSeconds"`
		*alias
	}{
		T:     u.CreatedAt.Unix(),
		alias: (*alias)(u),
	})
}

func (u *CodeforcesUser) ToImage() ([]byte, error) {
	var buffer bytes.Buffer
	if err := codeforcesUserProfileTemplate.Execute(&buffer, u); err != nil {
		return nil, Error{fmt.Sprintf("failed to execute template: %v", err)}
	}
	return Html(
		&playwright.BrowserNewPageOptions{
			DeviceScaleFactor: &[]float64{2.0}[0],
			Viewport: &playwright.Size{
				Width:  400,
				Height: 225,
			},
		}, &HtmlOptions{
			Path: fullTemplatePath,
			HTML: buffer.String(),
		},
	)
}

type CodeforcesRatingChange struct {
	At        time.Time `json:"-"`
	NewRating int       `json:"newRating"`
}

func (r *CodeforcesRatingChange) MarshalJSON() ([]byte, error) {
	type alias CodeforcesRatingChange
	return json.Marshal(&struct {
		T int64 `json:"ratingUpdateTimeSeconds"`
		*alias
	}{
		T:     r.At.Unix(),
		alias: (*alias)(r),
	})
}

type CodeforcesRatingChanges struct {
	Data   []CodeforcesRatingChange
	Handle string
}

func (r *CodeforcesRatingChanges) ToImage() ([]byte, error) {
	if len(r.Data) == 0 {
		return nil, errs.ErrNoRatingChanges
	}

	var buffer bytes.Buffer
	if err := codeforcesRatingChangeTemplate.Execute(&buffer, r); err != nil {
		return nil, Error{fmt.Sprintf("failed to execute template: %v", err)}
	}

	return Html(
		&playwright.BrowserNewPageOptions{
			DeviceScaleFactor: &[]float64{2.0}[0],
			Viewport: &playwright.Size{
				Width:  1000,
				Height: 500,
			},
		}, &HtmlOptions{
			Path: fullTemplatePath,
			HTML: buffer.String(),
		},
	)
}
