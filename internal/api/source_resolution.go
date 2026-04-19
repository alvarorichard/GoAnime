// Package api coordinates source resolution and playback-oriented orchestration.
package api

import (
	sourcepkg "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
)

type SourceKind = sourcepkg.SourceKind
type ResolvedSource = sourcepkg.ResolvedSource

const (
	SourceUnknown    SourceKind = sourcepkg.Unknown
	SourceAllAnime   SourceKind = sourcepkg.AllAnime
	SourceAnimefire  SourceKind = sourcepkg.AnimeFire
	SourceAnimeDrive SourceKind = sourcepkg.AnimeDrive
	SourceFlixHQ     SourceKind = sourcepkg.FlixHQ
	SourceSFlix      SourceKind = sourcepkg.SFlix
	SourceNineAnime  SourceKind = sourcepkg.NineAnime
	SourceGoyabu     SourceKind = sourcepkg.Goyabu
	SourceSuperFlix  SourceKind = sourcepkg.SuperFlix
)

func ResolveSource(anime *models.Anime) (ResolvedSource, error) {
	return sourcepkg.Resolve(anime)
}

func ResolveURL(url string) ResolvedSource {
	return sourcepkg.ResolveURL(url)
}

func IsAllAnimeSource(anime *models.Anime) bool {
	resolved, err := ResolveSource(anime)
	return err == nil && resolved.Kind == SourceAllAnime
}

func IsAllAnimeShortID(value string) bool {
	return sourcepkg.IsAllAnimeShortID(value)
}

func ExtractAllAnimeID(value string) string {
	return sourcepkg.ExtractAllAnimeID(value)
}
