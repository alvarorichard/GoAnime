package api

import (
	"fmt"
	"sync"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// SuperFlix offers, per episode, a list of servers — and each server is tagged
// Dublado (type 1) or Legendado (type 2). That single list answers BOTH questions
// a viewer has: which source to play from, and whether they want the dub or the
// original audio. So we ask them once, together, instead of twice.
//
// The list is only reachable through the player page + /player/bootstrap. The
// embed page the browser sniff lands on carries no servers at all, which is why
// playback used to take whatever the embed happened to play, with the user
// choosing neither.

// serverChoice is one row in the picker.
type serverChoice struct {
	Server superflix.SuperFlixServer
	Label  string
}

// serverAudioLabel names the audio a server carries, in the site's own terms.
func serverAudioLabel(s superflix.SuperFlixServer) string {
	switch s.Type {
	case superflix.SuperFlixAudioDubbed:
		return "🎙️  Dublado"
	case superflix.SuperFlixAudioSubtitled:
		return "💬 Legendado"
	default:
		return "❔ Áudio desconhecido"
	}
}

// superFlixServerChoices renders the servers for the picker, dubbed ones first
// (the default expectation for a PT-BR catalog) and otherwise in the order the
// site declared them, so the list is stable between episodes.
func superFlixServerChoices(servers []superflix.SuperFlixServer) []serverChoice {
	var dubbed, subtitled, other []serverChoice

	for _, s := range servers {
		c := serverChoice{
			Server: s,
			Label:  fmt.Sprintf("%s · %s", serverAudioLabel(s), s.Name),
		}
		switch s.Type {
		case superflix.SuperFlixAudioDubbed:
			dubbed = append(dubbed, c)
		case superflix.SuperFlixAudioSubtitled:
			subtitled = append(subtitled, c)
		default:
			other = append(other, c)
		}
	}

	out := make([]serverChoice, 0, len(servers))
	out = append(out, dubbed...)
	out = append(out, subtitled...)
	out = append(out, other...)
	return out
}

// sfServerPref is what we remember about a user's pick.
//
// It deliberately records the server's STABLE properties, not its name or id: a
// server is called "Servidor 159462", and that number changes with every episode,
// so remembering the name would never match again. Type (dublado/legendado) and
// IsFile (direct MP4 vs streaming) do carry across episodes, which is what lets a
// binge ask once and then honor the answer silently.
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

// sfServerPickFn shows the picker. A seam so the selection logic is testable
// without a TTY.
var sfServerPickFn = func(labels []string) (int, error) {
	return tui.Find(labels, func(i int) string { return labels[i] },
		fuzzyfinder.WithPromptString("Servidor (dublado ou legendado): "))
}

// matchRemembered returns the servers that fit a remembered preference, trying
// the exact match (same audio AND same kind) before relaxing to just the audio —
// so a title whose MP4 mirror disappears this episode still honors "dublado".
func matchRemembered(servers []serverChoice, pref sfServerPref) []serverChoice {
	var exact, sameAudio []serverChoice
	for _, c := range servers {
		if c.Server.Type != pref.Type {
			continue
		}
		sameAudio = append(sameAudio, c)
		if c.Server.IsFile == pref.IsFile {
			exact = append(exact, c)
		}
	}
	if len(exact) > 0 {
		return exact
	}
	return sameAudio
}

// selectSuperFlixServer picks the server to stream from, asking the user only
// when there is a real decision to make.
//
// It stays quiet when a prompt would be noise: a single server, or a title whose
// preference we already know and that still resolves to exactly one server. If the
// picker cannot run (no TTY) or the user backs out, it falls back to the first
// option — a dubbed server, since dubbed sorts first.
func selectSuperFlixServer(tmdbID string, servers []superflix.SuperFlixServer) (superflix.SuperFlixServer, error) {
	choices := superFlixServerChoices(servers)
	if len(choices) == 0 {
		return superflix.SuperFlixServer{}, fmt.Errorf("no servers to choose from")
	}

	if len(choices) == 1 {
		rememberSuperFlixServer(tmdbID, choices[0].Server)
		return choices[0].Server, nil
	}

	if pref, ok := recallSuperFlixServer(tmdbID); ok {
		if m := matchRemembered(choices, pref); len(m) == 1 {
			util.Debug("SuperFlix server: honoring the choice made for this title",
				"type", pref.Type, "isFile", pref.IsFile, "server", m[0].Server.Name)
			return m[0].Server, nil
		}
	}

	labels := make([]string, len(choices))
	for i, c := range choices {
		labels[i] = c.Label
	}

	idx, err := sfServerPickFn(labels)
	if err != nil || idx < 0 || idx >= len(choices) {
		util.Debug("SuperFlix server: picker unavailable, defaulting to the first option", "err", err)
		rememberSuperFlixServer(tmdbID, choices[0].Server)
		return choices[0].Server, nil
	}

	rememberSuperFlixServer(tmdbID, choices[idx].Server)
	return choices[idx].Server, nil
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
