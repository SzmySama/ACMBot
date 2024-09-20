package render

import (
	"github.com/SzmySama/ACMBot/app/types"
)

type HtmlOptions struct {
	Path string
	HTML string
}

type CodeforcesUserProfileData struct {
	types.User

	Level CodeforcesRatingLevel
}

type CodeforcesRatingChangesData struct {
	RatingChangesMetaData []types.RatingChange
	Handle                string
}

type CodeforcesRatingLevel string

func ConvertRatingToLevel(rating int) CodeforcesRatingLevel {
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
