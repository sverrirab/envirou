package util

import "sort"

// import "strings"

type Groups map[string]Patterns

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
