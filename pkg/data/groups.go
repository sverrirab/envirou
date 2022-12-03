package data

import (
	"fmt"
	"sort"
	"strings"
)

type Groups map[string]Patterns
type Envs []string
type GroupNameToEnvs map[string]Envs

func NewGroups() *Groups {
	g := make(map[string]Patterns)
	return (*Groups)(&g)
}

func (groups *Groups) ParseAndAdd(name string, patterns string) {
	(*groups)[name] = *ParsePatterns(patterns)
}

func (groups *Groups) GetPatterns(name string) (*Patterns, bool) {
	g, found := (*groups)[name]
	if !found {
		return nil, false
	}
	return &g, true
}

// GetAllNames returns all names sorted.
func (groups *Groups) GetAllNames() []string {
	keys := make([]string, 0, len(*groups))
	for key := range *groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (groups Groups) String() string {
	names := groups.GetAllNames()
	result := make([]string, 0, len(names))
	for _, name := range names {
		result = append(result, fmt.Sprintf("%s=%s", name, groups[name]))
	}
	return strings.Join(result, " | ")
}

// MatchAll returns a map of group names to all env variables as well as a list of unmatched ones
func (groups *Groups) MatchAll(envs Envs) (GroupNameToEnvs, Envs) {
	result := make(GroupNameToEnvs, len(*groups))
	matched := make(map[string]bool, len(*groups))
	unmatched := make(Envs, 0)

	for _, env := range envs {
		for _, group := range groups.GetAllNames() {
			patterns, found := groups.GetPatterns(group)
			if !found {
				continue
			}
			if MatchAny(env, patterns) {
				matched[env] = true
				result[group] = append(result[group], env)
			}
		}
	}
	for _, env := range envs {
		_, found := matched[env]
		if !found {
			unmatched = append(unmatched, env)
		}
	}
	return result, unmatched
}