package util

import (
	"fmt"
	"strings"
	"unicode"
)

// LogSafe renders v for a log line with every line break and control character
// removed. Titles, episode names and error messages that come from a source are
// remote input: a name containing "\n" would otherwise start a line the program
// never wrote, and an escape sequence would reach the terminal the log is
// tailed in.
func LogSafe(v any) string {
	text := fmt.Sprint(v)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, text)
}
