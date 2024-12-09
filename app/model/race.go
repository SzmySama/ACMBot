package model

import (
	"fmt"
	"time"
)

type Race struct {
	Source    string    `json:"source"`
	Name      string    `json:"name"`
	Link      string    `json:"link"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func (r *Race) String() string {
	d := r.EndTime.Sub(r.StartTime)
	var dStr string
	if h, m := int(d.Hours()), int(d.Minutes())%60; m > 0 {
		dStr = fmt.Sprintf("%d小时%d分钟", h, m)
	} else {
		dStr = fmt.Sprintf("%d小时", h)
	}

	startLeftTime := r.StartTime.Sub(time.Now())
	endLeftTime := r.EndTime.Sub(time.Now())

	started := startLeftTime.Milliseconds() < 0
	finished := endLeftTime.Milliseconds() < 0

	if !started {
		return fmt.Sprintf(
			""+
				"🕣此比赛尚未开始🕦\n"+
				"比赛来源: %s\n"+
				"比赛名称: %s\n"+
				"距离开始: %s\n"+
				"开始时间: %s\n"+
				"持续时间: %s\n"+
				"传送门🌈: %s",
			r.Source,
			r.Name,
			fmt.Sprintf("%02d天%02d小时%02d分钟", int(startLeftTime.Hours())/24, abs(int(startLeftTime.Hours()))%24, abs(int(startLeftTime.Minutes()))%60),
			r.StartTime.In(time.Local).Format("2006-01-02 15:04:05"),
			dStr,
			r.Link,
		)
	}
	if !finished {
		return fmt.Sprintf(
			""+
				"❗此比赛正在进行中❗\n"+
				"比赛来源: %s\n"+
				"比赛名称: %s\n"+
				"距离结束: %s\n"+
				"开始时间: %s\n"+
				"持续时间: %s\n"+
				"传送门🌈: %s",
			r.Source,
			r.Name,
			fmt.Sprintf("%02d天%02d小时%02d分钟", int(endLeftTime.Hours())/24, abs(int(endLeftTime.Hours()))%24, abs(int(endLeftTime.Minutes()))%60),
			r.StartTime.In(time.Local).Format("2006-01-02 15:04:05"),
			dStr,
			r.Link,
		)
	}
	return "Internal ERROR! Finished race shouldn't exist!"
}
