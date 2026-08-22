package util

import "strings"

// legacyStrip9AnimeParenMeta is the pre-Go-1.27 implementation of
// strip9AnimeParenMeta, kept verbatim as the oracle for the differential test
// in cutlast_paren_test.go. Do not "modernize" this copy.
func legacyStrip9AnimeParenMeta(name string) string {
	idx := strings.LastIndex(name, " (")
	if idx <= 0 {
		return name
	}
	suffix := name[idx:]
	closeIdx := strings.LastIndex(suffix, ")")
	if closeIdx < 0 || closeIdx != len(suffix)-1 {
		return name
	}
	candidate := strings.ToUpper(suffix)
	isMetadata := strings.Contains(candidate, "SUB") ||
		strings.Contains(candidate, "DUB") ||
		strings.Contains(candidate, "HD") ||
		strings.Contains(candidate, "MULTILANGUAGE") ||
		strings.Contains(candidate, "MULTI") ||
		strings.Contains(candidate, "EP ")
	if isMetadata {
		return strings.TrimSpace(name[:idx])
	}
	return name
}
