package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

var (
	cacheRace       CacheRaceData
	avilableSources = []string{
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
	defer res.Body.Close()

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
		for _, raceSource := range avilableSources {
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

func GetAllRaces() ([]Race, error) {
	if time.Since(cacheRace.UpdateAt).Hours() > 24 {
		if err := fetchAllRaces(); err != nil {
			return cacheRace.Races, err
		}
		sort.Slice(cacheRace.Races, func(i, j int) bool {
			return cacheRace.Races[i].StartTime.Before(cacheRace.Races[j].StartTime)
		})
	}
	return cacheRace.Races, nil
}
