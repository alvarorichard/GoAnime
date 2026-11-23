// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"reflect"
	"strings"
)

// Result contains the yt-dlp execution results, including stdout/stderr, exit code,
// and any output logs. Note that output logs should already be pre-processed via
// [Result.OutputLogs], which will be sorted by timestamp, and any JSON parsed
// (if configured with [Command.PrintJson] or similar).
type Result struct {
	// Executable is the path to the yt-dlp executable that was invoked.
	Executable string `json:"executable"`

	// Args are the arguments that were passed to yt-dlp, excluding the executable.
	Args []string `json:"args"`

	// ExitCode is the exit code of the yt-dlp process.
	ExitCode int `json:"exit_code"`

	// Stdout is the stdout of the yt-dlp process, with trailing newlines removed.
	Stdout string `json:"stdout"`

	// Stderr is the stderr of the yt-dlp process, with trailing newlines removed.
	Stderr string `json:"stderr"`

	// OutputLogs are the stdout/stderr logs, sorted by timestamp, and any JSON
	// parsed (if configured with [Command.PrintJson]).
	OutputLogs []*ResultLog `json:"output_logs"`
}

func (r *Result) asString(stdout, stderr, timestamps, maskJSON, exitCode bool) string {
	out := make([]string, 0, len(r.OutputLogs)+1)

	for _, l := range r.OutputLogs {
		if l.Pipe == PipeStdout && !stdout {
			continue
		}

		if l.Pipe == PipeStderr && !stderr {
			continue
		}

		out = append(out, l.asString(timestamps, maskJSON))
	}

	if exitCode {
		out = append(out, fmt.Sprintf("exit code: %d", r.ExitCode))
	}

	return strings.Join(out, "\n")
}

func (r *Result) String() string {
	return r.asString(true, true, true, true, true)
}

// GetExtractedInfo returns the extracted info from the yt-dlp output logs. Note that
// this will only return info if yt-dlp was invoked with [Command.PrintJson] or
// similar.
func (r *Result) GetExtractedInfo() (info []*ExtractedInfo, err error) {
	var e *ExtractedInfo

	for _, log := range r.OutputLogs {
		if log.JSON == nil {
			continue
		}

		e, err = ParseExtractedInfo(log.JSON)
		if err != nil {
			return nil, err
		}

		if e.Type == "" {
			continue // Not an extracted info result.
		}

		if info == nil {
			info = make([]*ExtractedInfo, 0, len(r.OutputLogs))
		}
		info = append(info, e)
	}

	return info, nil
}

// Extractor data fields are defined by yt-dlp's InfoExtractor schema:
//   - https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/extractor/common.py#L119-L489
// Additional generated fields are documented in yt-dlp's output template:
//   - https://github.com/yt-dlp/yt-dlp/blob/master/README.md#output-template

// ParseExtractedInfo parses the extracted info from msg. ParseExtractedInfo will
// also clean the returned results to remove some ytdlp-isims, such as "none" for
// some string fields.
func ParseExtractedInfo(msg *jsontext.Value) (info *ExtractedInfo, err error) {
	info = &ExtractedInfo{source: msg}

	err = json.Unmarshal(*msg, info)
	if err != nil {
		return nil, err
	}

	cleanJSONValue(reflect.ValueOf(info))
	return info, nil
}

func cleanJSONValue(v reflect.Value) {
	for v.IsValid() && v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	if !v.IsValid() || v.Kind() != reflect.Struct {
		return
	}

	for i := range v.NumField() {
		field := v.Field(i)
		if !field.CanSet() {
			continue
		}

		switch field.Kind() { //nolint:exhaustive // unsupported kinds need no cleaning
		case reflect.Struct:
			cleanJSONValue(field)
		case reflect.Pointer:
			if field.IsNil() {
				continue
			}

			value := field.Elem()
			if value.Kind() == reflect.Struct {
				cleanJSONValue(value)
				continue
			}

			if value.Kind() == reflect.String && (value.String() == "none" || value.String() == "") {
				// If field name == "Title", set to empty string instead of nil.
				// See [ExtractedInfo.Title] for more info.
				if v.Type().Field(i).Name == "Title" {
					value.SetString("")
					continue
				}
				field.SetZero()
			}
		case reflect.Slice:
			for j := range field.Len() {
				cleanJSONValue(field.Index(j))
			}
		case reflect.String:
			if field.String() == "none" {
				field.SetString("")
			}
		}
	}
}

// ExtractedInfo contains metadata returned by yt-dlp for a video, playlist, or
// other extracted result.
type ExtractedInfo struct {
	// ExtractedFormat fields which can also be returned for ExtractedInfo.
	*ExtractedFormat

	// source is the original JSON document.
	source *jsontext.Value `json:"-"`

	// yt-dlp documents several metadata values as integers, but does not
	// enforce that type: its numeric-field sanitizer preserves both int and
	// float values. Keep uncertain extractor-provided numeric values as float64
	// so parsing remains lossless. See:
	// https://github.com/yt-dlp/yt-dlp/blob/master/yt_dlp/YoutubeDL.py#L2853-L2859

	// Type is the type of the video or returned result.
	Type ExtractedType `json:"_type"`

	// Version information is sometimes returned in the result.
	Version *ExtractedVersion `json:"_version,omitempty"`

	// ID is the video identifier.
	ID string `json:"id"`

	// Title contains the video title, unescaped. Set to an empty string if video
	// has no title as opposed to nil which signifies that the extractor failed
	// to obtain a title.
	Title *string `json:"title"`

	//
	// ExtractedInfo must contain either formats or a URL.
	//

	// Formats contains a list of each format available, ordered from worst to best
	// quality.
	Formats []*ExtractedFormat `json:"formats"`

	// RequestedFormats contains the formats selected for download.
	RequestedFormats []*ExtractedFormat `json:"requested_formats,omitempty"`

	// URL is the final video URL.
	URL *string `json:"url,omitempty"`

	// Filename is the video filename. This is not set when simulation is requested/
	// enabled.
	Filename *string `json:"filename,omitempty"`

	// AltFilename is an alternative filename for the video. This maps to "_filename".
	// See [ExtractedInfo.Filename] for more info.
	AltFilename *string `json:"_filename,omitempty"`

	// Extension is the video filename extension.
	Extension string `json:"ext"`

	// Format is the video format. Defaults to [ExtractedInfo.Extension] (used for
	// get-format functionality).
	Format string `json:"format"`

	// PlayerURL is the SWF Player URL (used for rtmpdump).
	PlayerURL *string `json:"player_url,omitempty"`

	// Direct is true if a direct video file was given (must only be set by GenericIE).
	Direct *bool `json:"direct,omitempty"`

	// AltTitle is a secondary title of the video.
	AltTitle *string `json:"alt_title,omitempty"`

	// DisplayID is an alternative identifier for the video, not necessarily unique,
	// but available before title. Typically, id is something like "4234987",
	// title "Dancing naked mole rats", and display_id "dancing-naked-mole-rats".
	DisplayID *string `json:"display_id,omitempty"`

	// Thumbnails is a list of thumbnails for the video.
	Thumbnails []*ExtractedThumbnail `json:"thumbnails,omitempty"`

	// Thumbnail is a full URL to a video thumbnail image.
	Thumbnail *string `json:"thumbnail,omitempty"`

	// Description contains the full video description.
	Description *string `json:"description,omitempty"`

	// Uploader contains the full name of the video uploader.
	Uploader *string `json:"uploader,omitempty"`

	// License contains the license name the video is licensed under.
	License *string `json:"license,omitempty"`

	// Creator contains the creator of the video.
	//
	// Deprecated: use Creators instead.
	Creator *string `json:"creator,omitempty"`

	// Creators contains the creators of the video.
	Creators []string `json:"creators,omitempty"`

	// Timestamp contains the UNIX timestamp of the moment the video was uploaded.
	Timestamp *float64 `json:"timestamp,omitempty"`

	// UploadDate contains the video upload date in UTC (YYYYMMDD). If not explicitly
	// set, calculated from Timestamp.
	UploadDate *string `json:"upload_date,omitempty"`

	// ReleaseTimestamp contains the UNIX timestamp of the moment the video was
	// released. If it is not clear whether to use Timestamp or this, use the former.
	ReleaseTimestamp *float64 `json:"release_timestamp,omitempty"`

	// ReleaseDate contains the date (YYYYMMDD) when the video was released in UTC.
	// If not explicitly set, calculated from ReleaseTimestamp.
	ReleaseDate *string `json:"release_date,omitempty"`

	// ReleaseYear contains the year (YYYY) when the video or album was released.
	// It is used when no exact release date is known.
	ReleaseYear *int `json:"release_year,omitempty"`

	// ModifiedTimestamp contains the UNIX timestamp of the moment the video was
	// last modified.
	ModifiedTimestamp *float64 `json:"modified_timestamp,omitempty"`

	// ModifiedDate contains the date (YYYYMMDD) when the video was last modified
	// in UTC. If not explicitly set, calculated from ModifiedTimestamp.
	ModifiedDate *string `json:"modified_date,omitempty"`

	// UploaderID contains the nickname or id of the video uploader.
	UploaderID *string `json:"uploader_id,omitempty"`

	// UploaderURL contains the full URL to a personal webpage of the video uploader.
	UploaderURL *string `json:"uploader_url,omitempty"`

	// Channel contains the full name of the channel the video is uploaded on.
	// Note that channel fields may or may not repeat uploader fields. This depends
	// on a particular extractor.
	Channel *string `json:"channel,omitempty"`

	// ChannelID contains the id of the channel.
	ChannelID *string `json:"channel_id,omitempty"`

	// ChannelURL contains the full URL to a channel webpage.
	ChannelURL *string `json:"channel_url,omitempty"`

	// ChannelFollowerCount contains the number of followers of the channel.
	ChannelFollowerCount *float64 `json:"channel_follower_count,omitempty"`

	// ChannelIsVerified is true if the channel is verified on the platform.
	ChannelIsVerified *bool `json:"channel_is_verified,omitempty"`

	// Location contains the physical location where the video was filmed.
	Location *string `json:"location,omitempty"`

	// Subtitles contains the available subtitles, where the key is the language
	// code and the value is a list of subtitle formats. Each format has an
	// extension and either a URL or inline data.
	Subtitles map[string][]*ExtractedSubtitle `json:"subtitles,omitempty"`

	// RequestedSubtitles contains the subtitles selected for download, where the
	// key is the language code, and the value is the single chosen subtitle
	// format (yt-dlp's process_subtitles picks one format per language).
	RequestedSubtitles map[string]*ExtractedSubtitle `json:"requested_subtitles,omitempty"`

	// AutomaticCaptions contains the automatically generated captions instead of
	// normal subtitles.
	AutomaticCaptions map[string][]*ExtractedSubtitle `json:"automatic_captions,omitempty"`

	// Duration contains the length of the video in seconds.
	Duration *float64 `json:"duration,omitempty"`

	// ViewCount contains how many users have watched the video on the platform.
	ViewCount *float64 `json:"view_count,omitempty"`

	// ConcurrentViewCount contains how many users are currently watching the video
	// on the platform.
	ConcurrentViewCount *float64 `json:"concurrent_view_count,omitempty"`

	// SaveCount contains the number of times the video has been saved or bookmarked.
	SaveCount *float64 `json:"save_count,omitempty"`

	// LikeCount contains the number of positive ratings of the video.
	LikeCount *float64 `json:"like_count,omitempty"`

	// DislikeCount contains the number of negative ratings of the video.
	DislikeCount *float64 `json:"dislike_count,omitempty"`

	// RepostCount contains the number of reposts of the video.
	RepostCount *float64 `json:"repost_count,omitempty"`

	// AverageRating contains the average rating given by users, the scale used
	// depends on the webpage.
	AverageRating *float64 `json:"average_rating,omitempty"`

	// CommentCount contains the number of comments on the video.
	CommentCount *float64 `json:"comment_count,omitempty"`

	// Comments contains a list of comments.
	Comments []*ExtractedVideoComment `json:"comments,omitempty"`

	// AgeLimit contains the age restriction for the video (years).
	AgeLimit *float64 `json:"age_limit,omitempty"`

	// WebpageURL contains the URL to the video webpage, if given to yt-dlp it
	// should allow to get the same result again. (It will be set by YoutubeDL if
	// it's missing)
	WebpageURL *string `json:"webpage_url,omitempty"`

	// Categories contains a list of categories that the video falls in, for example
	// ["Sports", "Berlin"].
	Categories []string `json:"categories,omitempty"`

	// Tags contains a list of tags assigned to the video, e.g. ["sweden", "pop music"].
	Tags []string `json:"tags,omitempty"`

	// Cast contains a list of the video cast.
	Cast []string `json:"cast,omitempty"`

	// IsLive is true, false, or nil (=unknown). Whether this video is a live stream
	// that goes on instead of a fixed-length video.
	IsLive *bool `json:"is_live,omitempty"`

	// WasLive is true, false, or nil (=unknown). Whether this video was originally
	// a live stream.
	WasLive *bool `json:"was_live,omitempty"`

	// LiveStatus is nil (=unknown), 'is_live', 'is_upcoming', 'was_live', 'not_live',
	// or 'post_live' (was live, but VOD is not yet processed). If absent, automatically
	// set from IsLive, WasLive.
	LiveStatus *ExtractedLiveStatus `json:"live_status,omitempty"`

	// StartTime is the time in seconds where the reproduction should start, as
	// specified in the URL.
	StartTime *float64 `json:"start_time,omitempty"`

	// EndTime is the time in seconds where the reproduction should end, as
	// specified in the URL.
	EndTime *float64 `json:"end_time,omitempty"`

	// Chapters is a list of chapters.
	Chapters []*ExtractedChapterData `json:"chapters,omitempty"`

	// Heatmap is a list of heatmap data points.
	Heatmap []*ExtractedHeatmapData `json:"heatmap,omitempty"`

	// PlayableInEmbed is whether this video is allowed to play in embedded players
	// on other sites. Can be true (=always allowed), false (=never allowed), nil
	// (=unknown), or a string specifying the criteria for embedability; e.g.
	// 'whitelist'.
	PlayableInEmbed any `json:"playable_in_embed,omitempty"`

	// Availability is under what condition the video is available.
	Availability *ExtractedAvailability `json:"availability,omitempty"`

	// MediaType is the type of media as classified by the site, such as
	// "episode", "clip", or "trailer".
	MediaType *string `json:"media_type,omitempty"`

	//
	// Chapter data available when the video belongs to some logical chapter or
	// section.
	//

	// Chapter is the name or title of the chapter the video belongs to.
	Chapter *string `json:"chapter,omitempty"`

	// ChapterNumber is the number of the chapter the video belongs to.
	ChapterNumber *float64 `json:"chapter_number,omitempty"`

	// ChapterID is the ID of the chapter the video belongs to.
	ChapterID *string `json:"chapter_id,omitempty"`

	// Playlist is the name or id of the playlist that contains the video.
	Playlist *string `json:"playlist,omitempty"`

	// PlaylistIndex is the index of the video in the playlist.
	PlaylistIndex *int `json:"playlist_index,omitempty"`

	// PlaylistID is the playlist identifier.
	PlaylistID *string `json:"playlist_id,omitempty"`

	// PlaylistTitle is the playlist title.
	PlaylistTitle *string `json:"playlist_title,omitempty"`

	// PlaylistUploader is the full name of the playlist uploader.
	PlaylistUploader *string `json:"playlist_uploader,omitempty"`

	// PlaylistUploaderID is the nickname or id of the playlist uploader.
	PlaylistUploaderID *string `json:"playlist_uploader_id,omitempty"`

	// PlaylistCount is auto-generated by yt-dlp. It is the total number of videos
	// in the playlist.
	PlaylistCount *int `json:"playlist_count,omitempty"`

	//
	// Series data available when the video is an episode of some series, programme
	// or podcast.
	//

	// Series is the title of the series or programme the video episode belongs to.
	Series *string `json:"series,omitempty"`

	// SeriesID is the ID of the series or programme the video episode belongs to.
	SeriesID *string `json:"series_id,omitempty"`

	// Season is the title of the season the video episode belongs to.
	Season *string `json:"season,omitempty"`

	// SeasonNumber is the number of the season the video episode belongs to.
	SeasonNumber *float64 `json:"season_number,omitempty"`

	// SeasonID is the ID of the season the video episode belongs to.
	SeasonID *string `json:"season_id,omitempty"`

	// Episode is the title of the video episode. Unlike mandatory video title field,
	// this field should denote the exact title of the video episode without any
	// kind of decoration.
	Episode *string `json:"episode,omitempty"`

	// EpisodeNumber is the number of the video episode within a season.
	EpisodeNumber *float64 `json:"episode_number,omitempty"`

	// EpisodeID is the ID of the video episode.
	EpisodeID *string `json:"episode_id,omitempty"`

	//
	// Track data available when the media is a track or a part of a music album.
	//

	// Track is the title of the track.
	Track *string `json:"track,omitempty"`

	// TrackNumber is the number of the track within an album or a disc.
	TrackNumber *float64 `json:"track_number,omitempty"`

	// TrackID is the ID of the track (useful in case of custom indexing, e.g. 6.iii).
	TrackID *string `json:"track_id,omitempty"`

	// Artists contains the artist(s) of the track.
	Artists []string `json:"artists,omitempty"`

	// Artist is the artist(s) of the track, comma-separated.
	//
	// Deprecated: use Artists instead.
	Artist *string `json:"artist,omitempty"`

	// Composers contains the composer(s) of the piece.
	Composers []string `json:"composers,omitempty"`

	// Genres contains the genre(s) of the track.
	Genres []string `json:"genres,omitempty"`

	// Genre is the genre(s) of the track, comma-separated.
	//
	// Deprecated: use Genres instead.
	Genre *string `json:"genre,omitempty"`

	// Album is the title of the album the track belongs to.
	Album *string `json:"album,omitempty"`

	// AlbumType is the type of the album (e.g. "Demo", "Full-length", "Split", "Compilation", etc).
	AlbumType *string `json:"album_type,omitempty"`

	// AlbumArtists contains all artists who appeared on the album. It is useful
	// for splits and compilations.
	AlbumArtists []string `json:"album_artists,omitempty"`

	// AlbumArtist contains all artists who appeared on the album, comma-separated.
	//
	// Deprecated: use AlbumArtists instead.
	AlbumArtist *string `json:"album_artist,omitempty"`

	// DiscNumber is the number of the disc or other physical medium the track belongs to.
	DiscNumber *float64 `json:"disc_number,omitempty"`

	// Composer is the composer(s) of the piece, comma-separated.
	//
	// Deprecated: use Composers instead.
	Composer *string `json:"composer,omitempty"`

	//
	// Clip data available when the media is a clip that should be cut from the
	// original video.
	//

	// SectionStart is the start time of the section in seconds.
	SectionStart *float64 `json:"section_start,omitempty"`

	// SectionEnd is the end time of the section in seconds.
	SectionEnd *float64 `json:"section_end,omitempty"`

	//
	// Storyboard data available when the media is a storyboard.
	//

	// Rows is the number of rows in each storyboard fragment, as an integer.
	Rows *int `json:"rows,omitempty"`

	// Columns is the number of columns in each storyboard fragment, as an integer.
	Columns *int `json:"columns,omitempty"`

	//
	// Other misc auto-generated fields by yt-dlp.
	//

	// Extractor is the name of the extractor.
	Extractor *string `json:"extractor,omitempty"`

	// ExtractorKey is the key name of the extractor.
	ExtractorKey *string `json:"extractor_key,omitempty"`

	// WebpageURLBasename is the basename of [ExtractedInfo.WebpageURL].
	WebpageURLBasename *string `json:"webpage_url_basename,omitempty"`

	// WebpageURLDomain is the domain name of [ExtractedInfo.WebpageURL].
	WebpageURLDomain *string `json:"webpage_url_domain,omitempty"`

	// Autonumber is a five-digit number that will be increased with each download,
	// starting at the configured autonumber start (zero by default).
	Autonumber int `json:"autonumber"`

	// Epoch is the Unix epoch when information extraction was completed.
	Epoch *float64 `json:"epoch,omitempty"`

	// Entries contains playlist or multi-video entries when [ExtractedInfo.Type]
	// is [ExtractedTypePlaylist] or [ExtractedTypeMultiVideo].
	Entries []*ExtractedInfo `json:"entries"`
}

// ExtractedType identifies the type of an extracted result.
type ExtractedType string

const (
	ExtractedTypeAny            ExtractedType = "any"
	ExtractedTypeSingle         ExtractedType = "single"
	ExtractedTypeVideo          ExtractedType = "video"
	ExtractedTypePlaylist       ExtractedType = "playlist"
	ExtractedTypeMultiVideo     ExtractedType = "multi_video"
	ExtractedTypeURL            ExtractedType = "url"
	ExtractedTypeURLTransparent ExtractedType = "url_transparent"
)

// ExtractedLiveStatus describes the live-stream state of a video.
type ExtractedLiveStatus string

const (
	ExtractedLiveStatusIsLive     ExtractedLiveStatus = "is_live"
	ExtractedLiveStatusIsUpcoming ExtractedLiveStatus = "is_upcoming"
	ExtractedLiveStatusWasLive    ExtractedLiveStatus = "was_live"
	ExtractedLiveStatusNotLive    ExtractedLiveStatus = "not_live"
	ExtractedLiveStatusPostLive   ExtractedLiveStatus = "post_live" // Was live, but VOD is not yet processed.
)

// ExtractedAvailability describes the access condition of a video.
type ExtractedAvailability string

const (
	ExtractedAvailabilityPrivate        ExtractedAvailability = "private"
	ExtractedAvailabilityPremiumOnly    ExtractedAvailability = "premium_only"
	ExtractedAvailabilitySubscriberOnly ExtractedAvailability = "subscriber_only"
	ExtractedAvailabilityNeedsAuth      ExtractedAvailability = "needs_auth"
	ExtractedAvailabilityUnlisted       ExtractedAvailability = "unlisted"
	ExtractedAvailabilityPublic         ExtractedAvailability = "public"
)

// ExtractedHLSAES contains HLS AES-128 decryption information used by the
// native HLS downloader.
type ExtractedHLSAES struct {
	// URI is the URI from which the decryption key is downloaded.
	URI *string `json:"uri,omitempty"`

	// Key is the hexadecimal key used to decrypt fragments. When set, any key
	// URI is ignored.
	Key *string `json:"key,omitempty"`

	// IV is the hexadecimal initialization vector used to decrypt fragments.
	IV *string `json:"iv,omitempty"`
}

// ExtractedFormat is format information returned by yt-dlp.
//
// Some intentionally excluded fields:
//   - ws (not JSON-serializable).
//   - downloader_options (internal use only).
type ExtractedFormat struct {
	// URL is the URL representing the media:
	//   - plain file media: HTTP URL of this file.
	//   - RTMP: RTMP URL.
	//   - HLS: URL of the M3U8 media playlist.
	//   - HDS: URL of the F4M manifest.
	//   - DASH:
	//       - HTTP URL to plain file media (in case of
	//         unfragmented media)
	//       - URL of the MPD manifest or base URL
	//         representing the media if MPD manifest
	//         is parsed from a string (in case of
	//         fragmented media)
	//   - MSS: URL of the ISM manifest.
	URL string `json:"url"`

	// RequestData is the data to send in a POST request to the URL.
	RequestData *string `json:"request_data,omitempty"`

	// ManifestStreamNumber is the index of the stream in the manifest file.
	//
	// This field is for internal use by yt-dlp.
	ManifestStreamNumber *int `json:"manifest_stream_number,omitempty"`

	// ManifestURL is the URL of the manifest file in case of fragmented media:
	//   - HLS: URL of the M3U8 master playlist.
	//   - HDS: URL of the F4M manifest.
	//   - DASH: URL of the MPD manifest.
	//   - MSS: URL of the ISM manifest.
	ManifestURL *string `json:"manifest_url,omitempty"`

	// Extension is the video filename extension. Will be calculated from URL if missing.
	Extension *string `json:"ext,omitempty"`

	// A human-readable description of the format ("mp4 container with h264/opus").
	// Calculated from the format_id, width, height. and format_note fields if missing.
	Format *string `json:"format,omitempty"`

	// FormatID is a short description of the format ("mp4_h264_opus" or "19").
	// Technically optional, but strongly recommended.
	FormatID *string `json:"format_id,omitempty"`

	// FormatNote is additional info about the format ("3D" or "DASH video").
	FormatNote *string `json:"format_note,omitempty"`

	// Width of the video, if known.
	Width *float64 `json:"width,omitempty"`

	// Height of the video, if known.
	Height *float64 `json:"height,omitempty"`

	// AspectRatio is the aspect ratio of the video, if known. Automatically calculated
	// from width and height (e.g. 1.78, which would be 16:9).
	AspectRatio *float64 `json:"aspect_ratio,omitempty"`

	// Resolution is the textual description of width and height. Automatically
	// calculated from width and height.
	Resolution *string `json:"resolution,omitempty"`

	// DynamicRange is the dynamic range of the video, such as "SDR", "HDR10",
	// "HDR10+", "HDR12", "HLG", or "DV".
	DynamicRange *string `json:"dynamic_range,omitempty"`

	// TBR is the average bitrate of audio and video in KBit/s.
	TBR *float64 `json:"tbr,omitempty"`

	// ABR is the average audio bitrate in KBit/s.
	ABR *float64 `json:"abr,omitempty"`

	// ACodec is the name of the audio codec in use.
	ACodec *string `json:"acodec,omitempty"`

	// ASR is the audio sampling rate in Hertz.
	ASR *float64 `json:"asr,omitempty"`

	// AudioChannels contains the number of audio channels.
	AudioChannels *float64 `json:"audio_channels,omitempty"`

	// VBR is the average video bitrate in KBit/s.
	VBR *float64 `json:"vbr,omitempty"`

	// FPS is the framerate per second.
	FPS *float64 `json:"fps,omitempty"`

	// VCodec is the name of the video codec in use.
	VCodec *string `json:"vcodec,omitempty"`

	// Container is the name of the container format.
	Container *string `json:"container,omitempty"`

	// FileSize is the number of bytes, if known in advance.
	FileSize *int `json:"filesize,omitempty"`

	// FileSizeApprox is an estimate for the number of bytes.
	FileSizeApprox *int `json:"filesize_approx,omitempty"`

	// PlayerURL is the SWF Player URL (used for rtmpdump).
	PlayerURL *string `json:"player_url,omitempty"`

	// Protocol is the protocol that will be used for the actual download,
	// lower-case. One of "http", "https" or one of the protocols defined in
	// yt-dlp's downloader.PROTOCOL_MAP.
	Protocol *string `json:"protocol,omitempty"`

	// FragmentBaseURL is the base URL for fragments. Each fragment's path value
	// (if present) will be relative to this URL.
	FragmentBaseURL *string `json:"fragment_base_url,omitempty"`

	// Fragments is a list of fragments of a fragmented media. Each fragment entry
	// must contain either an url or a path. If an url is present it should be
	// considered by a client. Otherwise both path and [ExtractedFormat.FragmentBaseURL]
	// must be present.
	Fragments []*ExtractedFragment `json:"fragments,omitempty"`

	// HLSMediaPlaylistData is the M3U8 media playlist data when it must be
	// modified during extraction and the native HLS downloader should bypass
	// requesting the URL.
	HLSMediaPlaylistData *string `json:"hls_media_playlist_data,omitempty"`

	// IsFromStart is true if it's a live format that can be downloaded from the start.
	IsFromStart *bool `json:"is_from_start,omitempty"`

	// Preference is the order number of this format. If this field is present,
	// the formats get sorted by this field, regardless of all other values.
	//   - -1 for default (order by other properties).
	//   - -2 or smaller for less than default.
	//   - < -1000 to hide the format (if there is another one which is strictly better)
	Preference *int `json:"preference,omitempty"`

	// SourcePreference is the order number for this video source (quality takes
	// higher priority)
	//   - -1 for default (order by other properties).
	//   - -2 or smaller for less than default.
	SourcePreference *int `json:"source_preference,omitempty"`

	// Language is the language code, e.g. "de" or "en-US".
	Language *string `json:"language,omitempty"`

	// LanguagePreference is the order number of this language, based off a few factors.
	// Is this in the language mentioned in the URL?
	//   - 10 if it's what the URL is about.
	//   - -1 for default (don't know).
	//   - -10 otherwise, other values reserved for now.
	LanguagePreference *int `json:"language_preference,omitempty"`

	// Quality is the order number of the video quality of this format, irrespective
	// of the file format.
	//   - -1 for default (order by other properties).
	//   - -2 or smaller for less than default.
	Quality *float64 `json:"quality,omitempty"`

	// HTTPHeaders are additional HTTP headers to be sent with the request.
	HTTPHeaders map[string]string `json:"http_headers,omitempty"`

	// StretchRatio if given and not 1, indicates that the video's pixels are not
	// square. "width:height" ratio as float.
	StretchedRatio *float64 `json:"stretched_ratio,omitempty"`

	// NoResume is true if the download for this format cannot be resumed (HTTP or RTMP).
	NoResume *bool `json:"no_resume,omitempty"`

	// HasDRM is true if the format is DRM-protected and cannot be downloaded.
	// 'maybe' if the format may have DRM and has to be tested before download.
	HasDRM any `json:"has_drm,omitempty"`

	// ExtraParamToSegmentURL is a query string to append to each fragment's URL,
	// or to update each existing query string with. For HLS AES-128 streams, it
	// is also applied to the key URL unless [ExtractedFormat.ExtraParamToKeyURL]
	// is set. Only applied by the native HLS/DASH downloaders.
	ExtraParamToSegmentURL *string `json:"extra_param_to_segment_url,omitempty"`

	// ExtraParamToKeyURL is a query string to append to the URL of the format's
	// HLS AES-128 decryption key. Only applied by the native HLS downloader.
	ExtraParamToKeyURL *string `json:"extra_param_to_key_url,omitempty"`

	// HLSAES contains HLS AES-128 decryption information used by the native HLS
	// downloader to override values in the media playlist.
	HLSAES *ExtractedHLSAES `json:"hls_aes,omitempty"`

	// Impersonate identifies the browser client or target to impersonate when
	// requesting this format. In sanitized yt-dlp JSON, this is a bool, string,
	// or list of strings; ImpersonateTarget objects are serialized as strings.
	Impersonate any `json:"impersonate,omitempty"`

	// AvailableAt is the Unix timestamp when this format will be available to download.
	AvailableAt *float64 `json:"available_at,omitempty"`

	// HTTPChunkSize is the chunk size for HTTP downloads, in bytes.
	HTTPChunkSize *int `json:"http_chunk_size,omitempty"`

	// FFmpegArgs contains extra arguments for the ffmpeg downloader input.
	FFmpegArgs []string `json:"ffmpeg_args,omitempty"`

	// FFmpegArgsOut contains extra arguments for the ffmpeg downloader output.
	FFmpegArgsOut []string `json:"ffmpeg_args_out,omitempty"`

	// WSURL is the WebSocket URL used by the NiconicoLive downloader.
	WSURL *string `json:"ws_url,omitempty"`

	// MaxQuality is the maximum stream quality used by the NiconicoLive downloader.
	MaxQuality *string `json:"max_quality,omitempty"`

	// IsDASHPeriods is true when the format was created by merging multiple DASH
	// periods.
	IsDASHPeriods *bool `json:"is_dash_periods,omitempty"`

	//
	// Storyboard data available when the media is a storyboard.
	//

	// Rows is the number of rows in each storyboard fragment, as an integer.
	Rows *int `json:"rows,omitempty"`

	// Columns is the number of columns in each storyboard fragment, as an integer.
	Columns *int `json:"columns,omitempty"`

	//
	// Additional RTMP-specific information (if the protocol is RTMP).
	//

	// PageURL is the URL of the page containing the video.
	PageURL *string `json:"page_url,omitempty"`

	// App is the application to play this stream.
	App *string `json:"app,omitempty"`

	// PlayPath is the play path for this stream.
	PlayPath *string `json:"play_path,omitempty"`

	// TCURL is the TC URL for this stream.
	TCURL *string `json:"tc_url,omitempty"`

	// FlashVersion is the flash version for this stream.
	FlashVersion *string `json:"flash_version,omitempty"`

	// RTMPLive is true if this is a live stream.
	RTMPLive *bool `json:"rtmp_live,omitempty"`

	// RTMPConn is the RTMP connection for this stream.
	RTMPConn *string `json:"rtmp_conn,omitempty"`

	// RTMPProtocol is the RTMP protocol for this stream.
	RTMPProtocol *string `json:"rtmp_protocol,omitempty"`

	// RTMPRealTime is true if this is a real-time stream.
	RTMPRealTime *bool `json:"rtmp_real_time,omitempty"`
}

// ExtractedFragment describes one fragment of fragmented media.
type ExtractedFragment struct {
	// URL is the URL of the fragment. Either URL or Path is present.
	URL *string `json:"url,omitempty"`

	// Path is the path of the fragment, relative to [ExtractedFormat.FragmentBaseURL].
	Path *string `json:"path,omitempty"`

	// Duration is the optional duration of the fragment in seconds.
	Duration *float64 `json:"duration,omitempty"`

	// FileSize is the size of the fragment in bytes.
	FileSize *int `json:"filesize,omitempty"`
}

// ExtractedSubtitle describes one available subtitle format.
type ExtractedSubtitle struct {
	// Extension is the subtitle file extension. It is calculated from URL when
	// missing.
	Extension *string `json:"ext,omitempty"`

	// URL is the URL of the subtitle file. Either URL or Data is present.
	URL *string `json:"url,omitempty"`

	// Data contains the subtitle file contents. Either Data or URL is present.
	Data *string `json:"data,omitempty"`

	// Name or description of the subtitle.
	Name *string `json:"name,omitempty"`

	// HTTPHeaders are additional HTTP headers to be sent with the request.
	HTTPHeaders map[string]string `json:"http_headers,omitempty"`

	// Impersonate identifies the browser client or target to impersonate when
	// requesting this subtitle. In sanitized yt-dlp JSON, this is a bool, string,
	// or list of strings; ImpersonateTarget objects are serialized as strings.
	Impersonate any `json:"impersonate,omitempty"`
}

// ExtractedThumbnail describes one available thumbnail format.
type ExtractedThumbnail struct {
	// ID is the thumbnail format ID.
	ID *string `json:"id,omitempty"`

	// URL of the thumbnail.
	URL string `json:"url"`

	// Extension is the actual image extension, if it is not given in URL.
	Extension *string `json:"ext,omitempty"`

	// Preference is the quality ordering of the image.
	Preference *int `json:"preference,omitempty"`

	// Width of the thumbnail.
	Width *int `json:"width,omitempty"`

	// Height of the thumbnail.
	Height *int `json:"height,omitempty"`

	// Deprecated: Resolution is the textual description of width and height as "WIDTHxHEIGHT".
	Resolution *string `json:"resolution,omitempty"`

	// FileSize is the number of bytes, if known in advance.
	FileSize *int `json:"filesize,omitempty"`

	// HTTPHeaders are additional HTTP headers to be sent with the request.
	HTTPHeaders map[string]string `json:"http_headers,omitempty"`
}

// ExtractedVersion contains the yt-dlp version metadata included in results.
type ExtractedVersion struct {
	// CurrentGitHead is the git commit hash of the yt-dlp install that returned
	// this data.
	CurrentGitHead *string `json:"current_git_head,omitempty"`

	// ReleaseGitHead is the git commit hash of the currently available yt-dlp
	// release.
	ReleaseGitHead *string `json:"release_git_head,omitempty"`

	// Repository is the name of the repository where yt-dlp is hosted.
	Repository *string `json:"repository,omitempty"`

	// Version is the version number of the yt-dlp install that returned this data.
	Version *string `json:"version,omitempty"`
}

// ExtractedChapterData describes one video chapter.
type ExtractedChapterData struct {
	// StartTime of the chapter in seconds.
	StartTime *float64 `json:"start_time,omitempty"`

	// EndTime of the chapter in seconds.
	EndTime *float64 `json:"end_time,omitempty"`

	// Title of the chapter.
	Title *string `json:"title,omitempty"`
}

// ExtractedHeatmapData describes one point in a video's heatmap.
type ExtractedHeatmapData struct {
	// StartTime of the data point in seconds.
	StartTime *float64 `json:"start_time,omitempty"`

	// EndTime of the data point in seconds.
	EndTime *float64 `json:"end_time,omitempty"`

	// Value is the normalized value of the data point (float between 0 and 1).
	Value *float64 `json:"value,omitempty"`
}

// ExtractedVideoComment describes one video comment.
type ExtractedVideoComment struct {
	// Author is the human-readable name of the comment author.
	Author *string `json:"author,omitempty"`

	// AuthorID is the user ID of the comment author.
	AuthorID *string `json:"author_id,omitempty"`

	// AuthorThumbnail is the thumbnail of the comment author.
	AuthorThumbnail *string `json:"author_thumbnail,omitempty"`

	// AuthorURL is the URL to the comment author's page.
	AuthorURL *string `json:"author_url,omitempty"`

	// AuthorIsVerified is true if the author is verified on the platform.
	AuthorIsVerified *bool `json:"author_is_verified,omitempty"`

	// AuthorIsUploader is true if the comment is made by the video uploader.
	AuthorIsUploader *bool `json:"author_is_uploader,omitempty"`

	// ID is the comment ID.
	ID *string `json:"id,omitempty"`

	// HTML is the comment as HTML.
	HTML *string `json:"html,omitempty"`

	// Text is the plain text of the comment.
	Text *string `json:"text,omitempty"`

	// Timestamp is the UNIX timestamp of comment.
	Timestamp *float64 `json:"timestamp,omitempty"`

	// Parent is the ID of the comment this one is replying to. Set to "root" to
	// indicate that this is a comment to the original video.
	Parent *string `json:"parent,omitempty"`

	// LikeCount is the number of positive ratings of the comment.
	LikeCount *float64 `json:"like_count,omitempty"`

	// DislikeCount is the number of negative ratings of the comment.
	DislikeCount *float64 `json:"dislike_count,omitempty"`

	// IsFavorited is true if the comment is marked as favorite by the video uploader.
	IsFavorited *bool `json:"is_favorited,omitempty"`

	// IsPinned is true if the comment is pinned to the top of the comments.
	IsPinned *bool `json:"is_pinned,omitempty"`

	// StartTime is the time in seconds when the comment should be displayed.
	StartTime *float64 `json:"start_time,omitempty"`

	// EndTime is the time in seconds when the comment should stop being displayed.
	EndTime *float64 `json:"end_time,omitempty"`
}
