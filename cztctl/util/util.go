package util

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// Title returns a string with the first letter upper case.
func Title(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// SafeString converts the input string to a valid Go identifier.
func SafeString(in string) string {
	if len(in) == 0 {
		return in
	}

	var builder strings.Builder
	for i, r := range in {
		if i == 0 && !unicode.IsLetter(r) && r != '_' {
			builder.WriteRune('_')
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// WriteIndent writes indent to writer.
func WriteIndent(writer io.Writer, indent int) {
	for i := 0; i < indent; i++ {
		fmt.Fprint(writer, "\t")
	}
}
