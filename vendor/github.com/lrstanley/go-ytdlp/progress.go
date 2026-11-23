// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

var progressPrefix = []byte("progress:")

const progressFormat = "%()j"

type progressData struct {
	Info     *ExtractedInfo `json:"info"`
	Progress struct {
		Status             ProgressStatus `json:"status"`
		TotalBytes         int            `json:"total_bytes,omitempty"`
		TotalBytesEstimate float64        `json:"total_bytes_estimate,omitempty"`
		DownloadedBytes    int            `json:"downloaded_bytes"`
		Filename           string         `json:"filename,omitempty"`
		TmpFilename        string         `json:"tmpfilename,omitempty"`
		FragmentIndex      int            `json:"fragment_index,omitempty"`
		FragmentCount      int            `json:"fragment_count,omitempty"`
		PostProcessor      string         `json:"postprocessor,omitempty"`
		// There are technically other fields, but these are the important ones.
	} `json:"progress"`
	AutoNumber      int `json:"autonumber,omitempty"`
	VideoAutoNumber int `json:"video_autonumber,omitempty"`
}

type progressHandler struct {
	fn ProgressCallbackFunc

	mu       sync.Mutex
	started  map[string]time.Time // Used to track multiple independent downloads.
	finished map[string]time.Time // Used to track multiple independent downloads.
}

func newProgressHandler(fn ProgressCallbackFunc) *progressHandler {
	h := &progressHandler{
		fn:       fn,
		started:  make(map[string]time.Time),
		finished: make(map[string]time.Time),
	}
	return h
}

func (h *progressHandler) parse(raw jsontext.Value) {
	data := &progressData{}

	err := json.Unmarshal(raw, data)
	if err != nil {
		return
	}

	cleanJSONValue(reflect.ValueOf(data))

	status := data.Progress.Status
	if data.Progress.PostProcessor != "" && status != ProgressStatusError {
		// yt-dlp post-processor hooks use started/processing/finished, which
		// would otherwise look like a download completing.
		status = ProgressStatusPostProcessing
	}

	update := ProgressUpdate{
		Info:            data.Info,
		Status:          status,
		TotalBytes:      data.Progress.TotalBytes,
		DownloadedBytes: data.Progress.DownloadedBytes,
		FragmentIndex:   data.Progress.FragmentIndex,
		FragmentCount:   data.Progress.FragmentCount,
		Filename:        data.Progress.Filename,
		PostProcessor:   data.Progress.PostProcessor,
	}

	if update.TotalBytes == 0 {
		update.TotalBytes = int(data.Progress.TotalBytesEstimate)
	}

	if update.Filename == "" {
		if data.Progress.TmpFilename != "" {
			update.Filename = data.Progress.TmpFilename
		} else if data.Info != nil && data.Info.Filename != nil && *data.Info.Filename != "" {
			update.Filename = *data.Info.Filename
		}
	}

	update.ID = update.generateID()

	var ok bool

	h.mu.Lock()
	update.Started, ok = h.started[update.ID]
	if !ok {
		update.Started = time.Now()
		h.started[update.ID] = update.Started
	}

	update.Finished, ok = h.finished[update.ID]
	if !ok && update.Status.IsCompletedType() {
		update.Finished = time.Now()
		h.finished[update.ID] = update.Finished
	}
	h.mu.Unlock()

	h.fn(update)
}

// ProgressStatus is the status of the download progress.
type ProgressStatus string

func (s ProgressStatus) IsCompletedType() bool {
	return s == ProgressStatusError || s == ProgressStatusFinished
}

const (
	// ProgressStatusStarting is unused. yt-dlp does not emit this status.
	//
	// Deprecated: ProgressStatusStarting is never reported. Handle
	// [ProgressStatusDownloading], [ProgressStatusFinished],
	// [ProgressStatusError], and [ProgressStatusPostProcessing] instead.
	ProgressStatusStarting ProgressStatus = "starting"
	// ProgressStatusDownloading is reported while a file is being written.
	// Separate formats (for example video then audio) each get their own
	// downloading updates; compare [ProgressUpdate.ID], not Filename.
	ProgressStatusDownloading ProgressStatus = "downloading"
	// ProgressStatusPostProcessing is reported for yt-dlp post-processor
	// events (merge, recode, metadata, and similar). yt-dlp emits
	// started/processing/finished for those hooks; they are mapped here so
	// they are not mistaken for a download finishing.
	ProgressStatusPostProcessing ProgressStatus = "post_processing"
	// ProgressStatusError is reported when yt-dlp fails to download a file.
	ProgressStatusError ProgressStatus = "error"
	// ProgressStatusFinished is reported when a download hook completes
	// successfully. Post-processor finished events are mapped to
	// [ProgressStatusPostProcessing] instead.
	ProgressStatusFinished ProgressStatus = "finished"
)

// ProgressCallbackFunc is invoked for each yt-dlp progress event.
// It runs synchronously inside yt-dlp's stdout Write and must not block:
// a slow callback stalls yt-dlp. It cannot fail the run.
type ProgressCallbackFunc func(update ProgressUpdate)

// ProgressUpdate is a point-in-time snapshot of the download progress.
type ProgressUpdate struct {
	Info *ExtractedInfo `json:"info"`

	// Status is the current status of this update. Download hooks use
	// downloading/error/finished; post-processor hooks are mapped to
	// [ProgressStatusPostProcessing].
	Status ProgressStatus `json:"status"`
	// TotalBytes is the size of the current file being downloaded, not the
	// final merged or recoded output. Separate formats (for example video
	// then audio) each report their own size. Zero if yt-dlp cannot
	// determine the size.
	TotalBytes int `json:"total_bytes"`
	// DownloadedBytes is how many bytes of the current file have been
	// written. These values are not monotonic across a whole run; use
	// [ProgressUpdate.ID], not Filename, to tell files apart.
	DownloadedBytes int `json:"downloaded_bytes"`
	// FragmentIndex is the index of the current fragment being downloaded.
	FragmentIndex int `json:"fragment_index,omitempty"`
	// FragmentCount is the total number of fragments in the download.
	FragmentCount int `json:"fragment_count,omitempty"`

	// Filename is the file currently being downloaded or processed. This is
	// often a temporary per-format path (for example video.f137.mp4, then
	// audio.f140.m4a), not the final merged destination. Filename alone is
	// not a stable identity; use [ProgressUpdate.ID].
	Filename string `json:"filename"`
	// ID is the stable identity for this update: filename, extractor id,
	// playlist id, playlist index, and post-processor name. Persist and
	// compare ID, not Filename; video, audio, and post-processors of the
	// same item can share similar names.
	ID string `json:"id,omitempty"`
	// PostProcessor is the yt-dlp post-processor name when Status is
	// [ProgressStatusPostProcessing] (for example Merger or
	// FFmpegVideoConvertor). Empty during downloads.
	PostProcessor string `json:"postprocessor,omitempty"`

	// Started is the time the download started.
	Started time.Time `json:"started"`
	// Finished is the time the download finished. If the download is still in progress,
	// this will be zero. You can validate with IsZero().
	Finished time.Time `json:"finished,omitempty"`
}

// Key returns [ProgressUpdate.ID].
//
// Deprecated: Use [ProgressUpdate.ID] instead.
//
//go:fix inline
func (p *ProgressUpdate) Key() string { return p.ID }

func (p *ProgressUpdate) generateID() string {
	unique := []string{p.Filename}

	if p.Info != nil {
		unique = append(unique, p.Info.ID)

		if p.Info.PlaylistID != nil {
			unique = append(unique, *p.Info.PlaylistID)
		}

		if p.Info.PlaylistIndex != nil {
			unique = append(unique, strconv.Itoa(*p.Info.PlaylistIndex))
		}
	}

	if p.PostProcessor != "" {
		unique = append(unique, p.PostProcessor)
	}

	return strings.Join(unique, ":")
}

// Duration returns the duration of the download. If the download is still in progress,
// it will return the time since the download started.
func (p *ProgressUpdate) Duration() time.Duration {
	if p.Finished.IsZero() {
		return time.Since(p.Started)
	}
	return p.Finished.Sub(p.Started)
}

// ETA returns the estimated time until the download is complete. If the download is
// complete, or hasn't started yet, it will return 0.
func (p *ProgressUpdate) ETA() time.Duration {
	perc := p.Percent()
	if perc == 0 || perc == 100 {
		return 0
	}
	return time.Duration(float64(p.Duration().Nanoseconds()) / perc * (100 - perc))
}

// Percent returns the percentage of the download that has been completed. If yt-dlp
// is unable to determine the total bytes, it will return 0.
func (p *ProgressUpdate) Percent() float64 {
	if p.Status.IsCompletedType() {
		return 100
	}
	if p.TotalBytes == 0 {
		return 0
	}
	return float64(p.DownloadedBytes) / float64(p.TotalBytes) * 100
}

// PercentString is like Percent, but returns a string representation of the percentage.
func (p *ProgressUpdate) PercentString() string {
	return fmt.Sprintf("%.2f%%", p.Percent())
}

// ProgressFunc registers a callback for yt-dlp progress events. yt-dlp
// reports progress per file (and separately for post-processing), not as
// one aggregate for the final output. Use [ProgressUpdate.ID], not
// Filename alone, to identify a file or post-processor.
//
// The callback runs on the stdout pump and must not block.
//
// ProgressFunc also sets [Command.ProgressTemplate] twice (download and
// postprocess) so both kinds of events are emitted.
//
// See [Command.UnsetProgressFunc] to unset the callback.
func (c *Command) ProgressFunc(frequency time.Duration, fn ProgressCallbackFunc) *Command {
	if frequency < 100*time.Millisecond {
		frequency = 100 * time.Millisecond
	}

	c.Progress().
		ProgressDelta(frequency.Seconds()).
		ProgressTemplate("download:" + string(progressPrefix) + progressFormat).
		ProgressTemplate("postprocess:" + string(progressPrefix) + progressFormat).
		Newline()

	c.mu.Lock()
	c.progress = newProgressHandler(fn)
	c.mu.Unlock()

	return c
}

// UnsetProgressFunc can be used to unset the progress function that was previously set
// with [Command.ProgressFunc].
func (c *Command) UnsetProgressFunc() *Command {
	c.mu.Lock()
	c.progress = nil
	c.mu.Unlock()

	return c
}
