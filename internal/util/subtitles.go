package util

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/alvarorichard/Goanime/internal/tui"
)

// Subtitle session (tela 7): the first choice is asked once and remembered.
// "Use this choice for the rest of the session" skips the picker afterwards;
// "Always pick Portuguese when a source has it" is a rule evaluated first.
// All state resets per process; resetSubtitleSession is test-only.

// subtitleScreenFn opens the subtitles screen. Tests inject answers.
var subtitleScreenFn = tui.SelectSubtitleTrack

// subtitleSession records one remembered choice. It belongs to the show it
// was made for (forTitle): a "por" pick for one title must never
// auto-apply to another. The Portuguese rule is global by design.
type subtitleSession struct {
	set      bool
	all      bool
	none     bool
	language string
	forTitle string
}

var (
	subtitleSessionChoice subtitleSession
	subtitleLastLabel     string
	subtitleLastTitle     string
	subtitleAlwaysPT      bool
	subtitleCtx           subtitleContext
)

// subtitleContext feeds the screen's header and UP NEXT panel.
type subtitleContext struct {
	title    string
	source   string
	audio    string
	nextLine string
}

// SetSubtitleContext records the show around a subtitle prompt: its title,
// source and audio for the header, and the next-episode line for UP NEXT
// ("" omits each part).
func SetSubtitleContext(title, source, audio, nextLine string) {
	playbackStateMu.Lock()
	defer playbackStateMu.Unlock()
	subtitleCtx = subtitleContext{title: title, source: source, audio: audio, nextLine: nextLine}
}

// resetSubtitleSession clears the remembered choice and rules.
func resetSubtitleSession() {
	subtitleSessionChoice = subtitleSession{}
	subtitleLastLabel = ""
	subtitleLastTitle = ""
	subtitleAlwaysPT = false
	subtitleCtx = subtitleContext{}
}

// subtitleDisplayLabel is how a track reads on screen.
func subtitleDisplayLabel(sub SubtitleInfo, index int) string {
	if label := strings.TrimSpace(sub.Label); label != "" {
		return label
	}
	if lang := strings.TrimSpace(sub.Language); lang != "" {
		return lang
	}
	return fmt.Sprintf("Subtitle %d", index+1)
}

// isPortugueseSubtitle reports whether a track is Portuguese, by language
// code ("por", "pt"), name ("Portuguese", "Brasil") or PT-BR site word
// ("Legendado", "Dublado": a track labeled so serves PT speakers).
func isPortugueseSubtitle(sub SubtitleInfo) bool {
	text := strings.ToLower(strings.TrimSpace(sub.Language) + " " + strings.TrimSpace(sub.Label))
	for _, token := range strings.FieldsFunc(text, func(r rune) bool {
		return r != '-' && !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		switch token {
		case "pt", "por", "portugues", "portuguese", "brasil", "brazil", "brazilian", "ptbr", "pt-br",
			"legendado", "dublado":
			return true
		}
		if strings.HasPrefix(token, "portugu") || strings.HasPrefix(token, "brasil") || strings.HasPrefix(token, "brazil") {
			return true
		}
	}
	return false
}

// matchSubtitleLanguage finds the first track in the choice's language.
func matchSubtitleLanguage(tracks []SubtitleInfo, language string) int {
	language = strings.ToLower(strings.TrimSpace(language))
	if language == "" {
		return -1
	}
	for i, sub := range tracks {
		if strings.ToLower(strings.TrimSpace(sub.Language)) == language {
			return i
		}
	}
	for i, sub := range tracks {
		if strings.EqualFold(strings.TrimSpace(sub.Label), strings.TrimSpace(language)) {
			return i
		}
	}
	return -1
}

// applySubtitleSession applies the remembered choice or rule to the
// current tracks. True means the picker is skipped.
func applySubtitleSession() bool {
	if GlobalNoSubs || len(GlobalSubtitles) == 0 {
		return false
	}
	if subtitleAlwaysPT {
		for i, sub := range GlobalSubtitles {
			if isPortugueseSubtitle(sub) {
				GlobalSubtitles = []SubtitleInfo{GlobalSubtitles[i]}
				Debugf("Subtitles: Portuguese rule kept %s", subtitleDisplayLabel(sub, i))
				return true
			}
		}
	}
	choice := subtitleSessionChoice
	if !choice.set {
		return false
	}
	if choice.forTitle != strings.TrimSpace(subtitleCtx.title) {
		return false // another show; ask again
	}
	switch {
	case choice.all:
		return true // keep everything
	case choice.none:
		GlobalSubtitles = nil
		return true
	default:
		if i := matchSubtitleLanguage(GlobalSubtitles, choice.language); i >= 0 {
			kept := GlobalSubtitles[i]
			GlobalSubtitles = []SubtitleInfo{kept}
			return true
		}
		return false // the language is gone; ask again
	}
}

// applySubtitleChoiceLocked filters GlobalSubtitles to the screen's choice
// and records the session answers. Callers hold playbackStateMu.
func applySubtitleChoiceLocked(choice tui.SubtitleChoice, tracks []SubtitleInfo) {
	label := ""
	switch choice.Kind {
	case "all":
		label = "Both, switch in mpv with j"
	case "none":
		label = "None"
		GlobalSubtitles = nil
	default:
		if i := matchSubtitleLanguage(GlobalSubtitles, choice.Language); i >= 0 {
			kept := GlobalSubtitles[i]
			GlobalSubtitles = []SubtitleInfo{kept}
			label = subtitleDisplayLabel(kept, i)
		} else if len(tracks) > 0 {
			label = subtitleDisplayLabel(tracks[0], 0)
		}
	}
	subtitleLastLabel = label
	subtitleLastTitle = strings.TrimSpace(subtitleCtx.title)
	subtitleSessionChoice = subtitleSession{
		set:      choice.Session,
		all:      choice.Kind == "all",
		none:     choice.Kind == "none",
		language: choice.Language,
		forTitle: strings.TrimSpace(subtitleCtx.title),
	}
	subtitleAlwaysPT = choice.PreferPT
}

// buildSubtitleOptions renders the screen's rows from the current tracks.
func buildSubtitleOptions() tui.SubtitleOptions {
	ctx := subtitleCtx
	header := ""
	if ctx.source != "" {
		header = ctx.source
		if ctx.audio != "" {
			header += " · " + ctx.audio
		}
	}
	tracks := make([]tui.SubtitleTrack, 0, len(GlobalSubtitles))
	for _, sub := range GlobalSubtitles {
		tracks = append(tracks, tui.SubtitleTrack{Label: sub.Label, Language: sub.Language})
	}
	// The "last choice" marker follows the session binding: another show
	// does not inherit it.
	lastLabel := ""
	if subtitleLastTitle == strings.TrimSpace(ctx.title) {
		lastLabel = subtitleLastLabel
	}
	return tui.SubtitleOptions{
		Title:          ctx.title,
		Context:        header,
		Tracks:         tracks,
		LastLabel:      lastLabel,
		SessionChecked: true,
		NextTitle:      ctx.title,
		NextLine:       ctx.nextLine,
	}
}

// openSubtitlePicker runs the screen and applies its choice. False keeps the
// current tracks (cancel). Callers hold playbackStateMu for the apply.
func openSubtitlePicker() {
	opts := buildSubtitleOptions()
	choice, err := subtitleScreenFn(opts)
	if err != nil {
		Debugf("Subtitle choice cancelled; keeping current tracks")
		return
	}
	applySubtitleChoiceLocked(choice, GlobalSubtitles)
	switch choice.Kind {
	case "all":
		Debugf("User selected all %d subtitle track(s)", len(GlobalSubtitles))
	case "none":
		Debugf("User disabled subtitles")
	default:
		if len(GlobalSubtitles) == 1 {
			kept := GlobalSubtitles[0]
			Debugf("User selected subtitle: %s (%s)", kept.Label, kept.Language)
		}
	}
}

// SelectSubtitles displays the subtitles screen when there is a choice to
// make, and remembers the answer for the session. It updates GlobalSubtitles
// in-place so that the subsequent call to GetSubtitleArgs only includes the
// selected tracks. With 0 or 1 tracks there is nothing to choose.
func SelectSubtitles() {
	playbackStateMu.Lock()
	defer playbackStateMu.Unlock()
	if GlobalNoSubs || len(GlobalSubtitles) <= 1 {
		return
	}
	if applySubtitleSession() {
		return
	}
	openSubtitlePicker()
}

// PromptSubtitleLanguage always prompts for the subtitle language on
// multi-language platforms (e.g., 9Anime): every episode until the choice is
// remembered for the session. It updates GlobalSubtitles in-place.
func PromptSubtitleLanguage() {
	playbackStateMu.Lock()
	defer playbackStateMu.Unlock()
	if GlobalNoSubs {
		Debugf("Subtitles disabled by user (--no-subs), skipping subtitle prompt")
		GlobalSubtitles = nil
		return
	}
	if len(GlobalSubtitles) == 0 {
		fmt.Println("\nNo subtitle tracks available for this episode.")
		return
	}
	if applySubtitleSession() {
		printSubtitleResult(GlobalSubtitles)
		return
	}
	openSubtitlePicker()
	printSubtitleResult(GlobalSubtitles)
}

// printSubtitleResult reports the choice the way the old prompt did.
func printSubtitleResult(kept []SubtitleInfo) {
	switch len(kept) {
	case 0:
		fmt.Println("Subtitles: disabled")
	case 1:
		label := kept[0].Label
		if label == "" {
			label = kept[0].Language
		}
		if label == "" {
			label = "Unknown"
		}
		fmt.Printf("Subtitles: %s\n", label)
	default:
		fmt.Printf("Subtitles: all (%d tracks)\n", len(kept))
	}
}
