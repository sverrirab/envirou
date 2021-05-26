package util

import "strings"

type Pattern string
type Patterns []Pattern

// ParsePatterns parses a string with patterns
func ParsePatterns(s string) *Patterns {
	patterns := make(Patterns, 0, 8)
	for _, p := range strings.Split(s, ",") {
		trimmed := strings.TrimSpace(p)
		if len(trimmed) > 0 {
			patterns = append(patterns, Pattern(trimmed))
		}
	}
	return &patterns
}

// MatchAny Match any of the patterns
func MatchAny(s string, patterns *Patterns) bool {
	for _, p := range *patterns {
		if Match(s, p) {
			return true
		}
	}
	return false
}

// Match Simple glob macher where pattern can be *PATTERN, *PATTERN* or PATTERN*
func Match(s string, p Pattern) bool {
	pattern := string(p)
	if pattern == "" {
		return false
	} else if pattern == "*" {
		return true
	}
	first_char := pattern[0]
	last_char_pos := len(pattern) - 1
	last_char := pattern[last_char_pos]
	if first_char == '*' && last_char == '*' {
		if strings.Contains(s, pattern[1:last_char_pos]) {
			return true
		}
	} else if last_char == '*' {
		if strings.HasPrefix(s, pattern[0:last_char_pos]) {
			return true
		}
	} else if first_char == '*' {
		if strings.HasSuffix(s, pattern[1:last_char_pos+1]) {
			return true
		}
	} else if s == pattern {
		return true
	}
	return false
}
