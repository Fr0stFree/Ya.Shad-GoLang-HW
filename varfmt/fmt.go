//go:build !solution

package varfmt

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func Sprintf(format string, args ...interface{}) string {
	placeholderIdx := 0
	var builder strings.Builder
	builder.Grow(len(format))

	for idx := 0; idx < len(format); {
		r, _ := utf8.DecodeRuneInString(format[idx:])
		if r != '{' {
			builder.WriteRune(r)
			idx++
			continue
		}

		rIdx := idx + 1
		for format[rIdx] != '}' {
			rIdx++
		}
		subString := format[idx+1 : rIdx]
		if subString == "" {
			fmt.Fprint(&builder, args[placeholderIdx])
		} else {
			argIdx, _ := strconv.Atoi(subString)
			fmt.Fprint(&builder, args[argIdx])
		}
		placeholderIdx++
		idx = rIdx + 1

	}
	return builder.String()
}
