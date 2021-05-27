package state

import (
	"os"

	"github.com/sverrirab/envirou/pkg/util"
)

type Envirou struct {
	Env        map[string]string
	SortedKeys []string
}

func NewEnvirou(envList []string) *Envirou {
	envMap, sortedKeys := util.ParseEnvironment(os.Environ())
	env := Envirou{Env: envMap, SortedKeys: sortedKeys}
	return &env
}
