package source

import (
	"context"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Subtitle is one caption track that comes with a stream.
type Subtitle struct {
	URL      string
	Language string
	Label    string
}

// Stream is a resolved stream and the playback context that goes with it. A
// source returns it from FetchStream instead of leaving the context behind in
// process-wide state; Publish hands it to the player in one place.
type Stream struct {
	URL       string
	Referer   string // sent with every media request
	Origin    string // an explicit CORS origin, when the CDN wants one
	AudioLang string // the language the file carries, as a hint for mpv
	Subtitles []Subtitle
}

// StreamFetcher is the optional half of Source that returns a Stream. A source
// that implements it no longer writes playback state itself; one that does not
// keeps resolving through FetchStreamURL.
type StreamFetcher interface {
	FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (Stream, error)
}

// Publish hands the stream's playback context to the player. Fields the stream
// leaves empty are not touched: a source that knows no origin leaves whatever
// the caller cleared before resolving.
func (s Stream) Publish() {
	if s.Referer != "" {
		util.SetGlobalReferer(s.Referer)
	}
	if s.Origin != "" {
		util.SetGlobalOrigin(s.Origin)
	}
	if s.AudioLang != "" {
		util.SetGlobalAudioLanguage(s.AudioLang)
	}
	if len(s.Subtitles) > 0 && !util.GlobalNoSubs {
		subs := make([]util.SubtitleInfo, 0, len(s.Subtitles))
		for _, sub := range s.Subtitles {
			subs = append(subs, util.SubtitleInfo{URL: sub.URL, Language: sub.Language, Label: sub.Label})
		}
		util.SetGlobalSubtitles(subs)
	}
}
