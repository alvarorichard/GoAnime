// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

// FlattenExtractedInfo returns leaf videos from infos, recursing
// playlist and multi_video entries. video, url, url_transparent, and
// empty/unknown types are leaves. Nil-safe. --dump-json already emits one
// object per video, so this is a no-op on that path; it is required for
// --dump-single-json playlist wrappers.
func FlattenExtractedInfo(infos []*ExtractedInfo) []*ExtractedInfo {
	out := make([]*ExtractedInfo, 0, len(infos))
	for _, info := range infos {
		if info == nil {
			continue
		}
		if info.IsPlaylist() {
			out = append(out, FlattenExtractedInfo(info.Entries)...)
			continue
		}
		out = append(out, info)
	}
	return out
}

// IsPlaylist reports whether e is a playlist or multi_video wrapper.
func (e *ExtractedInfo) IsPlaylist() bool {
	if e == nil {
		return false
	}
	return e.Type == ExtractedTypePlaylist || e.Type == ExtractedTypeMultiVideo
}

// BestThumbnail returns the preferred thumbnail. [ExtractedInfo.Thumbnail]
// wins when non-empty; otherwise the highest [ExtractedThumbnail.Preference],
// then [ExtractedThumbnail.Width]. Nil-safe.
func (e *ExtractedInfo) BestThumbnail() *ExtractedThumbnail {
	if e == nil {
		return nil
	}
	if e.Thumbnail != nil && *e.Thumbnail != "" {
		return &ExtractedThumbnail{URL: *e.Thumbnail}
	}

	var best *ExtractedThumbnail
	for _, thumb := range e.Thumbnails {
		if thumb == nil || thumb.URL == "" {
			continue
		}
		if best == nil || thumbnailBetter(thumb, best) {
			best = thumb
		}
	}
	return best
}

// ThumbnailURL returns [ExtractedInfo.BestThumbnail]'s URL, or empty.
func (e *ExtractedInfo) ThumbnailURL() string {
	if thumb := e.BestThumbnail(); thumb != nil {
		return thumb.URL
	}
	return ""
}

func thumbnailBetter(a, b *ExtractedThumbnail) bool {
	ap, bp := 0, 0
	if a.Preference != nil {
		ap = *a.Preference
	}
	if b.Preference != nil {
		bp = *b.Preference
	}
	if ap != bp {
		return ap > bp
	}
	aw, bw := 0, 0
	if a.Width != nil {
		aw = *a.Width
	}
	if b.Width != nil {
		bw = *b.Width
	}
	return aw > bw
}
