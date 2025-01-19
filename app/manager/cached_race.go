package manager

import (
	"fmt"
	"github.com/YourSuzumiya/ACMBot/app/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/model"
	"github.com/YourSuzumiya/ACMBot/app/model/cache"
	"sort"
	"time"
)

const cacheExp = 24 * time.Hour
const updateExp = 5 * time.Hour

var AllResource = model.AllRaceResource

type CachedRace struct {
	source   model.Resource
	provider model.RaceProvider
	err      error
}

func (r *CachedRace) Update() error {
	race, err := r.provider()
	if err != nil {
		return err
	}
	return cache.SetRace(r.source, race, cacheExp)
}

func (r *CachedRace) Get() ([]model.Race, error) {
	if r.err != nil {
		return nil, r.err
	}
	return cache.GetRace(r.source)
}

type updater struct {
	AllCachedRace map[model.Resource]*CachedRace
	UpdateTicker  *time.Ticker
}

func (r *updater) update() {
	for _, race := range r.AllCachedRace {
		err := race.Update()
		if err != nil {
			race.err = fmt.Errorf("update error: %v", err)
		}
	}
}

func (r *updater) get(resource model.Resource) ([]model.Race, error) {
	race, ok := r.AllCachedRace[resource]
	if !ok {
		return nil, fmt.Errorf("%s not found", resource)
	}
	return race.Get()
}

func (r *updater) start() {
	r.update()
	for range r.UpdateTicker.C {
		r.update()
	}
}

func newUpdater(rp map[model.Resource]model.RaceProvider, t *time.Ticker) *updater {
	result := &updater{
		AllCachedRace: make(map[model.Resource]*CachedRace),
		UpdateTicker:  t,
	}
	for source, provider := range rp {
		result.AllCachedRace[source] = &CachedRace{source, provider, nil}
	}
	return result
}

var (
	raceAndProvider = map[model.Resource]model.RaceProvider{
		model.ResourceCodeforces: fetcher.FetchClistCodeforcesContests,
		model.ResourceAtcoder:    fetcher.FetchClistAtcoderContests,
		model.ResourceLeetcode:   fetcher.FetchClistLeetcodeContests,
		model.ResourceLuogu:      fetcher.FetchClistLuoguContests,
		model.ResourceNowcoder:   fetcher.FetchClistNowcoderContests,
	}
	defaultUpdater = newUpdater(raceAndProvider, time.NewTicker(updateExp))
)

func init() {
	go defaultUpdater.start()
}

func GetCachedRacesByResource(resource model.Resource) model.RaceProvider {
	return func() ([]model.Race, error) {
		return defaultUpdater.get(resource)
	}
}

func GetAllCachedRaces() ([]model.Race, error) {
	var results []model.Race
	for _, s := range AllResource {
		races, err := defaultUpdater.get(s)
		if err != nil {
			continue
		}
		results = append(results, races...)
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].StartTime.Before(results[j].StartTime)
	})
	return results, nil
}
