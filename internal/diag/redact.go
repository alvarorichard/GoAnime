// Package diag collects what a maintainer needs to understand a bug report —
// the machine, the terminal, the tools, the providers — and removes what must
// not leave the user's computer. It is a leaf package: callers hand it what
// only they know (versions, sources) and it never imports the app.
package diag

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const redacted = "[redacted]"

var (
	// ?token=abc&sig=... and the same in prose ("failed with token=abc") —
	// judged by the parameter's name, so a search for "keyword" or "q" is left
	// alone.
	paramRe = regexp.MustCompile(`(^|[?&;\s,("'{])([A-Za-z][A-Za-z0-9_.-]*)=([^&\s#"'<>;,)?]*)`)

	// A value ends at "?" so that a URL passed as a value ("url=https://h/p?pwd=x")
	// does not hide the parameters inside it from the next match.

	// "token":"abc" — the same secrets as JSON members.
	jsonRe = regexp.MustCompile(`(?i)("[^"]*(?:token|secret|password|passwd|ticket|signature|cookie|authorization|apikey|api_key|clearance)[^"]*"\s*:\s*)"[^"]*"`)

	// Authorization: ..., Cookie: ..., and their key=value forms.
	headerRe = regexp.MustCompile(`(?i)\b(authorization|proxy-authorization|cookie|set-cookie|x-api-key)(\s*[:=]\s*)[^\r\n]+`)

	bearerRe = regexp.MustCompile(`(?i)\bbearer\s+[A-Za-z0-9._~+/=-]{8,}`)

	// user:password@host in a URL.
	userinfoRe = regexp.MustCompile(`(?i)\b([a-z][a-z0-9+.-]*://)[^/\s:@]+:[^/\s@]+@`)

	// The one-shot tickets the RedeFlix player hands out: v1.<id>.<id>.<sig>.
	ticketRe = regexp.MustCompile(`\bv1\.[A-Za-z0-9_-]{6,}(?:\.[A-Za-z0-9_-]{4,}){1,3}`)

	jwtRe = regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{4,}`)

	clearanceRe = regexp.MustCompile(`(?i)\bcf_clearance=[^;\s"']+`)
)

// sensitiveWords make a parameter name secret wherever they appear in it.
var sensitiveWords = []string{
	"token", "secret", "signature", "ticket", "password", "passwd", "credential",
	"clearance", "session", "authorization", "apikey", "api_key", "api-key",
}

// sensitiveNames are secret only as a whole name ("key" is, "keyword" is not).
var sensitiveNames = map[string]bool{
	"key": true, "sig": true, "sid": true, "tsv": true, "auth": true, "pass": true, "pwd": true,
}

func sensitiveParam(name string) bool {
	n := strings.ToLower(name)
	if sensitiveNames[n] {
		return true
	}
	for _, w := range sensitiveWords {
		if strings.Contains(n, w) {
			return true
		}
	}
	return false
}

// triggers are the lowercase fragments without which no rule below can match.
// Built from the same word lists the rules use, so they cannot drift apart: a
// line with none of them skips every regular expression (the common case, and
// the difference between about 75µs and a few µs per log line).
var triggers = func() []string {
	t := []string{"authorization", "cookie", "x-api-key", "bearer", "v1.", "eyj", "clearance", "@"}
	t = append(t, sensitiveWords...)
	for n := range sensitiveNames {
		t = append(t, n)
	}
	return t
}()

func mayHoldASecret(s string) bool {
	lower := strings.ToLower(s)
	for _, t := range triggers {
		if strings.Contains(lower, t) {
			return true
		}
	}
	return false
}

// Redact removes credentials, tickets and the user's home directory from s.
// It is safe to apply twice.
func Redact(s string) string {
	if s == "" {
		return s
	}
	if !mayHoldASecret(s) {
		return RedactHome(s, homeDir())
	}
	return redactEverything(s)
}

// redactEverything applies every rule with no shortcut. Redact must give the
// same answer; TestFastPathAgreesWithEveryRule holds it to that.
func redactEverything(s string) string {
	s = userinfoRe.ReplaceAllString(s, "${1}"+redacted+"@")
	s = paramRe.ReplaceAllStringFunc(s, func(m string) string {
		parts := paramRe.FindStringSubmatch(m)
		if sensitiveParam(parts[2]) {
			return parts[1] + parts[2] + "=" + redacted
		}
		return m
	})
	s = jsonRe.ReplaceAllString(s, `${1}"`+redacted+`"`)
	s = headerRe.ReplaceAllString(s, "${1}${2}"+redacted)
	s = bearerRe.ReplaceAllString(s, "Bearer "+redacted)
	s = clearanceRe.ReplaceAllString(s, "cf_clearance="+redacted)
	s = jwtRe.ReplaceAllString(s, redacted)
	s = ticketRe.ReplaceAllString(s, "v1."+redacted)
	return RedactHome(s, homeDir())
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// RedactHome replaces the user's home directory with "~", so a shared log does
// not carry the account name. A home of "/" or "" is left alone: replacing it
// would mangle every path.
func RedactHome(s, home string) string {
	home = strings.TrimRight(home, `/\`)
	if len(home) < 2 {
		return s
	}
	for _, h := range []string{home, filepath.ToSlash(home), strings.ReplaceAll(home, `\`, `\\`)} {
		s = replaceFold(s, h, "~")
	}
	return s
}

// replaceFold replaces every occurrence of old, ignoring ASCII case: Windows
// paths differ in case between APIs, and an account name in a path is ASCII in
// practice. It compares bytes in place, so nothing can change length — the
// previous version lowercased both strings and indexed one with offsets from
// the other, which panics for "İ" and other runes whose lowercase is a
// different size.
func replaceFold(s, old, repl string) string {
	if old == "" {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		if hasPrefixASCIIFold(s[i:], old) {
			b.WriteString(repl)
			i += len(old)
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func hasPrefixASCIIFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		a, b := s[i], prefix[i]
		if 'A' <= a && a <= 'Z' {
			a += 'a' - 'A'
		}
		if 'A' <= b && b <= 'Z' {
			b += 'a' - 'A'
		}
		if a != b {
			return false
		}
	}
	return true
}
