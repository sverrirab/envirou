package util

import (
	"sort"
	"strings"
)

// ParseEnvironment parses a list of NAME=value text entries and returns a map and sorted name list
func ParseEnvironment(envList []string) (map[string]string, []string) {
	env := make(map[string]string)
	keys := make([]string, 0, len(envList))
	for _, kv := range envList {
		pair := strings.SplitN(kv, "=", 2)
		keys = append(keys, pair[0])
		env[pair[0]] = pair[1]
	}
	sort.Strings(keys)
	return env, keys
}
