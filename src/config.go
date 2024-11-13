package src

import (
	"os"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	cf := &codeForcesRemote{}
	yaml.NewDecoder(f).Decode(f)
	registry.Store("codeforces", cf)
	panic("TODO")
}
