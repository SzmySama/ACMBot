package fetcher

import (
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type CacheRaceData struct {
	Races                          []Race
	AllRacesMessageSegments        []message.MessageSegment
	CodeforcesRacesMessageSegments []message.MessageSegment
	UpdateAt                       time.Time
}

func (cr *CodeforcesRace) ToRace() *Race {
	return &Race{
		Source:    "Codeforces",
		Name:      cr.Name,
		Link:      "https://codeforces.com/contests/",
		StartTime: time.Unix(cr.StartTimeSeconds, 0),
		EndTime:   time.Unix(cr.StartTimeSeconds+int64(cr.DurationSeconds), 0),
	}
}
