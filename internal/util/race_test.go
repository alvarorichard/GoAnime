package util

import (
	"fmt"
	"sync"
	"testing"
)

func TestPlaybackGlobalsConcurrentAccess(t *testing.T) {
	GlobalNoSubs = false

	var wg sync.WaitGroup
	for worker := range 8 {
		wg.Go(func() {
			for iteration := range 1000 {
				value := fmt.Sprintf("%d-%d", worker, iteration)
				SetGlobalSubtitles([]SubtitleInfo{{URL: "https://example.test/" + value}})
				subs := GetGlobalSubtitles()
				if len(subs) > 0 {
					subs[0].URL = "mutated caller copy"
				}
				_ = GetSubtitleArgs()

				SetGlobalReferer(value)
				_ = GetGlobalReferer()
				SetGlobalAnimeSource(value)
				_ = GetGlobalAnimeSource()
				SetGlobalAudioLanguage(value)
				_ = GetGlobalAudioLanguage()
			}
		})
	}
	wg.Wait()
}
