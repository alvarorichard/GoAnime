// Package api provides enhanced episode URL fetching with AllAnime navigation support
package api

import (
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
)

// isAllAnimeSourceAPI reports whether anime is an AllAnime source. The
// navigation-aware stream path that used to live here (GetEpisodeStreamURLEnhanced
// / GetAllAnimeEpisodeURLDirect) moved into the AllAnime provider in Etapa 6.3;
// this predicate stays because allanime_smart.go's range download still needs it.
func isAllAnimeSourceAPI(anime *models.Anime) bool {
	if anime.Source == "AllAnime" {
		return true
	}

	if strings.Contains(anime.URL, "allanime") {
		return true
	}

	if len(anime.URL) < 30 &&
		strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") &&
		!strings.Contains(anime.URL, "http") {
		return true
	}

	return false
}
