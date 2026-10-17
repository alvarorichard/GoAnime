// Package models contains data structures for media content.
package models

import (
	"strings"
	"unicode"
)

// accentFold replaces common Portuguese/Spanish/French accented runes with
// their ASCII base. It covers what catalog titles actually contain; anything
// else passes through unchanged.
func accentFold(s string) string {
	replacer := strings.NewReplacer(
		"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
		"é", "e", "è", "e", "ê", "e", "ë", "e",
		"í", "i", "ì", "i", "î", "i", "ï", "i",
		"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
		"ú", "u", "ù", "u", "û", "u", "ü", "u",
		"ç", "c", "ñ", "n",
		"Á", "a", "À", "a", "Â", "a", "Ã", "a", "Ä", "a",
		"É", "e", "È", "e", "Ê", "e", "Ë", "e",
		"Í", "i", "Ì", "i", "Î", "i", "Ï", "i",
		"Ó", "o", "Ò", "o", "Ô", "o", "Õ", "o", "Ö", "o",
		"Ú", "u", "Ù", "u", "Û", "u", "Ü", "u",
		"Ç", "c", "Ñ", "n",
	)
	return replacer.Replace(s)
}

// normalizeForMatch lowercases, folds accents, strips bracket tags like
// "[Movie] [PT-BR]" and parenthesized audio tags, and keeps only letters,
// digits and single spaces.
func normalizeForMatch(s string) string {
	s = strings.ToLower(accentFold(s))
	// Strip bracketed tags: [movie], [pt-br], [tv], etc.
	for {
		start := strings.Index(s, "[")
		end := strings.Index(s, "]")
		if start < 0 || end < 0 || end < start {
			break
		}
		s = s[:start] + " " + s[end+1:]
	}
	// Strip trailing audio tags the scrapers append.
	for _, tag := range []string{"(dublado)", "(legendado)"} {
		s = strings.ReplaceAll(s, tag, " ")
	}
	var sb strings.Builder
	space := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			sb.WriteRune(r)
			space = false
		case unicode.IsSpace(r) || r == '-' || r == '_' || r == ':' || r == '.' || r == ',' || r == '\'' || r == '"' || r == '!' || r == '?' || r == '(' || r == ')' || r == '&':
			if !space {
				sb.WriteRune(' ')
				space = true
			}
		default:
			if !space {
				sb.WriteRune(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(strings.Join(strings.Fields(sb.String()), " "))
}

// queryStopWords are Portuguese/English glue words that must not decide a
// match on their own. "a" matching "a casa morbida" must not outrank a real
// title match. Content words ("todo", "poderoso") are never stopwords.
var queryStopWords = map[string]bool{
	"a": true, "o": true, "os": true, "as": true,
	"de": true, "do": true, "da": true, "dos": true, "das": true,
	"e": true, "em": true, "no": true, "na": true, "nos": true, "nas": true,
	"um": true, "uma": true, "uns": true, "umas": true,
	"para": true, "com": true, "que": true, "the": true, "of": true,
}

// queryTokens splits a normalized query into significant tokens: single
// characters and stopwords are dropped, pure numbers (years) are kept
// separately so they never force a zero score on their own.
func queryTokens(normalized string) (words []string, years []string) {
	for _, t := range strings.Fields(normalized) {
		if t == "" {
			continue
		}
		if len(t) <= 1 || queryStopWords[t] {
			continue
		}
		isNum := true
		for _, r := range t {
			if r < '0' || r > '9' {
				isNum = false
				break
			}
		}
		if isNum {
			// Only years matter; other numbers (e.g. "3" in "sonic 3")
			// are content, but single digits were dropped above and
			// longer numbers are rare in titles — keep them as years so
			// they match against the Year field, not the title.
			years = append(years, t)
			continue
		}
		words = append(words, t)
	}
	return words, years
}

// tokenMatches reports whether a query word matches a title: exact token
// equality, or substring for words of length >= 3 (so "poderoso" matches
// "poderosos" but "eu" does not match "europa").
func tokenMatches(queryWord, titleToken string) bool {
	if queryWord == titleToken {
		return true
	}
	if len(queryWord) >= 3 && strings.Contains(titleToken, queryWord) {
		return true
	}
	if len(titleToken) >= 3 && strings.Contains(queryWord, titleToken) {
		return true
	}
	return false
}

// RelevanceScore rates how well a catalog title matches a free-text query,
// 0 (unrelated) to 100 (exact). Year is the catalog year ("2003", may be
// empty); a year in the query that matches it adds a small bonus, a
// conflicting year subtracts.
func RelevanceScore(query, title, year string) int {
	nq := normalizeForMatch(query)
	nt := normalizeForMatch(title)
	if nq == "" || nt == "" {
		return 0
	}
	if nt == nq {
		return 100
	}
	if strings.Contains(nt, nq) {
		return 80
	}
	qWords, qYears := queryTokens(nq)
	if len(qWords) == 0 {
		// Query was only stopwords/years (e.g. "a", "2003"): fall back to
		// substring so single-word searches still rank.
		if strings.Contains(nt, nq) || strings.Contains(nq, nt) {
			return 50
		}
		return 0
	}
	titleTokens := strings.Fields(nt)
	matched := 0
	for _, qw := range qWords {
		for _, tt := range titleTokens {
			if tokenMatches(qw, tt) {
				matched++
				break
			}
		}
	}
	if matched == 0 {
		return 0
	}
	// All query words found: 60. Partial: proportional.
	score := matched * 60 / len(qWords)
	if matched == len(qWords) {
		// Bonus when the words appear in order ("todo poderoso" in
		// "a volta do todo poderoso").
		joined := " " + nt + " "
		ordered := true
		pos := 0
		for _, qw := range qWords {
			idx := strings.Index(joined[pos:], qw)
			if idx < 0 {
				// Allow substring-in-token for ordering too.
				found := false
				for _, tt := range titleTokens {
					if tokenMatches(qw, tt) {
						found = true
						break
					}
				}
				if !found {
					ordered = false
					break
				}
				continue
			}
			pos += idx + len(qw)
		}
		if ordered {
			score += 10
		}
		if score > 75 {
			score = 75
		}
	}
	// Year agreement: a query year matching the catalog year confirms the
	// exact cut ("todo poderoso 2003" vs 2003); a conflict demotes sequels
	// and remakes below the requested year.
	if len(qYears) > 0 && len(year) >= 4 {
		for _, y := range qYears {
			if len(y) == 4 && strings.Contains(year, y) {
				score += 5
			} else if len(y) == 4 {
				score -= 10
			}
		}
	}
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

// MatchesQuery reports whether a catalog title is related to the query at
// all: at least one significant query word appears in the title, or the
// normalized strings contain each other.
func MatchesQuery(query, title, year string) bool {
	return RelevanceScore(query, title, year) > 0
}
