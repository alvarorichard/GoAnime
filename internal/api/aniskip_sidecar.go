package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
)

// WriteAniSkipSidecar writes a "<video>.skips.json" file next to a downloaded
// episode so a player can skip the opening/ending without re-querying AniSkip.
//
// It used to live in the AllAnime-only smart-range downloader and hardcoded
// "AllAnime" as the sidecar's Source. Nothing about it is source-specific, so
// it survived that source's removal. models.Episode carries no source field,
// so the sidecar now records the skip-time provider instead of guessing.
func WriteAniSkipSidecar(videoPath string, ep *models.Episode) error {
	if ep == nil {
		return nil
	}
	// Nothing worth writing when no skip window is known.
	if ep.SkipTimes.Op.Start == 0 && ep.SkipTimes.Op.End == 0 &&
		ep.SkipTimes.Ed.Start == 0 && ep.SkipTimes.Ed.End == 0 {
		return nil
	}

	type skipFile struct {
		Format  string `json:"format"`
		OPStart int    `json:"op_start"`
		OPEnd   int    `json:"op_end"`
		EDStart int    `json:"ed_start"`
		EDEnd   int    `json:"ed_end"`
		Updated string `json:"updated"`
		Episode string `json:"episode"`
		Source  string `json:"source"`
	}

	payload := skipFile{
		Format:  "aniskip",
		OPStart: ep.SkipTimes.Op.Start,
		OPEnd:   ep.SkipTimes.Op.End,
		EDStart: ep.SkipTimes.Ed.Start,
		EDEnd:   ep.SkipTimes.Ed.End,
		Updated: time.Now().Format(time.RFC3339),
		Episode: ep.Number,
		Source:  "aniskip",
	}

	b, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	sidecar := strings.TrimSuffix(videoPath, filepath.Ext(videoPath)) + ".skips.json"
	// Restrictive permissions: owner read/write only.
	return os.WriteFile(sidecar, b, 0o600)
}
