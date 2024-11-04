package fetcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
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
		fmt.Sprintf("%02d天%02d小时%02d分钟", int(leftTime.Hours())/24, int(leftTime.Hours())%24, int(leftTime.Minutes())%60),
		r.StartTime.In(time.Local).Format("2006-01-02 15:04:05"),
		dStr,
		r.Link,
	)
}

type updatable interface {
	updater(ctx *zero.Ctx)
	beforeUpdate()
	afterUpdate()
}

type cacheRace struct {
	data            []Race
	MessageSegments []message.MessageSegment
	Err             error
	LastUpdate      time.Time
	sync.RWMutex
}

func (c *cacheRace) updater(_ *zero.Ctx) {
	c.Err = errors.New("undefined updater")
}
func (c *cacheRace) beforeUpdate() {
	c.RWMutex.Lock()
}

func (c *cacheRace) afterUpdate() {
	c.RWMutex.Unlock()
}

func (c *cacheRace) genMessageSegment(ctx *zero.Ctx) {
	selfID := ctx.GetLoginInfo().Get("user_id").Int()

	var newMessageSegment []message.MessageSegment

	for _, v := range c.data {
		MessageID := ctx.SendPrivateMessage(selfID, v.String())
		if MessageID == 0 {
			c.Err = errors.New("failed to gen message")
			return
		}
		newMessageSegment = append(newMessageSegment, message.Node(MessageID))
	}

	c.MessageSegments = newMessageSegment
	c.LastUpdate = time.Now()
}

type CacheStuACMRace struct {
	cacheRace
}

func (r *CacheStuACMRace) updater(ctx *zero.Ctx) {
	url := "https://contests.sdutacm.cn/contests.json"
	res, err := http.Get(url)
	if err != nil {
		r.Err = fmt.Errorf("failed to fetch all race API: %v", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Errorf("failed to close response body: %v", err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		r.Err = fmt.Errorf("failed to read res body: %v", err)
		return
	}

	var races []Race

	if err = json.Unmarshal(body, &races); err != nil {
		r.Err = fmt.Errorf("failed to unmarshal res data: %v", err)
		return
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
	r.data = targetRace

	r.genMessageSegment(ctx)
}

type CacheCodeforcesRace struct {
	cacheRace
}

func (r *CacheCodeforcesRace) updater(ctx *zero.Ctx) {
	races, err := FetchCodeforcesContestList(false)
	if err != nil {
		r.Err = fmt.Errorf("failed to fetch codeforces race API: %v", err)
		return
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

	r.data = result

	r.genMessageSegment(ctx)
}

var (
	codeforcesRaces = &CacheCodeforcesRace{}
	stuAcmRaces     = &CacheStuACMRace{}
	allRace         = []updatable{
		codeforcesRaces,
		stuAcmRaces,
	}
)

func GetCodeforcesRaces() *CacheCodeforcesRace {
	codeforcesRaces.RLock()
	defer codeforcesRaces.RUnlock()
	return codeforcesRaces
}

func GetStuAcmRaces() *CacheStuACMRace {
	codeforcesRaces.RLock()
	defer codeforcesRaces.RUnlock()
	return stuAcmRaces
}

func Updater(ctx *zero.Ctx) {
	update := func() {
		logrus.Infof("正在更新比赛数据")
		for _, v := range allRace {
			v.beforeUpdate()
			v.updater(ctx)
			v.afterUpdate()
		}
	}

	update()
	ticker := time.NewTicker(6 * time.Hour)
	for range ticker.C {
		update()
	}
}
