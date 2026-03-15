package data

import (
	"os"
	"sort"
	"strings"
)

// MergeMode indicates how a variable should be merged.
const (
	MergeReplace = iota // Default: replace entire value
	MergePrepend        // Prepend to path-like variable
	MergeAppend         // Append to path-like variable
)

type Profile struct {
	env             map[string]string // The actual key value pair
	rightCase       map[string]string // VAR -> Var (maps to actual case, used for case insensitive comparison on Windows)
	isNil           map[string]bool   // True if item is to be removed (uses UPPER case name)
	mergeMode       map[string]int    // MergeReplace, MergePrepend, or MergeAppend per variable
	caseInsensitive bool
}
type Profiles map[string]Profile

func NewProfile(caseInsensitive bool) *Profile {
	return &Profile{
		env:             make(map[string]string),
		rightCase:       make(map[string]string),
		isNil:           make(map[string]bool),
		mergeMode:       make(map[string]int),
		caseInsensitive: caseInsensitive,
	}
}

func (profile *Profile) GetCorrectCase(name string, create bool) string {
	if profile.caseInsensitive {
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

// Set will set an entry (replace mode)
func (profile *Profile) Set(name string, value string) {
	profile.SetWithMode(name, value, MergeReplace)
}

// SetWithMode sets an entry with a specific merge mode.
func (profile *Profile) SetWithMode(name string, value string, mode int) {
	correctCase := profile.GetCorrectCase(name, true)
	profile.env[correctCase] = value
	profile.mergeMode[correctCase] = mode
	delete(profile.isNil, correctCase)
}

// GetMergeMode returns the merge mode for a variable.
func (profile *Profile) GetMergeMode(name string) int {
	correctCase := profile.GetCorrectCase(name, false)
	if mode, ok := profile.mergeMode[correctCase]; ok {
		return mode
	}
	return MergeReplace
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
	p := NewProfile(profile.caseInsensitive)
	for k, v := range profile.env {
		p.env[k] = v
	}
	for k, v := range profile.rightCase {
		p.rightCase[k] = v
	}
	for k, v := range profile.isNil {
		p.isNil[k] = v
	}
	for k, v := range profile.mergeMode {
		p.mergeMode[k] = v
	}
	return p
}

// MergeResult describes what happened during a merge.
type MergeResult struct {
	// PathSkipped lists variable names where all components already existed
	// but not in the expected position (prepend/append was a no-op).
	PathSkipped []string
}

// Merge applies all elements from p into current profile.
// For prepend/append modes, path components are split on os.PathListSeparator.
// If a component already exists anywhere in the current value, it is skipped (no-op).
func (profile *Profile) Merge(p *Profile) MergeResult {
	result := MergeResult{}
	sep := string(os.PathListSeparator)
	for k, v := range p.env {
		mode := p.GetMergeMode(k)
		switch mode {
		case MergePrepend, MergeAppend:
			existing, _ := profile.Get(k)
			merged, allSkipped := mergePathComponents(existing, v, sep, mode)
			profile.Set(k, merged)
			if allSkipped && existing != "" {
				result.PathSkipped = append(result.PathSkipped, k)
			}
		default:
			profile.Set(k, v)
		}
	}
	for k := range p.isNil {
		profile.SetNil(k)
	}
	return result
}

// mergePathComponents merges new path components into an existing path-like value.
// Components already present anywhere in existing are skipped.
// Returns the merged string and whether all components were already present (allSkipped).
func mergePathComponents(existing, addition, sep string, mode int) (string, bool) {
	existingParts := splitPath(existing, sep)
	newParts := splitPath(addition, sep)

	// Build a set of existing components for fast lookup
	existingSet := make(map[string]bool, len(existingParts))
	for _, p := range existingParts {
		existingSet[p] = true
	}

	// Filter out components that already exist
	toAdd := make([]string, 0, len(newParts))
	for _, p := range newParts {
		if !existingSet[p] {
			toAdd = append(toAdd, p)
		}
	}

	if len(toAdd) == 0 {
		return existing, true
	}

	switch mode {
	case MergePrepend:
		return strings.Join(append(toAdd, existingParts...), sep), false
	case MergeAppend:
		return strings.Join(append(existingParts, toAdd...), sep), false
	}
	return existing, false
}

// splitPath splits a path string, filtering out empty components.
func splitPath(s, sep string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// IsMerged checks if profile has already been merged.
// For prepend/append variables, checks if all components are present anywhere in the current value.
func (profile *Profile) IsMerged(p *Profile) bool {
	sep := string(os.PathListSeparator)
	for k, v := range p.env {
		value, exists := profile.Get(k)
		if !exists {
			return false
		}
		mode := p.GetMergeMode(k)
		switch mode {
		case MergePrepend, MergeAppend:
			if !pathContainsAll(value, v, sep) {
				return false
			}
		default:
			if value != v {
				return false
			}
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

// pathContainsAll returns true if all components of required are present in current.
func pathContainsAll(current, required, sep string) bool {
	currentParts := splitPath(current, sep)
	requiredParts := splitPath(required, sep)
	currentSet := make(map[string]bool, len(currentParts))
	for _, p := range currentParts {
		currentSet[p] = true
	}
	for _, p := range requiredParts {
		if !currentSet[p] {
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

// FullDiff compares current against snapshot and returns added, changed, and removed variable names (sorted).
func FullDiff(current, snapshot *Profile) (added, changed, removed []string) {
	added = make([]string, 0)
	changed = make([]string, 0)
	removed = make([]string, 0)
	for _, name := range current.SortedNames(false) {
		snapshotVal, inSnapshot := snapshot.Get(name)
		if !inSnapshot {
			added = append(added, name)
		} else if currentVal, _ := current.Get(name); currentVal != snapshotVal {
			changed = append(changed, name)
		}
	}
	for _, name := range snapshot.SortedNames(false) {
		if _, inCurrent := current.Get(name); !inCurrent {
			removed = append(removed, name)
		}
	}
	return added, changed, removed
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
