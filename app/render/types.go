package render

import (
	"encoding/json"
	"github.com/YourSuzumiya/ACMBot/app/model/db"
	"time"
)

type HtmlOptions struct {
	Path string
	HTML string
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

func DB2RenderUser(user db.CodeforcesUser) CodeforcesUser {
	var solved uint

	result := db.GetDBConnection().Raw(`
		SELECT COUNT(DISTINCT codeforces_problem_id) 
		FROM codeforces_submissions 
		WHERE codeforces_user_id = ? AND status = ?`,
		user.ID, db.CodeforcesSubmissionStatusOk).Scan(&solved)

	if result.Error != nil {
		solved = 0
	}

	return CodeforcesUser{
		Handle:    user.Handle,
		Avatar:    user.Avatar,
		Rating:    user.Rating,
		Solved:    solved,
		FriendOf:  user.FriendCount,
		CreatedAt: user.CreatedAt,
		Level:     ConvertRatingToLevel(user.Rating),
	}
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

func DB2RenderCodeforcesRatingChanges(changes []db.CodeforcesRatingChange) []CodeforcesRatingChange {
	var result []CodeforcesRatingChange
	for _, change := range changes {
		result = append(result, CodeforcesRatingChange{
			At:        change.At,
			NewRating: change.NewRating,
		})
	}
	return result
}

type CodeforcesRatingChangesData struct {
	RatingChangesMetaData []CodeforcesRatingChange
	Handle                string
}

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
