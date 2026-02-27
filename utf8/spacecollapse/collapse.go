//go:build !solution

package spacecollapse

import (
	"strings"
	"unicode"
)


func CollapseSpaces(input string) string {
	var builder strings.Builder
	builder.Grow(len(input))

	isInSpaces := false
	for _, r := range input {
		if unicode.IsSpace(r) {
			if isInSpaces {
				continue
			}
			builder.WriteRune(' ')
			isInSpaces = true
			continue
		}
		builder.WriteRune(r)
		isInSpaces = false
	}
	return builder.String()
}
