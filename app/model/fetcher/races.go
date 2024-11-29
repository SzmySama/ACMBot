package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	availableSources = []string{
		"牛客竞赛",
		"洛谷",
		"AtCoder",
		"Codeforces",
	}
)

type Race struct {
	Source    string    `json:"source"`
	Name      string    `json:"name"`
	Link      string    `json:"link"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (r *Race) String() string {
	d := r.EndTime.Sub(r.StartTime)
	var dStr string
	if h, m := int(d.Hours()), int(d.Minutes())%60; m > 0 {
		dStr = fmt.Sprintf("%d小时%d分钟", h, m)
	} else {
		dStr = fmt.Sprintf("%d小时", h)
	}

	leftTime := r.StartTime.Sub(time.Now())

	return fmt.Sprintf(
		""+
			"比赛来源: %s\n"+
			"比赛名称: %s\n"+
			"剩余时间: %s\n"+
			"开始时间: %s\n"+
			"持续时间: %s\n"+
			"传送门🌈: %s",
		r.Source,
		r.Name,
		fmt.Sprintf("%02d天%02d小时%02d分钟", int(leftTime.Hours())/24, abs(int(leftTime.Hours()))%24, abs(int(leftTime.Minutes()))%60),
		r.StartTime.In(time.Local).Format("2006-01-02 15:04:05"),
		dStr,
		r.Link,
	)
}

func FetchStuACMRaces() ([]Race, error) {
	url := "https://contests.sdutacm.cn/contests.json"
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all race API: %v", err)

	}

	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			logrus.Errorf("failed to close response body: %v", err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read res body: %v", err)

	}

	var races []Race

	if err = json.Unmarshal(body, &races); err != nil {
		return nil, fmt.Errorf("failed to unmarshal res data: %v", err)
	}

	// filterRace

	var targetRace []Race

	for _, race := range races {
		for _, raceSource := range availableSources {
			if race.Source == raceSource {
				targetRace = append(targetRace, race)
				continue
			}
		}
	}
	return targetRace, nil
}

func FetchCodeforcesRaces() ([]Race, error) {
	races, err := FetchCodeforcesContestList(false)
	if err != nil {
		return nil, err
	}
	result := make([]Race, 0, len(*races))
	for _, race := range *races {
		if race.RelativeTimeSeconds > 0 {
			break
		}

		result = append(result, Race{
			Source:    "Codeforces",
			Name:      race.Name,
			Link:      "https://codeforces.com/contests",
			StartTime: time.Unix(race.StartTimeSeconds, 0),
			EndTime:   time.Unix(race.StartTimeSeconds+race.DurationSeconds, 0),
		})
	}
	slices.Reverse(result)
	return result, nil
}
