package source

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

type sourceMatcher struct {
	pattern  string
	resolved ResolvedSource
}

type sourceIndex struct {
	explicitExact    map[string]ResolvedSource
	explicitFold     []sourceMatcher
	explicitContains []sourceMatcher
	mediaTypes       map[models.MediaType]ResolvedSource
	tags             []sourceMatcher
	urls             []sourceMatcher
	ptbrKinds        map[SourceKind]struct{}
	definitions      map[SourceKind]SourceDefinition
	shortIDResolved  ResolvedSource
	hasShortID       bool
}

var (
	unknownEmptyURL    = ResolvedSource{Kind: Unknown, Name: string(Unknown), Reason: "empty URL"}
	unknownURLNotMatch = ResolvedSource{Kind: Unknown, Name: string(Unknown), Reason: "URL not matched"}
	resolverIndex      = buildSourceIndex()
)

// ResolvedSource is the immutable result of source resolution.
type ResolvedSource struct {
	Kind   SourceKind
	Name   string
	Reason string
}

// Apply normalizes the source field when a source was resolved.
func (r ResolvedSource) Apply(anime *models.Anime) {
	if anime == nil || r.Name == "" {
		return
	}

	anime.Source = r.Name
}

// Resolve determines the canonical source for an anime.
// Precedence is fixed and intentional:
//  1. Explicit anime.Source field
//  2. anime.MediaType
//  3. Tags in anime.Name
//  4. URL pattern / short ID
func Resolve(anime *models.Anime) (ResolvedSource, error) {
	if anime == nil {
		return ResolvedSource{}, fmt.Errorf("cannot resolve source for nil anime")
	}

	if resolved, ok := resolveSourceFromExplicit(anime.Source); ok {
		return resolved, nil
	}

	if resolved, ok := resolveSourceFromMediaType(anime.MediaType); ok {
		return resolved, nil
	}

	if resolved, handled, err := resolveSourceFromTags(anime.Name, anime.URL); handled {
		if err != nil {
			return ResolvedSource{}, err
		}
		return resolved, nil
	}

	if resolved := ResolveURL(anime.URL); resolved.Kind != Unknown {
		return resolved, nil
	}

	util.Warn("source resolution failed", "anime", anime.Name, "source", anime.Source, "url", anime.URL)
	return ResolvedSource{}, fmt.Errorf(
		"could not resolve source for %q (source=%q mediaType=%q url=%q)",
		anime.Name,
		anime.Source,
		anime.MediaType,
		anime.URL,
	)
}

// ResolveURL resolves a source from a raw URL string only.
func ResolveURL(rawURL string) ResolvedSource {
	if rawURL == "" {
		return unknownEmptyURL
	}

	for _, matcher := range resolverIndex.urls {
		if containsASCIIFold(rawURL, matcher.pattern) {
			return matcher.resolved
		}
	}

	if resolverIndex.hasShortID && IsAllAnimeShortID(rawURL) {
		return resolverIndex.shortIDResolved
	}

	return unknownURLNotMatch
}

func newResolvedSource(kind SourceKind, reason string) ResolvedSource {
	return ResolvedSource{
		Kind:   kind,
		Name:   string(kind),
		Reason: reason,
	}
}

func resolveSourceFromExplicit(explicit string) (ResolvedSource, bool) {
	explicit = strings.TrimSpace(explicit)
	if explicit == "" {
		return ResolvedSource{}, false
	}

	if resolved, ok := resolverIndex.explicitExact[explicit]; ok {
		return resolved, true
	}

	for _, matcher := range resolverIndex.explicitFold {
		if equalASCIIFold(explicit, matcher.pattern) {
			return matcher.resolved, true
		}
	}

	for _, matcher := range resolverIndex.explicitContains {
		if containsASCIIFold(explicit, matcher.pattern) {
			return matcher.resolved, true
		}
	}

	return ResolvedSource{}, false
}

func resolveSourceFromMediaType(mediaType models.MediaType) (ResolvedSource, bool) {
	if mediaType == "" {
		return ResolvedSource{}, false
	}

	if resolved, ok := resolverIndex.mediaTypes[mediaType]; ok {
		return resolved, true
	}

	return ResolvedSource{}, false
}

func resolveSourceFromTags(name, rawURL string) (ResolvedSource, bool, error) {
	if strings.TrimSpace(name) == "" {
		return ResolvedSource{}, false, nil
	}

	for _, matcher := range resolverIndex.tags {
		if containsASCIIFold(name, matcher.pattern) {
			return matcher.resolved, true, nil
		}
	}

	if hasPTBRTag(name) {
		if resolved, ok := resolvePTBRSource(rawURL); ok {
			resolved.Reason = "PT-BR tag + URL"
			return resolved, true, nil
		}

		return ResolvedSource{}, true, fmt.Errorf(
			"could not resolve PT-BR source for %q without an explicit source or recognizable URL",
			name,
		)
	}

	return ResolvedSource{}, false, nil
}

func resolvePTBRSource(rawURL string) (ResolvedSource, bool) {
	resolved := ResolveURL(rawURL)
	if resolved.Kind == Unknown {
		return ResolvedSource{}, false
	}

	if _, ok := resolverIndex.ptbrKinds[resolved.Kind]; ok {
		return resolved, true
	}

	return ResolvedSource{}, false
}

// IsAllAnimeShortID returns true if s looks like an AllAnime short ID.
func IsAllAnimeShortID(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || len(value) >= 30 {
		return false
	}

	hasLetter := false
	numericLike := true
	dotCount := 0
	for i := 0; i < len(value); i++ {
		c := value[i]
		if c == '/' || c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '\f' || c == '\v' {
			return false
		}
		if asciiLowerAt(value, i) == 'h' && hasASCIIFoldPrefix(value[i:], "http") {
			return false
		}
		switch {
		case c >= '0' && c <= '9':
		case c == '.':
			dotCount++
			if dotCount > 1 {
				numericLike = false
			}
		case isASCIILetter(c):
			hasLetter = true
			numericLike = false
		default:
			numericLike = false
		}
	}

	return hasLetter && !numericLike
}

// ExtractAllAnimeID extracts the AllAnime ID from a URL or bare string.
func ExtractAllAnimeID(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}

	if IsAllAnimeShortID(value) {
		return value
	}

	if idx := indexASCIIFold(value, "/anime/"); idx >= 0 {
		start := idx + len("/anime/")
		end := start
		for end < len(value) && value[end] != '/' {
			end++
		}
		candidate := value[start:end]
		if IsAllAnimeShortID(candidate) {
			return candidate
		}
	}

	if containsASCIIFold(value, "allanime") {
		for start := 0; start < len(value); {
			for start < len(value) && value[start] == '/' {
				start++
			}
			if start >= len(value) {
				break
			}

			end := start
			for end < len(value) && value[end] != '/' {
				end++
			}

			part := value[start:end]
			if strings.IndexByte(part, '.') >= 0 {
				start = end + 1
				continue
			}
			if IsAllAnimeShortID(part) {
				return part
			}
			start = end + 1
		}
	}

	return value
}

func hasPTBRTag(name string) bool {
	return containsASCIIFold(name, "[pt-br]") ||
		containsASCIIFold(name, "[portugu")
}

func buildSourceIndex() sourceIndex {
	index := sourceIndex{
		explicitExact: make(map[string]ResolvedSource),
		mediaTypes:    make(map[models.MediaType]ResolvedSource),
		ptbrKinds:     make(map[SourceKind]struct{}),
		definitions:   make(map[SourceKind]SourceDefinition),
	}

	for _, def := range sourceDefs {
		index.definitions[def.Kind] = def

		explicitResolved := newResolvedSource(def.Kind, "explicit source")
		tagResolved := newResolvedSource(def.Kind, "tag")

		for _, explicit := range def.Explicit {
			if _, ok := index.explicitExact[explicit]; !ok {
				index.explicitExact[explicit] = explicitResolved
			}
			index.explicitFold = append(index.explicitFold, sourceMatcher{
				pattern:  strings.ToLower(explicit),
				resolved: explicitResolved,
			})
		}

		for _, pattern := range def.ExplicitContains {
			index.explicitContains = append(index.explicitContains, sourceMatcher{
				pattern:  pattern,
				resolved: explicitResolved,
			})
		}

		for _, mediaType := range def.MediaTypes {
			if _, ok := index.mediaTypes[mediaType]; !ok {
				index.mediaTypes[mediaType] = newResolvedSource(def.Kind, "media type")
			}
		}

		for _, tag := range def.Tags {
			index.tags = append(index.tags, sourceMatcher{
				pattern:  tag,
				resolved: tagResolved,
			})
		}

		for _, pattern := range def.URLMatchers {
			index.urls = append(index.urls, sourceMatcher{
				pattern:  pattern,
				resolved: newResolvedSource(def.Kind, "URL contains "+pattern),
			})
		}

		if def.ShortID {
			index.shortIDResolved = newResolvedSource(def.Kind, "short ID")
			index.hasShortID = true
		}

		if def.PTBR {
			index.ptbrKinds[def.Kind] = struct{}{}
		}
	}

	return index
}

func indexASCIIFold(value, pattern string) int {
	if pattern == "" {
		return 0
	}
	if len(pattern) > len(value) {
		return -1
	}

	first := pattern[0]
	limit := len(value) - len(pattern)
	for i := 0; i <= limit; i++ {
		if asciiLowerAt(value, i) != first {
			continue
		}
		if hasASCIIFoldPrefix(value[i:], pattern) {
			return i
		}
	}

	return -1
}

func containsASCIIFold(value, pattern string) bool {
	return indexASCIIFold(value, pattern) >= 0
}

func equalASCIIFold(value, pattern string) bool {
	return len(value) == len(pattern) && hasASCIIFoldPrefix(value, pattern)
}

func hasASCIIFoldPrefix(value, pattern string) bool {
	if len(pattern) > len(value) {
		return false
	}

	for i := 0; i < len(pattern); i++ {
		if asciiLowerAt(value, i) != pattern[i] {
			return false
		}
	}

	return true
}

func asciiLowerAt(value string, index int) byte {
	if index >= len(value) {
		return 0
	}
	return asciiLowerByte(value[index])
}

func asciiLowerByte(value byte) byte {
	if value >= 'A' && value <= 'Z' {
		return value + ('a' - 'A')
	}
	return value
}

func isASCIILetter(value byte) bool {
	return (value >= 'a' && value <= 'z') || (value >= 'A' && value <= 'Z')
}
