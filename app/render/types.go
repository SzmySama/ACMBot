package render

import (
	"github.com/SzmySama/ACMBot/app/types"
)

type RenderHTMLOptions struct {
	Path string
	HTML string
}

type CodeforcesUserProfileData struct {
	types.User

	Level codeforcesRatingLevel
}

type CodeforcesRatingChangesData struct {
	RatingChangesMetaData []types.RatingChange
	Handle                string
}

type codeforcesRatingLevel string

func ConvertRatingToLevel(rating int) codeforcesRatingLevel {
	const (
		CodeforcesRatingLevelNewbie                   codeforcesRatingLevel = "newbie"
		CodeforcesRatingLevelPupil                    codeforcesRatingLevel = "pupil"
		CodeforcesRatingLevelSpecialist               codeforcesRatingLevel = "specialist"
		CodeforcesRatingLevelExpert                   codeforcesRatingLevel = "expert"
		CodeforcesRatingLevelCandidateMaster          codeforcesRatingLevel = "candidate-master"
		CodeforcesRatingLevelMaster                   codeforcesRatingLevel = "master"
		CodeforcesRatingLevelInternationalMaster      codeforcesRatingLevel = "international-master"
		CodeforcesRatingLevelGrandmaster              codeforcesRatingLevel = "grandmaster"
		CodeforcesRatingLevelInternationalGrandmaster codeforcesRatingLevel = "international-grandmaster"
		CodeforcesRatingLevelLegendaryGrandmaster     codeforcesRatingLevel = "legendary-grandmaster"
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
