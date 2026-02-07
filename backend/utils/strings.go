package utils

import "strings"

func UpperCompact(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ToUpper(s)
	// remove internal whitespace too (your previous behavior)
	out := make([]rune, 0, len(s))
	for _, ch := range s {
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			out = append(out, ch)
		}
	}
	return string(out)
}

func IsIATACode(s string) bool {
	if len(s) != 3 {
		return false
	}
	for i := 0; i < 3; i++ {
		c := s[i]
		if c < 'A' || c > 'Z' {
			return false
		}
	}
	return true
}
