package util

import (
	"encoding/json"

	"github.com/alvarorichard/Goanime/internal/models"
)

// EncodeEpisodes encodes episodes to JSON bytes for caching
func EncodeEpisodes(episodes []models.Episode) ([]byte, error) {
	return json.Marshal(episodes)
}

// DecodeEpisodes decodes episodes from cached JSON bytes
func DecodeEpisodes(data []byte) []models.Episode {
	var episodes []models.Episode
	if err := json.Unmarshal(data, &episodes); err != nil {
		Debugf("Failed to decode episodes from cache: %v", err)
		return nil
	}
	return episodes
}
