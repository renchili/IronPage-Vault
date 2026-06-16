package core

import (
	"strings"
	"unicode"
)

func ExtractMentionUsernames(text string) []string {
	found := map[string]bool{}
	out := []string{}
	fields := strings.Fields(text)

	for _, f := range fields {
		if !strings.HasPrefix(f, "@") || len(f) < 2 {
			continue
		}

		name := strings.TrimFunc(strings.TrimPrefix(f, "@"), func(r rune) bool {
			return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.')
		})
		if name == "" || found[name] {
			continue
		}

		found[name] = true
		out = append(out, name)
	}

	return out
}
