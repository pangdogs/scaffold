package main

import (
	"bytes"
	"strings"
	"unicode"
)

func snake2Camel(s string) string {
	var buf bytes.Buffer
	upper := true
	for _, c := range s {
		if c == '_' {
			upper = true
			continue
		}
		if upper {
			buf.WriteRune(unicode.ToUpper(c))
			upper = false
		} else {
			buf.WriteRune(c)
		}
	}
	return strings.Title(buf.String())
}
