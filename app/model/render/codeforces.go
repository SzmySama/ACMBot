package render

import (
	"bytes"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model/errs"
	"github.com/playwright-community/playwright-go"
)

type CodeforcesUser struct {
	Handle   string `gorm:"primaryKey;not null;type:varchar(255)" json:"handle"`
	Avatar   string `json:"avatar"`
	Rating   int    `json:"rating"`
	Solved   int
	FriendOf int `json:"friendOfCount"`
	Level    CodeforcesRatingLevel
}

func (u *CodeforcesUser) ToImage() ([]byte, error) {
	var buffer bytes.Buffer
	if err := codeforcesUserProfileV1Template.Execute(&buffer, u); err != nil {
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
	At        int64 `json:"at"`
	NewRating int   `json:"newRating"`
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

type CodeforcesRatingLevel string

func ConvertRatingToLevel(rating int) CodeforcesRatingLevel {
	const (
		CodeforcesRatingLevelNewbie                   CodeforcesRatingLevel = "Newbie"
		CodeforcesRatingLevelPupil                    CodeforcesRatingLevel = "Pupil"
		CodeforcesRatingLevelSpecialist               CodeforcesRatingLevel = "Specialist"
		CodeforcesRatingLevelExpert                   CodeforcesRatingLevel = "Expert"
		CodeforcesRatingLevelCandidateMaster          CodeforcesRatingLevel = "CM"
		CodeforcesRatingLevelMaster                   CodeforcesRatingLevel = "Master"
		CodeforcesRatingLevelInternationalMaster      CodeforcesRatingLevel = "IM"
		CodeforcesRatingLevelGrandmaster              CodeforcesRatingLevel = "GM"
		CodeforcesRatingLevelInternationalGrandmaster CodeforcesRatingLevel = "IGM"
		CodeforcesRatingLevelLegendaryGrandmaster     CodeforcesRatingLevel = "LGM"
		CodeforcesRatingLevelTourist                  CodeforcesRatingLevel = "Tourist"
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
	case rating < 4000:
		return CodeforcesRatingLevelLegendaryGrandmaster
	default:
		return CodeforcesRatingLevelTourist
	}
}

type CodeforcesUserSolvedData struct {
	Range   string
	Count   int
	Percent float32
}

type CodeforcesUserProfile struct {
	Avatar    string
	Handle    string
	MaxRating int
	FriendOf  int
	Rating    int
	Level     CodeforcesRatingLevel
	Solved    int

	SolvedData []CodeforcesUserSolvedData
}

func (u *CodeforcesUserProfile) ToImage() ([]byte, error) {
	var buffer bytes.Buffer
	if err := codeforcesUserProfileV2Template.Execute(&buffer, u); err != nil {
		return nil, Error{fmt.Sprintf("failed to execute template: %v", err)}
	}
	return Html(
		&playwright.BrowserNewPageOptions{
			DeviceScaleFactor: &[]float64{2.0}[0],
			Viewport: &playwright.Size{
				Width:  300,
				Height: 400,
			},
		}, &HtmlOptions{
			Path: fullTemplatePath,
			HTML: buffer.String(),
		},
	)
}
