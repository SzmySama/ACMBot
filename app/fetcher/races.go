package fetcher

import (
	"encoding/json"
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/model"
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

func FetchStuACMRaces() ([]model.Race, error) {
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

	var races []model.Race

	if err = json.Unmarshal(body, &races); err != nil {
		return nil, fmt.Errorf("failed to unmarshal res data: %v", err)
	}

	// filterRace

	var targetRace []model.Race

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

func FetchCodeforcesRaces() ([]model.Race, error) {
	races, err := FetchCodeforcesContestList(false)
	if err != nil {
		return nil, err
	}
	result := make([]model.Race, 0, len(races))
	for _, race := range races {
		if race.RelativeTimeSeconds > 0 {
			break
		}

		result = append(result, model.Race{
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

func FetchAtCoderRaces() ([]model.Race, error) {
	race, err := FetchStuACMRaces()
	if err != nil {
		return nil, err
	}
	result := make([]model.Race, 0, len(race))
	for _, race := range race {
		if race.Source == "AtCoder" {
			result = append(result, race)
		}
	}
	slices.Reverse(result)
	return result, nil
}

func FetchNowCoderRaces() ([]model.Race, error) {
	race, err := FetchStuACMRaces()
	if err != nil {
		return nil, err
	}
	result := make([]model.Race, 0, len(race))
	for _, race := range race {
		if race.Source == "牛客竞赛" {
			result = append(result, race)
		}
	}
	slices.Reverse(result)
	return result, nil
}

func FetchLuoguRaces() ([]model.Race, error) {
	race, err := FetchStuACMRaces()
	if err != nil {
		return nil, err
	}
	result := make([]model.Race, 0, len(race))
	for _, race := range race {
		if race.Source == "洛谷" {
			result = append(result, race)
		}
	}
	slices.Reverse(result)
	return result, nil
}
