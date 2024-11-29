package manager

import (
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/wdvxdr1123/ZeroBot/message"
	"sync"
	"time"
)

// TODO: 比赛数据缓存到Redis

type Races struct {
	Data []fetcher.Race
	Lock sync.RWMutex

	Err error

	updater func() ([]fetcher.Race, error)
}

func (r *Races) ToQQMixForwardMessage() ([]message.MessageSegment, error) {
	r.Lock.RLock()
	defer r.Lock.RUnlock()
	result := make([]message.MessageSegment, 0, len(r.Data))
	for _, race := range r.Data {
		result = append(result, message.CustomNode("", 0, race.String()))
	}
	return result, r.Err
}

func (r *Races) Update() error {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	race, err := r.updater()
	if err != nil {
		return err
	}
	r.Data = race
	return nil
}

var (
	codeforcesRaces = &Races{
		updater: fetcher.FetchCodeforcesRaces,
	}
	stuACMRaces = &Races{
		updater: fetcher.FetchStuACMRaces,
	}
	allRaces = []*Races{
		codeforcesRaces,
		stuACMRaces,
	}
)

func GetCodeforcesRaces() *Races {
	return codeforcesRaces
}

func GetStuACMRaces() *Races {
	return stuACMRaces
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
