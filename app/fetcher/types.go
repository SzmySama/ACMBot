package fetcher

import (
	"fmt"
	"time"
)

type codeforcesResponse struct {
	Status  string           `json:"status"`
	Result  []map[string]any `json:"result"`
	Comment string           `json:"comment"`
}

type Race struct {
	Source    string    `json:"source"`
	Name      string    `json:"name"`
	Link      string    `json:"link"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

type AllRace struct {
	Races    []Race
	UpdateAt time.Time
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
