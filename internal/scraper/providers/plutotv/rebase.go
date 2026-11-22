package plutotv

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// A title read out of the stitched playlist is several ranges. In most titles
// the ranges are cut from one encode and their timestamps run on. In some each
// range is a file of its own, its clock restarting at about one second, and a
// player that seeks into such a title looks for a timestamp that no range ever
// reaches: it freezes (Jumper, 2026-11-21). For those titles the player is not
// handed the CDN's segments but ours: the same bytes, decrypted, with each
// range's offset added to the timestamps inside, so the title is on one
// timeline. Titles that already are on one are left alone, and played from the
// CDN as before.

const (
	// rebaseThreshold is how far a range may start from where the playlist's
	// durations put it before it is taken for a restarted clock: whole seconds
	// of rounding drift on a continuous title, hundreds on a restarted one.
	rebaseThreshold = 30 * 90000

	// probeBytes is how much of a range's first segment is read to learn where
	// its clock starts. A multiple of the cipher block.
	probeBytes = 64 << 10

	// maxSegmentBytes bounds what one segment may weigh.
	maxSegmentBytes = 64 << 20

	// segmentPrefetch is how many segments ahead of the one asked for are
	// fetched while the player is busy with it.
	segmentPrefetch = 2

	// segmentKeepBehind is how many segments behind the one asked for stay in
	// memory: the player often asks for one again.
	segmentKeepBehind = 2
)

var (
	keyURIRe = regexp.MustCompile(`URI="([^"]+)"`)
	keyIVRe  = regexp.MustCompile(`IV=0[xX]([0-9A-Fa-f]{1,32})`)
)

// segmentRef is one segment of a rebased title: where it is, what opens it and
// how far its timestamps move.
type segmentRef struct {
	url, keyURL string
	iv          [16]byte
	shift       int64 // 90 kHz ticks
	extinf      string
	group       int
}

// fetchFn reads a URL's body, at most limit bytes of it. With prefix it asks
// for the first limit bytes only and stops there, whatever the server does
// with the Range header; without, a body over limit is an error.
type fetchFn func(ctx context.Context, url string, limit int64, prefix bool) ([]byte, error)

// rebasePlan is a title whose ranges are put on one timeline.
type rebasePlan struct {
	header string // the playlist's tags before the first segment
	segs   []segmentRef
	fetch  fetchFn
}

// render is the playlist the player is given, its segments addressed under id.
func (p *rebasePlan) render(id string) string {
	var b strings.Builder
	b.WriteString(p.header)
	for i, s := range p.segs {
		if i > 0 && s.group != p.segs[i-1].group {
			b.WriteString("#EXT-X-DISCONTINUITY\n")
		}
		b.WriteString(s.extinf + "\n")
		fmt.Fprintf(&b, "%s/%d.ts\n", id, i)
	}
	b.WriteString("#EXT-X-ENDLIST\n")
	return b.String()
}

// parseTitlePlaylist reads the playlist titlePlaylist makes: tags, then for
// each segment its key, its duration and its URL, a range apart wherever a
// discontinuity is.
func parseTitlePlaylist(playlist string) (header string, segs []segmentRef, err error) {
	var head strings.Builder
	var key, extinf string
	group := 0
	for _, line := range strings.Split(strings.ReplaceAll(playlist, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case line == "" || line == "#EXT-X-ENDLIST":
		case strings.HasPrefix(line, "#EXT-X-KEY:"):
			key = line
		case line == "#EXT-X-DISCONTINUITY":
			group++
		case strings.HasPrefix(line, "#EXTINF:"):
			extinf = line
		case strings.HasPrefix(line, "#"):
			if len(segs) == 0 {
				head.WriteString(line + "\n")
			}
		default:
			ref := segmentRef{url: line, extinf: extinf, group: group}
			if extinf == "" {
				return "", nil, errors.New("segment without a duration")
			}
			extinf = ""
			if key != "" && !strings.Contains(key, "METHOD=NONE") {
				m := keyURIRe.FindStringSubmatch(key)
				iv := keyIVRe.FindStringSubmatch(key)
				if m == nil || iv == nil {
					return "", nil, errors.New("key without a URI or an IV")
				}
				ref.keyURL = m[1]
				// keyIVRe admits 1 to 32 hex digits: padded to 32 they decode.
				raw, _ := hex.DecodeString(strings.Repeat("0", 32-len(iv[1])) + iv[1])
				copy(ref.iv[:], raw)
			}
			segs = append(segs, ref)
		}
	}
	if len(segs) == 0 {
		return "", nil, errors.New("no segments")
	}
	return head.String(), segs, nil
}

// extinfSeconds is the duration an #EXTINF line states.
func extinfSeconds(line string) float64 {
	v := strings.TrimSuffix(strings.SplitN(strings.TrimPrefix(line, "#EXTINF:"), ",", 2)[0], ",")
	d, _ := strconv.ParseFloat(strings.TrimSpace(v), 64)
	return d
}

// planRebase reads where each range of the playlist starts its clock and
// answers a plan when they do not run on from one another. It answers nil, nil
// when they do (nothing to do) and an error when it could not tell, in which
// case the caller plays the title as it is.
func (c *Client) planRebase(ctx context.Context, playlist string) (*rebasePlan, error) {
	header, segs, err := parseTitlePlaylist(playlist)
	if err != nil {
		return nil, err
	}
	for _, s := range segs {
		if !c.trusted(s.url) || (s.keyURL != "" && !c.trusted(s.keyURL)) {
			return nil, errors.New("segment on a host that is not Pluto TV's")
		}
	}
	plan := &rebasePlan{header: header, segs: segs, fetch: c.getBytes}

	// The first segment of each range, and where the playlist's durations say
	// the range begins.
	type start struct {
		first int
		at    float64
	}
	var starts []start
	elapsed := 0.0
	for i, s := range segs {
		if i == 0 || s.group != segs[i-1].group {
			starts = append(starts, start{first: i, at: elapsed})
		}
		elapsed += extinfSeconds(s.extinf)
	}
	if len(starts) < 2 {
		return nil, nil // one range has no other to disagree with
	}

	clocks := make([]int64, len(starts))
	errs := make([]error, len(starts))
	var wg sync.WaitGroup
	for i, st := range starts {
		wg.Add(1)
		go func() {
			defer wg.Done()
			clocks[i], errs[i] = plan.probe(ctx, segs[st.first])
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			return nil, err
		}
	}

	// Every range is moved to where the durations put it, relative to the
	// first, whose own clock is the reference.
	shifts := make([]int64, len(starts))
	needed := false
	for i, st := range starts {
		shifts[i] = int64(math.Round(st.at*90000)) + clocks[0] - clocks[i]
		if shifts[i] > rebaseThreshold || shifts[i] < -rebaseThreshold {
			needed = true
		}
	}
	if !needed {
		return nil, nil
	}
	for i, st := range starts {
		end := len(segs)
		if i+1 < len(starts) {
			end = starts[i+1].first
		}
		for k := st.first; k < end; k++ {
			plan.segs[k].shift = shifts[i]
		}
	}
	util.Debug("Pluto TV ranges restart their clocks; the title is put on one timeline", "ranges", len(starts))
	return plan, nil
}

// probe is the clock of a segment's first timestamp, read from its first bytes.
func (p *rebasePlan) probe(ctx context.Context, s segmentRef) (int64, error) {
	head, err := p.fetch(ctx, s.url, probeBytes, true)
	if err != nil {
		return 0, err
	}
	if s.keyURL != "" {
		key, err := p.fetch(ctx, s.keyURL, 64, false)
		if err != nil {
			return 0, err
		}
		if head, err = decryptPrefix(head, key, s.iv); err != nil {
			return 0, err
		}
	}
	at, ok := firstTimestamp(head)
	if !ok {
		return 0, errNotTS
	}
	return at, nil
}

// segment is the bytes the player is given for segment i: fetched, decrypted
// and moved.
func (p *rebasePlan) segment(ctx context.Context, i int, keys *keyCache) ([]byte, error) {
	if i < 0 || i >= len(p.segs) {
		return nil, errors.New("no such segment")
	}
	s := p.segs[i]
	data, err := p.fetch(ctx, s.url, maxSegmentBytes, false)
	if err != nil {
		return nil, err
	}
	if s.keyURL != "" {
		key, err := keys.get(ctx, s.keyURL, p.fetch)
		if err != nil {
			return nil, err
		}
		if data, err = decryptAES128(data, key, s.iv); err != nil {
			return nil, err
		}
	}
	if err := shiftTimestamps(data, s.shift); err != nil {
		return nil, err
	}
	return data, nil
}

// keyCache remembers the keys of a title: one per range, asked for by every
// segment of it.
type keyCache struct {
	mu   sync.Mutex
	keys map[string][]byte
}

func (k *keyCache) get(ctx context.Context, url string, fetch fetchFn) ([]byte, error) {
	k.mu.Lock()
	key, ok := k.keys[url]
	k.mu.Unlock()
	if ok {
		return key, nil
	}
	key, err := fetch(ctx, url, 64, false)
	if err != nil {
		return nil, err
	}
	if len(key) != 16 {
		return nil, errors.New("key is not 16 bytes")
	}
	k.mu.Lock()
	if k.keys == nil {
		k.keys = make(map[string][]byte)
	}
	k.keys[url] = key
	k.mu.Unlock()
	return key, nil
}

// getBytes is the fetchFn of a client.
func (c *Client) getBytes(ctx context.Context, target string, limit int64, prefix bool) ([]byte, error) {
	if !c.trusted(target) {
		return nil, errors.New("host is not Pluto TV's")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, http.NoBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "*/*")
	if prefix {
		req.Header.Set("Range", fmt.Sprintf("bytes=0-%d", limit-1))
	}
	resp, err := c.http.Do(req) // #nosec G704 -- hosts are checked by trusted()
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	if prefix {
		return io.ReadAll(io.LimitReader(resp.Body, limit))
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, errors.New("response exceeds size limit")
	}
	return body, nil
}
