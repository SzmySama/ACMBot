package manager

import (
	"github.com/YourSuzumiya/ACMBot/app/model/fetcher"
	"github.com/wdvxdr1123/ZeroBot/message"
	"sync"
	"time"
)

type Races struct {
	Data []fetcher.Race
	Lock sync.RWMutex

	Err error

	updater func(*Races) error
}

func (r *Races) GetMessage() ([]message.MessageSegment, error) {
	r.Lock.RLock()
	defer r.Lock.RUnlock()
	result := make([]message.MessageSegment, 0, len(r.Data))
	for _, race := range r.Data {
		result = append(result, message.CustomNode("", 0, race.String()))
	}
	return result, r.Err
}

func (r *Races) Update() error {
	return r.updater(r)
}

var (
	codeforcesRaces = &Races{
		updater: func(r *Races) error {
			r.Lock.Lock()
			defer r.Lock.Unlock()
			race, err := fetcher.FetchCodeforcesRaces()
			if err != nil {
				return err
			}
			r.Data = race
			return nil
		},
	}
	stuACMRaces = &Races{
		updater: func(r *Races) error {
			r.Lock.Lock()
			defer r.Lock.Unlock()
			race, err := fetcher.FetchStuACMRaces()
			if err != nil {
				return err
			}
			r.Data = race
			return nil
		},
	}
	allRaces = []*Races{
		codeforcesRaces,
		stuACMRaces,
	}
)

func GetCodeforcesRaces() ([]message.MessageSegment, error) {
	return codeforcesRaces.GetMessage()
}

func GetStuACMRaces() ([]message.MessageSegment, error) {
	return stuACMRaces.GetMessage()
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
