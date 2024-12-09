package manager

import (
	"github.com/YourSuzumiya/ACMBot/app/fetcher"
	"github.com/YourSuzumiya/ACMBot/app/model"
	"sync"
	"time"
)

// TODO: 比赛数据缓存到Redis

type Races struct {
	Data []model.Race
	Lock sync.RWMutex

	Err error

	provider model.RaceProvider
}

func (r *Races) Update() error {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	race, err := r.provider()
	if err != nil {
		return err
	}
	r.Data = race
	return nil
}

var (
	codeforcesRaces = &Races{
		provider: fetcher.FetchCodeforcesRaces,
	}
	stuACMRaces = &Races{
		provider: fetcher.FetchStuACMRaces,
	}
	atcoderRaces = &Races{
		provider: fetcher.FetchAtCoderRaces,
	}
	nowcoderRaces = &Races{
		provider: fetcher.FetchNowCoderRaces,
	}
	luoguRaces = &Races{
		provider: fetcher.FetchLuoguRaces,
	}
	allRaces = []*Races{
		nowcoderRaces,
		luoguRaces,
		atcoderRaces,
		codeforcesRaces,
		stuACMRaces,
	}
)

func GetCodeforcesRaces() ([]model.Race, error) {
	return codeforcesRaces.Data, codeforcesRaces.Err
}

func GetStuACMRaces() ([]model.Race, error) {
	return stuACMRaces.Data, stuACMRaces.Err
}

func GetAtCoderRaces() ([]model.Race, error) {
	return atcoderRaces.Data, atcoderRaces.Err
}

func GetNowCoderRaces() ([]model.Race, error) {
	return nowcoderRaces.Data, nowcoderRaces.Err
}

func GetLuoguRaces() ([]model.Race, error) {
	return luoguRaces.Data, luoguRaces.Err
}

func RaceUpdater() {
	update := func() {
		for _, race := range allRaces {
			race.Err = race.Update()
		}
	}

	ticker := time.NewTicker(5 * time.Hour)

	update()
	for range ticker.C {
		update()
	}
}
