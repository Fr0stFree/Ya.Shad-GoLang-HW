//go:build !solution

package ciletters

import (
	_ "embed"
	"strings"
	"text/template"
)

//go:embed letter.tmpl
var letterTemplate string

func MakeLetter(n *Notification) (string, error) {
	functions := template.FuncMap{"trim": trim, "last": last}
	templ, err := template.New("letter").Funcs(functions).Parse(letterTemplate)
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.Grow(1024)
	err = templ.Execute(&builder, n)
	if err != nil {
		return "", err
	}

	return builder.String(), nil
}

func trim(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func last(s string, n int) string {
	const prefix = "            "
	trimmed := strings.TrimRight(s, "\n")
	if trimmed == "" {
		return ""
	}
	lines := strings.Split(trimmed, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	for i := range lines {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n") + "\n"
}