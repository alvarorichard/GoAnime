package api

import (
	"fmt"
	"sort"
	"sync"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// SuperFlix offers, per episode, a list of servers, and each server is tagged
// Dublado (type 1) or Legendado (type 2). Those are the two questions a viewer
// actually has — "do I want the dub or the original with subtitles?" and, if there
// is more than one source for that choice, "which server?" — so we ask them as two
// clear, ordered steps rather than one flat list mixing both.
//
// The list is only reachable through the player page + /player/bootstrap. The
// embed page the browser sniff lands on carries no servers at all, which is why
// playback used to take whatever the embed happened to play, with the user
// choosing nothing. When the list is unavailable the caller falls back to choosing
// audio from the stream's own tracks (see selectSuperFlixAudio).

// sfPickFn is the single selection seam for every SuperFlix prompt. A package var
// so tests drive the whole flow without a TTY.
var sfPickFn = func(prompt string, labels []string) (int, error) {
	return tui.Find(labels, func(i int) string { return labels[i] },
		fuzzyfinder.WithPromptString(prompt))
}

// audioKindName is the site's own word for an audio type.
func audioKindName(kind int) string {
	switch kind {
	case superflix.SuperFlixAudioDubbed:
		return "Dublado"
	case superflix.SuperFlixAudioSubtitled:
		return "Legendado"
	default:
		return "Outro áudio"
	}
}

// audioKindLabel is the row shown in the audio step: the site's word plus a plain
// explanation, because "Dublado / Legendado" alone leaves newcomers guessing.
func audioKindLabel(kind int) string {
	switch kind {
	case superflix.SuperFlixAudioDubbed:
		return "🎙️  Dublado  ·  áudio em português"
	case superflix.SuperFlixAudioSubtitled:
		return "💬 Legendado  ·  áudio original com legendas"
	default:
		return "❔ Outro áudio"
	}
}

// audioKindRank orders the audio types: dubbed first (the default expectation for
// a PT-BR catalog), then subtitled, then anything unclassified.
func audioKindRank(kind int) int {
	switch kind {
	case superflix.SuperFlixAudioDubbed:
		return 0
	case superflix.SuperFlixAudioSubtitled:
		return 1
	default:
		return 2
	}
}

// orderedServers returns the playable servers in a stable, sensible order: by
// audio type (dubbed first), and within a type streaming servers before direct
// MP4 files. Deterministic ordering keeps the menu — and the "first option"
// fallback — stable between episodes.
func orderedServers(servers []superflix.SuperFlixServer) []superflix.SuperFlixServer {
	out := append([]superflix.SuperFlixServer(nil), servers...)
	sort.SliceStable(out, func(i, j int) bool {
		if ri, rj := audioKindRank(out[i].Type), audioKindRank(out[j].Type); ri != rj {
			return ri < rj
		}
		// Streaming before file, so index 0 is a plain streaming server.
		return !out[i].IsFile && out[j].IsFile
	})
	return out
}

// distinctAudioKinds lists the audio types present, in ranked order.
func distinctAudioKinds(servers []superflix.SuperFlixServer) []int {
	seen := map[int]bool{}
	var kinds []int
	for _, s := range servers {
		if !seen[s.Type] {
			seen[s.Type] = true
			kinds = append(kinds, s.Type)
		}
	}
	sort.SliceStable(kinds, func(i, j int) bool { return audioKindRank(kinds[i]) < audioKindRank(kinds[j]) })
	return kinds
}

// serversOfAudioKind filters to one audio type, preserving order.
func serversOfAudioKind(servers []superflix.SuperFlixServer, kind int) []superflix.SuperFlixServer {
	var out []superflix.SuperFlixServer
	for _, s := range servers {
		if s.Type == kind {
			out = append(out, s)
		}
	}
	return out
}

// serverRowLabel names a server for the server step.
//
// The site's own name is an opaque per-episode id ("Servidor 159462"), useless to
// a human, so we number the servers 1..N in the order shown and only surface the
// one distinction that means something: a direct MP4 (downloadable, usually the
// steadier source) versus a streaming server.
func serverRowLabel(s superflix.SuperFlixServer, position int) string {
	if s.IsFile {
		return fmt.Sprintf("⬇  Servidor %d  ·  MP4 (download direto)", position)
	}
	return fmt.Sprintf("▶  Servidor %d", position)
}

// sfServerPref is what we remember about a user's pick.
//
// It records the server's STABLE properties, not its name or id: a server is
// called "Servidor 159462" and that number changes every episode, so remembering
// the name would never match again. Type (dublado/legendado) and IsFile (direct
// MP4 vs streaming) carry across episodes, which is what lets a binge ask once and
// then honor the answer silently.
type sfServerPref struct {
	Type   int
	IsFile bool
}

var (
	sfServerPrefMu sync.Mutex
	sfServerPrefs  = map[string]sfServerPref{}
)

func rememberSuperFlixServer(tmdbID string, s superflix.SuperFlixServer) {
	sfServerPrefMu.Lock()
	defer sfServerPrefMu.Unlock()
	sfServerPrefs[tmdbID] = sfServerPref{Type: s.Type, IsFile: s.IsFile}
}

func recallSuperFlixServer(tmdbID string) (sfServerPref, bool) {
	sfServerPrefMu.Lock()
	defer sfServerPrefMu.Unlock()
	p, ok := sfServerPrefs[tmdbID]
	return p, ok
}

// resetSuperFlixServerPrefs clears the per-title memory. Tests only.
func resetSuperFlixServerPrefs() {
	sfServerPrefMu.Lock()
	defer sfServerPrefMu.Unlock()
	sfServerPrefs = map[string]sfServerPref{}
}

// narrowByMemory returns the servers matching a remembered preference, trying the
// exact match (same audio AND same kind) before relaxing to just the audio — so a
// title whose MP4 mirror vanished this episode still honors "dublado". Returns the
// full list unchanged when nothing matches (the memory is stale).
func narrowByMemory(servers []superflix.SuperFlixServer, pref sfServerPref) []superflix.SuperFlixServer {
	var exact, sameAudio []superflix.SuperFlixServer
	for _, s := range servers {
		if s.Type != pref.Type {
			continue
		}
		sameAudio = append(sameAudio, s)
		if s.IsFile == pref.IsFile {
			exact = append(exact, s)
		}
	}
	switch {
	case len(exact) > 0:
		return exact
	case len(sameAudio) > 0:
		return sameAudio
	default:
		return servers
	}
}

// pickOrDefault runs a prompt, defaulting to the first option when the picker
// cannot run (no TTY) or the user backs out — so playback never dead-ends on a
// selection step. Returns the chosen index.
func pickOrDefault(prompt string, labels []string) int {
	idx, err := sfPickFn(prompt, labels)
	if err != nil || idx < 0 || idx >= len(labels) {
		util.Debug("SuperFlix: picker unavailable, taking the first option", "prompt", prompt, "err", err)
		return 0
	}
	return idx
}

// selectSuperFlixServer chooses the server to stream from, in two intuitive steps
// and asking only when there is a genuine choice.
//
//	Step 1 — audio:  Dublado or Legendado, shown only when both exist.
//	Step 2 — server: which source, shown only when the chosen audio has more than one.
//
// A binge is asked once: the first episode's answer is remembered per title (by
// audio type + streaming/file), so later episodes resolve silently. A single
// server, or a memory that pins things down, skips the prompts entirely.
func selectSuperFlixServer(tmdbID string, servers []superflix.SuperFlixServer) (superflix.SuperFlixServer, error) {
	playable := orderedServers(servers)
	if len(playable) == 0 {
		return superflix.SuperFlixServer{}, fmt.Errorf("no servers to choose from")
	}
	if len(playable) == 1 {
		rememberSuperFlixServer(tmdbID, playable[0])
		return playable[0], nil
	}

	// Honor a remembered choice: it narrows the field, and often to exactly one.
	candidates := playable
	if pref, ok := recallSuperFlixServer(tmdbID); ok {
		candidates = narrowByMemory(playable, pref)
		if len(candidates) == 1 {
			util.Debug("SuperFlix server: honoring this title's remembered choice",
				"type", pref.Type, "isFile", pref.IsFile)
			return candidates[0], nil
		}
	}

	// Step 1 — audio kind. Only ask when the candidates span more than one.
	kinds := distinctAudioKinds(candidates)
	chosenKind := kinds[0]
	if len(kinds) > 1 {
		labels := make([]string, len(kinds))
		for i, k := range kinds {
			labels[i] = audioKindLabel(k)
		}
		chosenKind = kinds[pickOrDefault("Você quer assistir dublado ou legendado? ", labels)]
	}
	ofKind := serversOfAudioKind(candidates, chosenKind)

	// Step 2 — server. Only ask when the chosen audio has more than one source.
	chosen := ofKind[0]
	if len(ofKind) > 1 {
		labels := make([]string, len(ofKind))
		for i, s := range ofKind {
			labels[i] = serverRowLabel(s, i+1)
		}
		prompt := fmt.Sprintf("Escolha o servidor (%s): ", audioKindName(chosenKind))
		chosen = ofKind[pickOrDefault(prompt, labels)]
	}

	rememberSuperFlixServer(tmdbID, chosen)
	util.Debug("SuperFlix server chosen", "type", chosen.Type, "isFile", chosen.IsFile, "id", chosen.IDString())
	return chosen, nil
}

// rememberServerAudioChoice records, in the audio memory, the audio implied by a
// chosen server — so a later CACHED play of the title (which yields no server, and
// so falls to the audio-from-tracks picker) replays the same choice silently
// instead of prompting again.
//
// A dubbed server maps to the Portuguese track; a subtitled (or other) server maps
// to the zero option, whose empty Code makes mpvAudioLanguage return "" — i.e.
// keep the stream's own default audio, exactly as audioForServer does.
func rememberServerAudioChoice(tmdbID string, s superflix.SuperFlixServer) {
	opt := audioOption{}
	if s.Type == superflix.SuperFlixAudioDubbed {
		opt = audioOption{Code: "por", Dubbed: true}
	}
	rememberSuperFlixAudio(tmdbID, opt)
}

// audioForServer derives mpv's --alang from the server the user picked, so we
// never ask them the same question twice.
//
// A Dublado server means the viewer asked for the Portuguese dub, so prefer the
// Portuguese track. A Legendado server needs no preference at all: it IS the
// original-audio source, so whatever it defaults to is already right.
//
// We deliberately do NOT try to name the original language for a legendado
// server. Guessing "the first non-Portuguese track" picks by manifest order and
// gets it wrong on exactly the titles that need it most: Tehran's audio list is
// [por, und, eng, spa, kor, jpn, chi] and its original audio is Hebrew — carried
// as the undetermined ("und") track, with "eng" merely being another dub. Forcing
// a language there would hand the viewer English in the name of avoiding
// Portuguese.
//
// Subtitles are NOT this function's business: they are loaded whenever the stream
// ships them, dub or not (see GetSuperFlixStreamURL).
func audioForServer(s superflix.SuperFlixServer, streamAudio []string) string {
	if s.Type != superflix.SuperFlixAudioDubbed {
		return ""
	}

	// Labels are irrelevant here — only the Dubbed flag is read — so the subtitle
	// hint does not matter.
	for _, o := range superFlixAudioOptions(streamAudio, false) {
		if o.Dubbed {
			return mpvAudioLanguage(o)
		}
	}
	// The server says dubbed but the manifest exposes no Portuguese track; let mpv
	// keep its default rather than forcing a track that isn't there.
	return ""
}
