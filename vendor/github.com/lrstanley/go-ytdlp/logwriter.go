// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"slices"
	"sync"
	"time"
	"unicode"
)

// Pipe identifies the log stream for a [ResultLog].
type Pipe string

const (
	PipeStdout Pipe = "stdout"
	PipeStderr Pipe = "stderr"
)

// ResultLog is a single stdout or stderr line from yt-dlp.
type ResultLog struct {
	// The timestamp of the log line as it was received from yt-dlp.
	Timestamp time.Time `json:"timestamp"`
	// The raw log line.
	Line string `json:"line"`
	// May be nil if the log line wasn't valid JSON.
	JSON *jsontext.Value `json:"json,omitempty"`
	// [PipeStdout] or [PipeStderr].
	Pipe Pipe `json:"pipe"`
	// Overwrite is true when the line was terminated by \r (ffmpeg in-place
	// terminal progress). These are delivered to [Command.LogFunc] but are
	// not stored in [Result.OutputLogs] or [Result.Stderr].
	Overwrite bool `json:"overwrite,omitempty"`
}

func (r *ResultLog) asString(timestamps, maskJSON bool) string {
	line := r.Line

	if maskJSON && r.JSON != nil {
		line = "<json-data>"
	}

	if timestamps {
		return fmt.Sprintf("[%s::%s] %s", r.Timestamp.Format(time.DateTime), r.Pipe, line)
	}

	return line
}

func (r *ResultLog) String() string {
	return r.asString(true, true)
}

// LogCallbackFunc is called for each stdout/stderr log line from yt-dlp,
// excluding progress-prefix JSON consumed by [Command.ProgressFunc]. It
// runs on the pipe's Write and must not block.
//
// Stderr lines terminated by \r (ffmpeg in-place progress) are delivered
// with [ResultLog.Overwrite] set and are not stored on [Result].
//
// Callbacks are serialized across stdout and stderr (never concurrent).
// Same-pipe order matches yt-dlp. Cross-pipe order is commit order:
// the shared lock covers LogFunc only, so the other pipe can still Write.
type LogCallbackFunc func(log *ResultLog)

// LogFunc registers a callback for real yt-dlp log lines as they arrive.
// Progress JSON is not delivered here; use [Command.ProgressFunc].
// ffmpeg \r progress on stderr is delivered with [ResultLog.Overwrite].
//
// Unlike [Command.ProgressFunc], no yt-dlp flags are injected.
//
//   - See [Command.UnsetLogFunc] for unsetting the log function.
func (c *Command) LogFunc(fn LogCallbackFunc) *Command {
	c.mu.Lock()
	c.log = fn
	c.mu.Unlock()
	return c
}

// UnsetLogFunc removes the log callback that was previously set with
// [Command.LogFunc].
func (c *Command) UnsetLogFunc() *Command {
	c.mu.Lock()
	c.log = nil
	c.mu.Unlock()
	return c
}

// StderrCallbackFunc is a callback function that is called when stderr output
// is received from yt-dlp. Each invocation receives a single line of output.
//
// Deprecated: Use [LogCallbackFunc] with [Command.LogFunc] and filter on
// [ResultLog.Pipe].
type StderrCallbackFunc func(line string)

// StderrFunc registers a callback for stderr lines from yt-dlp.
//
// Deprecated: Use [Command.LogFunc] and filter on [ResultLog.Pipe].
//
//go:fix inline
func (c *Command) StderrFunc(fn StderrCallbackFunc) *Command {
	return c.LogFunc(func(log *ResultLog) {
		if log.Pipe == PipeStderr {
			fn(log.Line)
		}
	})
}

// UnsetStderrFunc removes the callback previously set with [Command.StderrFunc].
//
// Deprecated: Use [Command.UnsetLogFunc].
//
//go:fix inline
func (c *Command) UnsetStderrFunc() *Command {
	return c.UnsetLogFunc()
}

// logSink serializes LogFunc across stdout and stderr writers.
type logSink struct {
	mu sync.Mutex
	fn LogCallbackFunc
}

func (s *logSink) log(r *ResultLog) {
	if s == nil {
		return
	}
	s.mu.Lock()
	s.fn(r)
	s.mu.Unlock()
}

type timestampWriter struct {
	checkJSON bool
	pipe      Pipe
	progress  *progressHandler
	log       *logSink

	buf            bytes.Buffer
	lastWriteStart time.Time
	results        []*ResultLog
}

func (w *timestampWriter) Write(p []byte) (n int, err error) {
	n = len(p)

	for len(p) > 0 {
		if w.lastWriteStart.IsZero() {
			w.lastWriteStart = time.Now()
		}

		// Prefer newline over carriage return so a chunk that contains both
		// is one log line, matching ordinary stdout/stderr.
		if i := bytes.IndexByte(p, '\n'); i >= 0 {
			_, _ = w.buf.Write(p[:i+1])
			w.flush()
			p = p[i+1:]
			continue
		}

		if w.pipe == PipeStderr {
			if i := bytes.IndexByte(p, '\r'); i >= 0 {
				_, _ = w.buf.Write(p[:i+1])
				w.flushCR()
				p = p[i+1:]
				continue
			}
		}

		_, err = w.buf.Write(p)
		return n, err
	}

	return n, nil
}

func (w *timestampWriter) flush() {
	if w.buf.Len() == 0 {
		return
	}

	line := bytes.TrimRightFunc(w.buf.Bytes(), unicode.IsSpace)

	if w.progress != nil {
		if v, ok := bytes.CutPrefix(line, progressPrefix); ok {
			var raw jsontext.Value
			if err := json.Unmarshal(v, &raw); err == nil {
				w.progress.parse(raw)
				w.lastWriteStart = time.Time{}
				w.buf.Reset()
				return
			}
		}
	}

	result := &ResultLog{
		Timestamp: w.lastWriteStart,
		Line:      string(line),
		Pipe:      w.pipe,
	}

	if w.checkJSON && len(line) > 0 {
		var raw jsontext.Value
		if err := json.Unmarshal(line, &raw); err == nil {
			result.JSON = &raw
		}
	}

	w.results = append(w.results, result)
	w.log.log(result)

	w.lastWriteStart = time.Time{}
	w.buf.Reset()
}

// flushCR delivers a \r-terminated stderr line to LogFunc only. These
// ephemeral ffmpeg progress frames are not stored on the result.
func (w *timestampWriter) flushCR() {
	if w.buf.Len() == 0 {
		return
	}

	line := string(bytes.TrimSpace(w.buf.Bytes()))
	if line != "" {
		w.log.log(&ResultLog{
			Timestamp: w.lastWriteStart,
			Line:      line,
			Pipe:      w.pipe,
			Overwrite: true,
		})
	}

	w.lastWriteStart = time.Time{}
	w.buf.Reset()
}

func (w *timestampWriter) mergeResults(otherWriters ...*timestampWriter) []*ResultLog {
	w.flush()

	resultCount := len(w.results)
	for _, other := range otherWriters {
		resultCount += len(other.results)
	}

	results := make([]*ResultLog, 0, resultCount)
	results = append(results, w.results...)

	for _, other := range otherWriters {
		results = append(results, other.results...)
	}

	slices.SortFunc(results, func(a, b *ResultLog) int {
		return a.Timestamp.Compare(b.Timestamp)
	})
	return results
}

func (w *timestampWriter) String() string {
	w.flush()

	var buf bytes.Buffer
	for i, r := range w.results {
		buf.WriteString(r.Line)
		if i < len(w.results)-1 {
			buf.WriteByte('\n')
		}
	}
	return buf.String()
}
