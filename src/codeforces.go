package src

import (
	"github.com/YourSuzumiya/ACMBot/src/internal/api"
	"github.com/YourSuzumiya/ACMBot/src/internal/img"
)

func init() {
	registry.Store("codeforces", &CodeForcesRemote{})
}

type CodeForcesStatus struct {
}

func (cfs *CodeForcesStatus) String() string {
	panic("TODO")
}

func (cfs *CodeForcesStatus) Render(*img.Option) ([]byte, error) {
	panic("TODO")
}

type CodeForcesUser struct {
}

func (cfu *CodeForcesUser) String() string {
	panic("TODO")
}

func (cfu *CodeForcesUser) StatusOf(api.StatusType) (Status, error) {
	panic("TODO")
}

type CodeForcesRemote struct {
}

func (cfr *CodeForcesRemote) String() string {
	panic("TODO")
}

func (cfr *CodeForcesRemote) NewUser(api.UserID) (User, error) {
	panic("TODO")
}
