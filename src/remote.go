package src

import (
	"fmt"

	"github.com/RomiChan/syncx"
	"github.com/YourSuzumiya/ACMBot/src/internal/img"
)

type Status interface {
	fmt.Stringer
	Render(*img.Option) ([]byte, error)
}

type Remote interface {
	fmt.Stringer
	Race() (Status, error)
	Users(ids ...string) ([]Status, error)
}

var registry = syncx.Map[string, Remote]{}

func From(remote string) (r Remote, ok bool) {
	return registry.Load(remote)
}
