package data

import (
	"strings"
)

type Pattern string
type Patterns []Pattern

// ParsePatterns parses a string with patterns
func ParsePatterns(s string, caseInsensitive bool) *Patterns {
	patterns := make(Patterns, 0, 8)
	for _, p := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(p)
		if len(trimmed) > 0 {
			if caseInsensitive {
				trimmed = strings.ToUpper(trimmed)
			}
			patterns = append(patterns, Pattern(trimmed))
		}
	}
	return &patterns
}

// MatchAny Match any of the patterns
func MatchAny(s string, patterns *Patterns, caseInsensitive bool) bool {
	for _, p := range *patterns {
		if Match(s, p, caseInsensitive) {
			return true
		}
	}
	return false
}

// Match Simple glob matcher where pattern can be PATTERN, *PATTERN, *PATTERN* or PATTERN*
func Match(s string, p Pattern, caseInsensitive bool) bool {
	pattern := string(p)
	if caseInsensitive {
		s = strings.ToUpper(s)
	}
	if pattern == "" {
		return false
	} else if pattern == "*" {
		return true
	}
	firstChar := pattern[0]
	lastCharPos := len(pattern) - 1
	lastChar := pattern[lastCharPos]
	if firstChar == '*' && lastChar == '*' {
		if strings.Contains(s, pattern[1:lastCharPos]) {
			return true
		}
	} else if lastChar == '*' {
		if strings.HasPrefix(s, pattern[0:lastCharPos]) {
			return true
		}
	} else if firstChar == '*' {
		if strings.HasSuffix(s, pattern[1:lastCharPos+1]) {
			return true
		}
	} else if s == pattern {
		return true
	}
	return false
}
