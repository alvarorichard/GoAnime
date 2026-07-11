package api

import (
	"fmt"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
)

// SuperFlix streams are multi-audio HLS: one manifest carries the Portuguese dub
// alongside the original audio (and often Spanish/Korean/Chinese), and the
// Portuguese subtitle track is a separate file listed on the player page.
//
// "Dublado ou legendado" is therefore an AUDIO-TRACK choice, not a server choice
// — SuperFlix's embed no longer exposes a server list at all. This file turns the
// raw track codes the stream declares (["por","und","eng","spa","kor","jpn"])
// into a choice a person can actually make, and maps it to mpv's language
// preferences.

// audioOption is one selectable audio track on a SuperFlix stream.
type audioOption struct {
	// Code is the track's language as the stream declares it (ISO-639-2: "por").
	Code string
	// Label is what the user sees in the picker.
	Label string
	// Dubbed marks the Portuguese track: for a PT-BR audience that IS "dublado",
	// and it is the one case where burning in the Portuguese subtitles would just
	// duplicate the audio.
	Dubbed bool
}

// audioLanguageNames maps the ISO-639-2 codes SuperFlix emits to Portuguese
// names. Codes we don't know are shown as-is rather than dropped — an unnamed
// track the user can still pick beats a track we silently hide.
var audioLanguageNames = map[string]string{
	"por": "Português",
	"pob": "Português",
	"pt":  "Português",
	"jpn": "Japonês",
	"jpa": "Japonês",
	"ja":  "Japonês",
	"eng": "Inglês",
	"en":  "Inglês",
	"spa": "Espanhol",
	"es":  "Espanhol",
	"kor": "Coreano",
	"chi": "Chinês",
	"zho": "Chinês",
	"fre": "Francês",
	"fra": "Francês",
	"ger": "Alemão",
	"deu": "Alemão",
	"ita": "Italiano",
	"rus": "Russo",
}

// portugueseAudioCodes are the codes that mean "the Portuguese dub".
var portugueseAudioCodes = map[string]bool{"por": true, "pob": true, "pt": true, "ptb": true}

// mpvAudioSynonyms expands a track code into the alias list mpv's --alang matches
// against. mpv compares against whatever the manifest tags the track with, and
// different SuperFlix sources tag the same language differently, so we pass every
// plausible spelling instead of betting on one.
var mpvAudioSynonyms = map[string][]string{
	"por": {"por", "pob", "pt-BR", "ptbr", "pt", "portuguese"},
	"jpn": {"jpn", "ja", "jp", "japanese"},
	"eng": {"eng", "en", "english"},
	"spa": {"spa", "es", "spanish"},
	"kor": {"kor", "ko", "korean"},
	"chi": {"chi", "zho", "zh", "chinese"},
	"fre": {"fre", "fra", "fr", "french"},
	"ger": {"ger", "deu", "de", "german"},
	"ita": {"ita", "it", "italian"},
	"rus": {"rus", "ru", "russian"},
}

// superFlixAudioOptions turns the stream's raw audio-track codes into the choices
// offered to the user.
//
// It drops "und" (undetermined — SuperFlix pads the list with it and it is not a
// language anyone can choose) and de-duplicates, then puts Portuguese first
// because that is what most users of a PT-BR catalog want. Everything else keeps
// the order the stream declared, so the list stays stable between plays.
//
// hasSubtitles changes what the non-Portuguese tracks HONESTLY are. "Legendado"
// promises subtitles, and not every SuperFlix stream ships them: Mushoku Tensei's
// carries the same seven audio tracks as Dexter's but no subtitle file at all, so
// labelling its Japanese track "Legendado" would sell the viewer an unwatchable
// episode. When there are no subtitles those tracks are offered as raw original
// audio, and named that way.
//
// A stream with fewer than two real tracks offers no choice; the caller uses that
// to skip the prompt entirely rather than asking a question with one answer.
func superFlixAudioOptions(codes []string, hasSubtitles bool) []audioOption {
	seen := make(map[string]bool, len(codes))
	var portuguese []audioOption
	var others []audioOption

	for _, raw := range codes {
		code := strings.ToLower(strings.TrimSpace(raw))
		if code == "" || code == "und" || seen[code] {
			continue
		}
		seen[code] = true

		name, known := audioLanguageNames[code]
		if !known {
			name = strings.ToUpper(code)
		}

		if portugueseAudioCodes[code] {
			portuguese = append(portuguese, audioOption{
				Code:   code,
				Label:  fmt.Sprintf("🎙️  %s (Dublado)", name),
				Dubbed: true,
			})
			continue
		}

		label := fmt.Sprintf("💬 %s (Legendado)", name)
		if !hasSubtitles {
			label = fmt.Sprintf("🔊 %s (áudio original — sem legendas)", name)
		}
		others = append(others, audioOption{Code: code, Label: label})
	}

	return append(portuguese, others...)
}

// mpvAudioLanguage renders an option as an mpv --alang value: the chosen language
// first, with its aliases, so mpv picks that track whatever the manifest calls it.
func mpvAudioLanguage(opt audioOption) string {
	if syn, ok := mpvAudioSynonyms[opt.Code]; ok {
		return strings.Join(syn, ",")
	}
	return opt.Code
}

// sfAudioChoices remembers, per title, what the user picked — so a binge asks once
// instead of once per episode. Session-scoped on purpose: it is a playback
// preference, not something worth persisting to disk.
var (
	sfAudioChoiceMu sync.Mutex
	sfAudioChoices  = map[string]audioOption{}
)

func rememberSuperFlixAudio(tmdbID string, opt audioOption) {
	sfAudioChoiceMu.Lock()
	defer sfAudioChoiceMu.Unlock()
	sfAudioChoices[tmdbID] = opt
}

func recallSuperFlixAudio(tmdbID string) (audioOption, bool) {
	sfAudioChoiceMu.Lock()
	defer sfAudioChoiceMu.Unlock()
	opt, ok := sfAudioChoices[tmdbID]
	return opt, ok
}

// userPinnedAudioLang snapshots --audio-lang the first time we look, which is
// always BEFORE we write GlobalAudioLanguage ourselves.
//
// The snapshot is the whole point: without it, the language we set for episode 1
// would read back as a user preference on episode 2 — suppressing the picker and
// leaking one title's choice into every title after it.
var (
	audioPinOnce sync.Once
	audioPinned  string
)

func userPinnedAudioLang() string {
	audioPinOnce.Do(func() { audioPinned = util.GlobalAudioLanguage })
	return audioPinned
}

// resetSuperFlixAudioChoices clears the per-title memory and the --audio-lang
// snapshot. Tests only.
func resetSuperFlixAudioChoices() {
	sfAudioChoiceMu.Lock()
	defer sfAudioChoiceMu.Unlock()
	sfAudioChoices = map[string]audioOption{}
	audioPinOnce = sync.Once{}
	audioPinned = ""
}

// sfAudioPickFn shows the audio picker. A seam so the selection logic can be
// tested without a TTY.
var sfAudioPickFn = func(labels []string) (int, error) {
	return tui.Find(labels, func(i int) string { return labels[i] },
		fuzzyfinder.WithPromptString("Áudio (dublado ou legendado): "))
}

// selectSuperFlixAudio decides which audio track to play, asking the user only
// when there is a real decision to make.
//
// It stays quiet in every case where a prompt would be noise: a stream with a
// single track, a title the user already chose for this session, and a user who
// pinned a language on the command line. If the picker cannot run (no TTY) or the
// user backs out, it falls back to the first option — Portuguese, i.e. the dub —
// which is the safe default for a PT-BR catalog.
//
// The bool reports whether we have an option to act on at all.
func selectSuperFlixAudio(tmdbID string, codes []string, hasSubtitles bool) (audioOption, bool) {
	// An explicit --audio-lang wins: the user already answered this question.
	if userPinnedAudioLang() != "" {
		return audioOption{}, false
	}

	opts := superFlixAudioOptions(codes, hasSubtitles)
	if len(opts) == 0 {
		return audioOption{}, false
	}

	if prev, ok := recallSuperFlixAudio(tmdbID); ok {
		util.Debug("SuperFlix audio: reusing the choice made for this title", "code", prev.Code)
		return prev, true
	}

	// Only one track: nothing to ask.
	if len(opts) == 1 {
		rememberSuperFlixAudio(tmdbID, opts[0])
		return opts[0], true
	}

	labels := make([]string, len(opts))
	for i, o := range opts {
		labels[i] = o.Label
	}

	idx, err := sfAudioPickFn(labels)
	if err != nil || idx < 0 || idx >= len(opts) {
		util.Debug("SuperFlix audio: picker unavailable, defaulting to the first track", "err", err)
		rememberSuperFlixAudio(tmdbID, opts[0])
		return opts[0], true
	}

	rememberSuperFlixAudio(tmdbID, opts[idx])
	return opts[idx], true
}
