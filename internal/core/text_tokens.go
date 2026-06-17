package core

import (
	"strings"
	"unicode"
)

func ExtractMentionUsernames(text string) []string {
	found := map[string]bool{}
	out := []string{}
	runes := []rune(text)

	isName := func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.'
	}
	isBoundary := func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.')
	}

	for i := 0; i < len(runes); i++ {
		if runes[i] != '@' {
			continue
		}
		if i > 0 && !isBoundary(runes[i-1]) {
			continue
		}
		j := i + 1
		for j < len(runes) && isName(runes[j]) {
			j++
		}
		name := strings.TrimRight(string(runes[i+1:j]), ".")
		if name == "" || found[name] {
			continue
		}
		found[name] = true
		out = append(out, name)
		i = j - 1
	}

	return out
}
