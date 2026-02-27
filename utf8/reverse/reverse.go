//go:build !solution

package reverse

import (
	"strings"
	"unicode/utf8"
)

func Reverse(input string) string {
	if input == "" { return "" }

	var builder strings.Builder
	builder.Grow(len(input))

	for idx := len(input); idx > 0; {
		r, rSize := utf8.DecodeLastRuneInString(input[:idx])
		if r == utf8.RuneError {
			builder.WriteRune(utf8.RuneError)
		} else {
			builder.WriteString(input[idx-rSize:idx])
		}
		idx -= rSize
	}
	return builder.String()
}
