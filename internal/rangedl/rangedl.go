// Package rangedl is the part-file mechanics of a parallel download: cutting a
// file into byte ranges, fetching one range with resume, and joining the part
// files back together. It knows nothing about which site a URL belongs to; the
// caller supplies the client, the request headers and the progress sink.
package rangedl

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// DefaultRetryDelay is the pause before a part resumes after a failed read.
const DefaultRetryDelay = 500 * time.Millisecond

// Sink is where a download reports the bytes it has received.
type Sink interface {
	AddProgressReceived(delta int64)
}

// Part names one byte range of a download and the file it is written to.
type Part struct {
	URL      string
	From, To int64 // inclusive
	Index    int   // part number; Combine joins parts by it
	Dest     string
}

// Options tune how a part is fetched; every field is optional.
type Options struct {
	Prepare    func(*http.Request) // sets the request headers a CDN expects
	Progress   Sink
	RetryDelay time.Duration
}

// ByteRange is an inclusive span of a file, as an HTTP Range header counts it.
type ByteRange struct{ From, To int64 }

// Split cuts a file of contentLength bytes into at most n contiguous
// ranges that cover it exactly once; the last range takes the remainder. There
// are never more ranges than bytes, so none is empty, and n below one means a
// single range. Parts are numbered by position, which is how Combine puts
// them back together.
func Split(contentLength int64, n int) []ByteRange {
	if contentLength <= 0 {
		return nil
	}
	n = max(n, 1)
	if int64(n) > contentLength {
		n = int(contentLength)
	}
	chunk := contentLength / int64(n)
	ranges := make([]ByteRange, n)
	for i := range ranges {
		from := int64(i) * chunk
		to := from + chunk - 1
		if i == n-1 {
			to = contentLength - 1
		}
		ranges[i] = ByteRange{From: from, To: to}
	}
	return ranges
}

// FetchPart downloads a part of the video file using HTTP Range Requests.
// Automatically retries with resume on connection drops (up to 20 stale retries).
func FetchPart(p Part, client *http.Client, opts Options) error {
	url, from, to, part, destPath := p.URL, p.From, p.To, p.Index, p.Dest
	retryDelay := opts.RetryDelay
	if retryDelay <= 0 {
		retryDelay = DefaultRetryDelay
	}
	partFilePath, err := PartPath(destPath, part)
	if err != nil {
		return err
	}
	// #nosec G304: path validated by PartPath to remain within destination directory
	file, err := os.Create(partFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			util.Logger.Warn("Error closing file", "error", err)
		}
	}()

	current := from
	maxStaleRetries := 20
	staleRetries := 0

	for attempt := 0; current <= to; attempt++ {
		if staleRetries >= maxStaleRetries {
			return fmt.Errorf("part %d: max retries (%d) exceeded at offset %d/%d", part, maxStaleRetries, current, to)
		}
		if attempt > 0 {
			util.Debugf("Download part %d: resuming at byte %d (attempt %d)", part, current, attempt+1)
			time.Sleep(retryDelay)
		}

		beforeRead := current

		req, err := http.NewRequest("GET", url, http.NoBody)
		if err != nil {
			return err
		}
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", current, to))
		if opts.Prepare != nil {
			opts.Prepare(req)
		}

		resp, err := client.Do(req) // #nosec G704
		if err != nil {
			util.Debugf("Download part %d: request error: %v", part, err)
			staleRetries++
			continue
		}

		if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
			statusCode := resp.StatusCode
			status := resp.Status
			if cErr := resp.Body.Close(); cErr != nil {
				util.Logger.Warn("Error closing response body", "error", cErr)
			}
			if statusCode == http.StatusForbidden || statusCode == http.StatusNotFound {
				return netx.NewDownloadExpiredError("Download", "http-range", statusCode, fmt.Errorf("HTTP %d: %s", statusCode, status))
			}
			util.Debugf("Download part %d: unexpected status %d", part, statusCode)
			staleRetries++
			continue
		}

		buf := make([]byte, 256*1024)
		for {
			n, readErr := resp.Body.Read(buf)
			if n > 0 {
				if _, writeErr := file.Write(buf[:n]); writeErr != nil {
					if cErr := resp.Body.Close(); cErr != nil {
						util.Logger.Warn("Error closing response body", "error", cErr)
					}
					return writeErr
				}
				current += int64(n)
				if opts.Progress != nil {
					opts.Progress.AddProgressReceived(int64(n))
				}
			}
			if readErr == io.EOF {
				if cErr := resp.Body.Close(); cErr != nil {
					util.Logger.Warn("Error closing response body", "error", cErr)
				}
				break
			}
			if readErr != nil {
				if cErr := resp.Body.Close(); cErr != nil {
					util.Logger.Warn("Error closing response body", "error", cErr)
				}
				util.Debugf("Download part %d: read error after %d bytes: %v", part, current-from, readErr)
				break
			}
		}

		if current > to {
			break
		}

		if current == beforeRead {
			staleRetries++
		} else {
			staleRetries = 0
		}
	}

	return nil
}

// Combine combines downloaded parts into a single final file.
func Combine(destPath string, numThreads int) error {
	outFile, err := os.Create(filepath.Clean(destPath))
	if err != nil {
		return err
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			util.Logger.Warn("Error closing output file", "error", err)
		}
	}()
	for i := range numThreads {
		partFilePath, err := PartPath(destPath, i)
		if err != nil {
			return err
		}
		// #nosec G304: path validated by PartPath to remain within destination directory
		partFile, err := os.Open(partFilePath)
		if err != nil {
			return err
		}
		if _, err := io.Copy(outFile, partFile); err != nil {
			if closeErr := partFile.Close(); closeErr != nil {
				util.Logger.Warn("Error closing part file", "error", closeErr)
			}
			return err
		}
		if err := partFile.Close(); err != nil {
			util.Logger.Warn("Error closing part file", "error", err)
		}
		if err := os.Remove(partFilePath); err != nil {
			return err
		}
	}
	return nil
}

// PartPath builds the part file path and ensures it stays within the destination directory
func PartPath(destPath string, part int) (string, error) {
	dir := filepath.Clean(filepath.Dir(destPath))
	base := filepath.Base(destPath)
	name := fmt.Sprintf("%s.part%d", base, part)
	joined := filepath.Join(dir, name)
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	absFile, err := filepath.Abs(joined)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absDir, absFile)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("invalid part path: %s", joined)
	}
	return joined, nil
}

// Download fetches a file of contentLength bytes as up to threads parallel
// parts and joins them into dest. fetch downloads one part (typically through
// FetchPart) and is called once per part, concurrently; its error stops the
// join. Download removes nothing on failure: the part files are what a retry
// resumes from.
func Download(contentLength int64, threads int, dest string, fetch func(Part) error) error {
	ranges := Split(contentLength, threads)
	var wg sync.WaitGroup
	errs := make(chan error, len(ranges))
	for i, r := range ranges {
		wg.Add(1)
		go func(p Part) {
			defer wg.Done()
			if err := fetch(p); err != nil {
				errs <- err
			}
		}(Part{From: r.From, To: r.To, Index: i, Dest: dest})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		return fmt.Errorf("chunk download failed: %w", err)
	}
	if err := Combine(dest, len(ranges)); err != nil {
		return fmt.Errorf("failed to combine parts: %w", err)
	}
	return nil
}
