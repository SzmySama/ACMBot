package fetcher

import (
	"encoding/json"
	"fmt"
	"github.com/SzmySama/ACMBot/app/utils/slice"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"io"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/wdvxdr1123/ZeroBot/message"
)

var (
	cacheRace        CacheRaceData
	availableSources = []string{
		"牛客竞赛",
		"洛谷",
		"AtCoder",
		"Codeforces",
	}
)

func fetchAllRaces() error {
	url := "https://contests.sdutacm.cn/contests.json"
	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch all race API: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.Errorf("failed to close response body: %v", err)
		}
	}(res.Body)

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("failed to read res body: %v", err)
	}

	if err = json.Unmarshal(body, &cacheRace.Races); err != nil {
		return fmt.Errorf("failed to unmarshal res data: %v", err)
	}

	// filterRace

	var targetRace []Race

	for _, race := range cacheRace.Races {
		for _, raceSource := range availableSources {
			if race.Source == raceSource {
				targetRace = append(targetRace, race)
				continue
			}
		}
	}

	cacheRace.Races = targetRace

	cacheRace.UpdateAt = time.Now()
	return nil
}

func GetAndFetchRaces(ctx *zero.Ctx) (*CacheRaceData, error) {
	if time.Since(cacheRace.UpdateAt).Hours() > 24 {
		if err := fetchAllRaces(); err != nil {
			return &cacheRace, err
		}
		sort.Slice(cacheRace.Races, func(i, j int) bool {
			return cacheRace.Races[i].StartTime.Before(cacheRace.Races[j].StartTime)
		})
		cacheRace.AllRacesMessageSegments = cacheRace.AllRacesMessageSegments[0:0]
		cacheRace.CodeforcesRacesMessageSegments = cacheRace.CodeforcesRacesMessageSegments[0:0]

		BotQID, err := strconv.ParseInt(ctx.GetLoginInfo().Get("user_id").String(), 10, 64)
		if err != nil {
			fmt.Println("Error:", err)
			return &cacheRace, fmt.Errorf("failed to parse bot_id: %v", err)
		}
		for _, v := range cacheRace.Races {
			MessageID := ctx.SendPrivateMessage(BotQID, v.String())
			cacheRace.AllRacesMessageSegments = append(cacheRace.AllRacesMessageSegments, message.Node(MessageID))
		}

		// 近期cf直接从codeforces的API获取, 下面在获取codeforces
		codeforcesRaces, err := FetchCodeforcesContestList(false)
		if err != nil {
			return &cacheRace, fmt.Errorf("failed to fetch codeforces: %v", err)
		}
		for _, v := range *codeforcesRaces {
			if time.Unix(v.StartTimeSeconds, 0).Before(time.Now()) {
				break
			}
			if v.Phase != "BEFORE" {
				continue
			}
			MessageID := ctx.SendPrivateMessage(BotQID, v.ToRace().String())
			cacheRace.CodeforcesRacesMessageSegments = append(cacheRace.CodeforcesRacesMessageSegments, message.Node(MessageID))
		}
		slice.Reverse(cacheRace.CodeforcesRacesMessageSegments)
	}
	return &cacheRace, nil
}
