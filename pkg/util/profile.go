package util

import (
	"sort"
	"strings"
)

type Profile struct {
	env   map[string]string
	isNil map[string]bool
}

func NewProfile() *Profile {
	return &Profile{env: make(map[string]string), isNil: make(map[string]bool)}
}

// Set will set an entry
func (profile *Profile) Set(name string, value string) {
	profile.env[name] = value
	delete(profile.isNil, name)
}

// SetNil will mark entry as nil
func (profile *Profile) SetNil(name string) {
	delete(profile.env, name)
	profile.isNil[name] = true
}

// Get retrieve value
func (profile *Profile) Get(name string) (string, bool) {
	value, ok := profile.env[name]
	return value, ok
}

// GetNil returns true if the value has been explitly set to nil.
func (profile *Profile) GetNil(name string) bool {
	_, ok := profile.isNil[name]
	return ok
}

func (profile *Profile) SortedNames(includeNil bool) []string {
	keys := make([]string, 0, len(profile.env) + len(profile.isNil))
	for k := range profile.env {
		keys = append(keys, k)
	}
	if includeNil {
		for k := range profile.isNil {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}

// Clone creates a new copy of the Profile.
func (profile *Profile) Clone() *Profile {
	p := NewProfile()
	for k, v := range profile.env {
		p.env[k] = v
	}
	for k, v := range profile.isNil {
		p.isNil[k] = v
	}
	return p
}

// Merge applies all elements from p into current profile.
func (profile *Profile) Merge(p *Profile) {
	for k, v := range p.env {
		profile.Set(k, v)
	}
	for k := range p.isNil {
		profile.SetNil(k)
	}
}

// MergeStrings parses a list of NAME=value text entries and adds to profile.
func (profile *Profile) MergeStrings(envList []string) {
	for _, kv := range envList {
		pair := strings.SplitN(kv, "=", 2)
		if len(pair) == 2 {
			profile.Set(pair[0], pair[1])
		} else {
			profile.SetNil(pair[0])
		}
	}
}
