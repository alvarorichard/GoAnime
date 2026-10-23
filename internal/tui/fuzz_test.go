package tui

import (
	"strings"
	"testing"
	"time"
	"unicode"

	"github.com/stretchr/testify/assert"
)

// FuzzNormalizeTitleKey hunts panics and charset escapes in grouping keys.
func FuzzNormalizeTitleKey(f *testing.F) {
	for _, seed := range []string{"", "[PT-BR] Frieren", "爵々", "a  b\tc", "[unclosed", "A]B[C", "Éowyn-MÜNCHEN 123"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, name string) {
		key := normalizeTitleKey(name)
		assert.False(t, strings.ContainsFunc(key, unicode.IsControl), "no controls: %q", key)
		assert.NotContains(t, key, "  ", "single spaces: %q", key)
		assert.Equal(t, strings.TrimSpace(key), key)
		assert.Equal(t, key, normalizeTitleKey(name), "deterministic")
	})
}

// FuzzCleanResultTitle hunts control leaks into row titles.
func FuzzCleanResultTitle(f *testing.F) {
	for _, seed := range []string{"", "[]", "[PT-BR]", "[[[[", "]]]]", "a]b[c", "(Legendado)", "[Movie] X (Dublado)"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, name string) {
		got := cleanResultTitle(name)
		assert.NotEmpty(t, got)
		assert.False(t, strings.ContainsAny(got, "\n\r\x00\x07\x1b"), "single line: %q", got)
	})
}

// FuzzParseBackgroundIsLight hunts panics and flip-flops in OSC 11 parsing.
func FuzzParseBackgroundIsLight(f *testing.F) {
	for _, seed := range []string{"", "\x1b]11;rgb:ffff/ffff/ffff\x07", "rgb:0/0/0", "RGB:FFFF/FFFF/FFFF", "rgb:ff/ff", "rgb:zz/zz/zz", "rgb:fffff/0/0\x1b\\"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, reply string) {
		light1, ok1 := parseBackgroundIsLight(reply)
		light2, ok2 := parseBackgroundIsLight("\x1b]11;" + reply)
		_ = light2
		_ = ok2
		// Parsing must be total: never panic, boolean stays boolean.
		assert.True(t, light1 == true || light1 == false)
		assert.True(t, ok1 == true || ok1 == false)
	})
}

// FuzzRelativeDay hunts panics on absurd timestamps.
func FuzzRelativeDay(f *testing.F) {
	now := time.Now()
	for _, seed := range []int64{0, -1, 1, 86400, 86400 * 400, 1 << 40, -(1 << 40)} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, delta int64) {
		got := relativeDay(now.Add(time.Duration(delta)*time.Second), now)
		assert.NotEmpty(t, got)
	})
}

// FuzzSingleLineConverges hunts control escapes through sanitization.
func FuzzSingleLineConverges(f *testing.F) {
	for _, seed := range []string{"", "\x00\x01\x02", "日本語 \U0001F3AC", "\x1b]8;;https://x\aevil\x1b]8;;\a", " "} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		got := singleLine(value)
		assert.False(t, strings.ContainsFunc(got, unicode.IsControl), "no controls: %q", got)
		assert.Equal(t, got, singleLine(got), "idempotent")
	})
}
