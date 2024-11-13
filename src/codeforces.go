package src

import (
	"github.com/YourSuzumiya/ACMBot/src/internal/img"
)

type codeForcesStatusUser struct {
}

func (cfs *codeForcesStatusUser) String() string {
	panic("TODO")
}

func (cfs *codeForcesStatusUser) Render(*img.Option) ([]byte, error) {
	panic("TODO")
}

type codeForcesStatusRace struct {
}

func (cfs *codeForcesStatusRace) String() string {
	panic("TODO")
}

func (cfs *codeForcesStatusRace) Render(*img.Option) ([]byte, error) {
	panic("TODO")
}

type codeForcesRemote struct {
	Key    string `mapstructure:"key" toml:"key"`
	Secret string `mapstructure:"secret" toml:"secret"`
}

func (cfr *codeForcesRemote) String() string {
	panic("TODO")
}

func (cfr *codeForcesRemote) Race() (Status, error) {
	return &codeForcesStatusRace{}, nil
}

func (cfr *codeForcesRemote) Users(...string) ([]Status, error) {
	return []Status{&codeForcesStatusUser{}, &codeForcesStatusUser{}}, nil
}
