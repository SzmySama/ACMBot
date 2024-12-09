package fetcher

import (
	"github.com/YourSuzumiya/ACMBot/app/model"
	"time"
)

func (cr *CodeforcesRace) ToRace() *model.Race {
	return &model.Race{
		Source:    "Codeforces",
		Name:      cr.Name,
		Link:      "https://codeforces.com/contests/",
		StartTime: time.Unix(cr.StartTimeSeconds, 0),
		EndTime:   time.Unix(cr.StartTimeSeconds+cr.DurationSeconds, 0),
	}
}
