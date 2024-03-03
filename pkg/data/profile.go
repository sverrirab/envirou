package data

import (
	"runtime"
	"sort"
	"strings"
)

type Profile struct {
	env       map[string]string // The actual key value pair
	rightCase map[string]string // VAR -> Var (maps to actual case, used for case insensitive comparison on Windows)
	isNil     map[string]bool   // True if item is to be removed (uses UPPER case name)
}
type Profiles map[string]Profile

func NewProfile() *Profile {
	return &Profile{env: make(map[string]string), rightCase: make(map[string]string), isNil: make(map[string]bool)}
}

func (profile *Profile) GetCorrectCase(name string, create bool) string {
	if runtime.GOOS == "windows" {
		upper := strings.ToUpper(name)
		existingCase, exists := profile.rightCase[upper]
		if exists {
			return existingCase
		} else if create {
			profile.rightCase[upper] = name
		}
	}
	return name
}

// Set will set an entry
func (profile *Profile) Set(name string, value string) {
	correctCase := profile.GetCorrectCase(name, true)
	profile.env[correctCase] = value
	delete(profile.isNil, correctCase)
}

// SetNil will mark entry as nil
func (profile *Profile) SetNil(name string) {
	correctCase := profile.GetCorrectCase(name, true)
	_, exists := profile.env[correctCase]
	if exists {
		delete(profile.env, correctCase)
	}
	profile.isNil[correctCase] = true
}

// Get retrieve value
func (profile *Profile) Get(name string) (string, bool) {
	correctCase := profile.GetCorrectCase(name, false)
	value, ok := profile.env[correctCase]
	return value, ok
}

// GetNil returns true if the value has been explitly set to nil.
func (profile *Profile) GetNil(name string) bool {
	correctCase := profile.GetCorrectCase(name, false)
	_, ok := profile.isNil[correctCase]
	return ok
}

// SortedNames gets names in sorted order
func (profile *Profile) SortedNames(includeNil bool) []string {
	keys := make([]string, 0, len(profile.env)+len(profile.isNil))
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
	for k, v := range profile.rightCase {
		p.rightCase[k] = v
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

// IsMerged checkes if profile has already been merged
func (profile *Profile) IsMerged(p *Profile) bool {
	for k, v := range p.env {
		value, exists := profile.Get(k)
		if !exists || value != v {
			return false
		}
	}
	for k := range p.isNil {
		_, exists := profile.Get(k)
		if exists {
			return false
		}
	}
	return true
}

// Diff returns two lists, changed and removed in profile
func (profile *Profile) Diff(p *Profile) ([]string, []string) {
	changed := make([]string, 0)
	removed := make([]string, 0)
	for k, v := range p.env {
		value, exists := profile.Get(k)
		if !exists || value != v {
			changed = append(changed, k)
		}
	}
	for k := range p.isNil {
		_, exists := profile.Get(k)
		if exists {
			removed = append(removed, k)
		}
	}
	return changed, removed
}

// MergeStrings parses a list of NAME=value text entries and adds to profile.
func (profile *Profile) MergeStrings(envList []string) {
	for _, kv := range envList {
		pair := strings.SplitN(kv, "=", 2)
		if pair[0] != "" {
			// Ignoring extra environ variables with no names (Windows)
			if len(pair) == 2 {
				profile.Set(pair[0], pair[1])
			} else {
				profile.SetNil(pair[0])
			}
		}
	}
}

func (profile *Profile) String() string {
	return strings.Join(profile.SortedNames(true), ",")
}

func (profiles *Profiles) FindProfile(name string) (profile *Profile, found bool) {
	var result *Profile
	for profileName, profile := range *profiles {
		if profileName == name {
			result = &profile
			break
		}
	}
	return result, result != nil
}
