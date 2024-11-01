package fetcher

import (
	"fmt"
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
)

type codeforcesResponse[T any] struct {
	Status  string `json:"status"`
	Result  T      `json:"result"`
	Comment string `json:"comment"`
}

type Race struct {
	Source    string    `json:"source"`
	Name      string    `json:"name"`
	Link      string    `json:"link"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type CodeforcesRace struct {
	ID                  int    `json:"id"`
	Name                string `json:"name"`
	Type                string `json:"type"`
	Phase               string `json:"phase"`
	Frozen              bool   `json:"frozen"`
	DurationSeconds     int    `json:"durationSeconds"`
	StartTimeSeconds    int64  `json:"startTimeSeconds"`
	RelativeTimeSeconds int    `json:"relativeTimeSeconds"`
}

type CacheRaceData struct {
	Races                          []Race
	AllRacesMessageSegments        []message.MessageSegment
	CodeforcesRacesMessageSegments []message.MessageSegment
	UpdateAt                       time.Time
}

func (r *Race) String() string {
	d := r.EndTime.Sub(r.StartTime)
	var dStr string
	if h, m := int(d.Hours()), int(d.Minutes())%60; m > 0 {
		dStr = fmt.Sprintf("%d小时%d分钟", h, m)
	} else {
		dStr = fmt.Sprintf("%d小时", h)
	}
	return fmt.Sprintf(
		""+
			"比赛来源: %s\n"+
			"比赛名称: %s\n"+
			"开始时间: %s\n"+
			"持续时间: %s\n"+
			"传送门🌈: %s",
		r.Source,
		r.Name,
		r.StartTime.In(time.Local).Format("2006-01-02 15:04:05"),
		dStr,
		r.Link,
	)
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
