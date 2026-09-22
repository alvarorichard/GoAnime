package hianime

import (
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// The ZokoAnime embed ships its whole player config in one string.
//
// The page is a stub: a <div id="player">, a module script, and
//
//	window.__P="<base64>"
//
// The bytes behind that base64 are the config JSON XOR'd with a repeating
// ASCII key. The key is the player's own build tag, "otaku-embed-v1" — the
// suffix is why obfuscationKey is a named constant and why decodePayload
// reports a decode failure as "the player's obfuscation changed" rather than as
// malformed JSON: a v2 would land here first.
//
// Verified 2026-09-22 against four embeds (Naruto sub, Bleach sub and dub,
// Boruto hsub); all four decoded to well-formed JSON carrying an .m3u8 src.
const obfuscationKey = "otaku-embed-v1"

// payloadRe captures the base64 blob assigned to window.__P.
var payloadRe = regexp.MustCompile(`window\.__P\s*=\s*"([A-Za-z0-9+/=]+)"`)

// embedPayload is the part of the player config this scraper reads.
type embedPayload struct {
	Src       string          `json:"src"`
	Subtitles []embedSubtitle `json:"subtitles"`
	Skip      struct {
		Intro *embedSkipRange `json:"intro"`
		Outro *embedSkipRange `json:"outro"`
	} `json:"skip"`
}

type embedSubtitle struct {
	Lang    string `json:"lang"`
	Label   string `json:"label"`
	Src     string `json:"src"`
	Default bool   `json:"default"`
}

type embedSkipRange struct {
	Start int `json:"start"`
	End   int `json:"end"`
}

// decodePayload turns an embed page into its player config.
func decodePayload(page []byte) (*embedPayload, error) {
	m := payloadRe.FindSubmatch(page)
	if m == nil {
		return nil, netx.NewParserError(sourceLabel, "embed",
			"no player payload in the embed page (layout changed?)", nil)
	}
	raw, err := base64.StdEncoding.DecodeString(string(m[1]))
	if err != nil {
		return nil, netx.NewParserError(sourceLabel, "embed",
			"the player payload was not valid base64", err)
	}

	key := []byte(obfuscationKey)
	for i := range raw {
		raw[i] ^= key[i%len(key)]
	}

	var payload embedPayload
	if err := jsonx.Unmarshal(raw, &payload); err != nil {
		return nil, netx.NewParserError(sourceLabel, "embed",
			"the player's obfuscation changed; its config no longer decodes", err)
	}
	if strings.TrimSpace(payload.Src) == "" {
		return nil, netx.NewParserError(sourceLabel, "embed",
			"the player config carried no stream URL", nil)
	}
	return &payload, nil
}
