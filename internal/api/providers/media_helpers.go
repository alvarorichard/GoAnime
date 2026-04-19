package providers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

func chooseIndex[T any](items []T, render func(int) string, prompt string) (int, error) {
	return tui.ChooseIndex(items, render, prompt)
}

func preferredSubsLanguage() string {
	if subsLanguage := strings.TrimSpace(util.GetPreferredSubtitleLanguage()); subsLanguage != "" {
		return subsLanguage
	}
	return "english"
}

func contextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if ctx == nil {
		return context.WithTimeout(context.Background(), timeout)
	}
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func selectFlixHQQualityOption(client *scraper.FlixHQClient, qualities []scraper.FlixHQQualityOption) (scraper.FlixHQQualityOption, error) {
	if len(qualities) == 0 {
		return scraper.FlixHQQualityOption{}, fmt.Errorf("no qualities available")
	}

	idx, err := chooseIndex(qualities, func(i int) string {
		return client.QualityToLabel(qualities[i].Quality)
	}, "Select video quality: ")
	if err != nil {
		return qualities[0], err
	}
	return qualities[idx], nil
}

func selectSFlixQualityOption(client *scraper.SFlixClient, qualities []scraper.SFlixQualityOption) (scraper.SFlixQualityOption, error) {
	if len(qualities) == 0 {
		return scraper.SFlixQualityOption{}, fmt.Errorf("no qualities available")
	}

	idx, err := chooseIndex(qualities, func(i int) string {
		return client.QualityToLabel(qualities[i].Quality)
	}, "Select video quality: ")
	if err != nil {
		return qualities[0], err
	}
	return qualities[idx], nil
}

func applyFlixHQPlaybackState(streamInfo *scraper.FlixHQStreamInfo) {
	if streamInfo == nil {
		return
	}

	if referer := strings.TrimSpace(streamInfo.Referer); referer != "" {
		util.SetGlobalReferer(referer)
	}

	if util.SubtitlesDisabled() || len(streamInfo.Subtitles) == 0 {
		return
	}

	subInfos := make([]util.SubtitleInfo, 0, len(streamInfo.Subtitles))
	for _, sub := range streamInfo.Subtitles {
		subURL := strings.TrimSpace(sub.URL)
		if subURL == "" {
			continue
		}

		label := strings.TrimSpace(sub.Label)
		if label == "" {
			label = strings.TrimSpace(sub.Language)
		}

		subInfos = append(subInfos, util.SubtitleInfo{
			URL:      subURL,
			Language: strings.TrimSpace(sub.Language),
			Label:    label,
		})
	}

	if len(subInfos) > 0 {
		util.SetGlobalSubtitles(subInfos)
	}
}

func applySFlixPlaybackState(streamInfo *scraper.SFlixStreamInfo) {
	if streamInfo == nil {
		return
	}

	if referer := strings.TrimSpace(streamInfo.Referer); referer != "" {
		util.SetGlobalReferer(referer)
	}

	if util.SubtitlesDisabled() || len(streamInfo.Subtitles) == 0 {
		return
	}

	subInfos := make([]util.SubtitleInfo, 0, len(streamInfo.Subtitles))
	for _, sub := range streamInfo.Subtitles {
		subURL := strings.TrimSpace(sub.URL)
		if subURL == "" {
			continue
		}

		label := strings.TrimSpace(sub.Label)
		if label == "" {
			label = strings.TrimSpace(sub.Language)
		}

		subInfos = append(subInfos, util.SubtitleInfo{
			URL:      subURL,
			Language: strings.TrimSpace(sub.Language),
			Label:    label,
		})
	}

	if len(subInfos) > 0 {
		util.SetGlobalSubtitles(subInfos)
	}
}

func applySuperFlixPlaybackResult(anime *models.Anime, result *scraper.SuperFlixStreamResult) {
	if result == nil {
		return
	}

	if referer := strings.TrimSpace(result.Referer); referer != "" {
		util.SetGlobalReferer(referer)
	}

	if anime != nil && anime.ImageURL == "" && result.Thumb != "" {
		anime.ImageURL = scraper.NormalizeSuperFlixImageURL(result.Thumb)
	}

	if util.SubtitlesDisabled() || len(result.Subtitles) == 0 {
		return
	}

	subInfos := make([]util.SubtitleInfo, 0, len(result.Subtitles))
	for _, sub := range result.Subtitles {
		subURL := strings.TrimSpace(sub.URL)
		if subURL == "" {
			continue
		}

		label := strings.TrimSpace(sub.Lang)
		if label == "" {
			label = "Unknown"
		}

		subInfos = append(subInfos, util.SubtitleInfo{
			URL:      subURL,
			Language: strings.ToLower(label),
			Label:    label,
		})
	}

	if len(subInfos) > 0 {
		util.SetGlobalSubtitles(subInfos)
	}
}
