package startflix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec G501 -- md5 is the player's key derivation, not a security choice of ours
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"math/bits"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Abyss (abyss.to; on StartFlix's panel as playembedapi.site) is the host most
// titles are offered on. Its embed page carries the video's metadata sealed,
// and the video itself is served in one of two layouts. Measured against the
// player bundles (iamcdn.net/player-v2 core + sw) on 2026-10-07 and the live
// player on 2026-10-08:
//
//   - The page sets `const datas = "<base64 JSON>"` with slug, md5_id, user_id
//     and "media", a Latin-1 string of AES-256-CTR ciphertext. The key is the
//     UTF-8 bytes of md5hex("<user_id>:<slug>:<md5_id>"); the counter block is
//     the first 16 of those bytes.
//   - The plaintext lists renditions. One that carries both "url" and "path"
//     is a single MP4 at url + "/" + path, served with Range — but only to a
//     request bearing the embed's Origin and Referer (403 without them).
//   - Bytes [0, 65536) of that file are AES-256-CTR encrypted with the key
//     md5hex(<last path segment>), counter as above; the rest is plain. The
//     player checks the decrypted head for "ftyp" before using it.
//   - Renditions without url/path ("sub" plus a "domains" list instead) are
//     served in chunks: the first partSize bytes come from a ".fd" URL with a
//     Range (same head-only encryption as above, keyed by its last path
//     segment), and every 2 MiB part after that comes from
//     https://<domain>/sora/<size>/<token>, where token is the double base64
//     (padding stripped) of the AES-256-CTR encryption of the part path
//     "/mp4/<md5_id>/<res_id>/<size>/<chunk>/<part>". The token key is
//     md5hex(<size>) — but computed by the player's own md5 over the size as
//     a JS number: it stringifies to decimal and then hashes the digit VALUES
//     ([3,9,2,...], not the ASCII bytes [51,57,...]), because its bytesToWords
//     coerces each character with `<<`. The counter block is the first 16 key
//     bytes, as usual. Verified byte-for-byte against tokens the live player
//     minted on 2026-10-08.
//
// The local proxy maps byte ranges to .fd ranges and chunk tokens, so mpv and
// the downloader still see one ordinary MP4.

const abyssEncryptedHead = 65536

var abyssDatasRe = regexp.MustCompile(`datas\s*=\s*"([A-Za-z0-9+/=]+)"`)

// isAbyssHost reports whether a player host serves the Abyss player.
func isAbyssHost(host string) bool {
	host = strings.ToLower(host)
	return strings.Contains(host, "playembedapi") || strings.Contains(host, "abyss") || strings.Contains(host, "hydrax")
}

type abyssDatas struct {
	Slug   string `json:"slug"`
	MD5ID  int64  `json:"md5_id"`
	UserID int64  `json:"user_id"`
	Media  string `json:"media"`
}

type abyssSource struct {
	Label  string `json:"label"`
	ResID  int    `json:"res_id"`
	Size   int64  `json:"size"`
	Codec  string `json:"codec"`
	Path   string `json:"path"`
	URL    string `json:"url"`
	Status bool   `json:"status"`
	Sub    string `json:"sub"`
}

type abyssFirstData struct {
	ResID    int    `json:"res_id"`
	Size     int64  `json:"size"`
	Codec    string `json:"codec"`
	URL      string `json:"url"`
	PartSize int64  `json:"partSize"`
}

type abyssMedia struct {
	MP4 struct {
		Sources    []abyssSource    `json:"sources"`
		Domains    []string         `json:"domains"`
		FristDatas []abyssFirstData `json:"fristDatas"`
	} `json:"mp4"`
}

// latin1ToUTF8 widens each byte to the code point of the same value. The
// decoded datas JSON carries raw bytes ≥0x80 inside a string; read as UTF-8
// they would be replaced, and the ciphertext with them.
func latin1ToUTF8(b []byte) []byte {
	out := make([]byte, 0, len(b)*2)
	for _, c := range b {
		out = utf8.AppendRune(out, rune(c))
	}
	return out
}

// utf8ToLatin1 narrows a string of code points ≤ 0xFF back to bytes. Ranging
// over a string never yields a negative rune (invalid UTF-8 comes out as
// U+FFFD), but the lower bound is spelled out so the narrowing is checked on
// both sides rather than assumed.
func utf8ToLatin1(s string) ([]byte, error) {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r < 0 || r > 0xFF {
			return nil, fmt.Errorf("code point %U is not a byte", r)
		}
		out = append(out, byte(r))
	}
	return out, nil
}

// parseAbyssDatas reads the sealed metadata from an embed page.
func parseAbyssDatas(page []byte) (*abyssDatas, error) {
	m := abyssDatasRe.FindSubmatch(page)
	if m == nil {
		return nil, errors.New("no datas on the embed page")
	}
	raw, err := base64.StdEncoding.DecodeString(string(m[1]))
	if err != nil {
		return nil, fmt.Errorf("datas: %w", err)
	}
	var d abyssDatas
	if err := jsonx.Unmarshal(latin1ToUTF8(raw), &d); err != nil {
		return nil, fmt.Errorf("datas: %w", err)
	}
	if d.Slug == "" || d.Media == "" {
		return nil, errors.New("datas without slug or media")
	}
	return &d, nil
}

// abyssKey derives the player's AES-256-CTR key and counter block from a seed.
func abyssKey(seed string) (key, iv []byte) {
	sum := md5.Sum([]byte(seed)) // #nosec G401 -- see import
	key = []byte(hex.EncodeToString(sum[:]))
	return key, key[:aes.BlockSize]
}

// abyssXOR applies the AES-CTR keystream for seed to buf, which starts at
// byte offset off of the stream. CTR is seekable: the counter for offset off
// is the initial block plus off/16, and off%16 keystream bytes are skipped.
func abyssXOR(seed string, off int64, buf []byte) error {
	key, iv := abyssKey(seed)
	return abyssCTR(key, iv, off, buf)
}

// abyssCTR applies an AES-CTR keystream to buf, which starts at byte offset
// off of the stream.
func abyssCTR(key, iv []byte, off int64, buf []byte) error {
	if off < 0 {
		return fmt.Errorf("negative CTR offset %d", off)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	var ctr [aes.BlockSize]byte
	copy(ctr[:], iv)
	addToCounter(&ctr, uint64(off/aes.BlockSize))
	stream := cipher.NewCTR(block, ctr[:])
	if skip := int(off % aes.BlockSize); skip > 0 {
		var discard [aes.BlockSize]byte
		stream.XORKeyStream(discard[:skip], discard[:skip])
	}
	stream.XORKeyStream(buf, buf)
	return nil
}

// abyssChunkToken mirrors the player's token builder: AES-CTR encrypts the
// virtual chunk path, then wraps it in Base64 twice (without padding). The key
// is the player's own md5 of the file size — computed over the size as a JS
// number, which stringifies to decimal and then gets hashed as digit VALUES
// (see abyssChunkKey), with the first 16 key bytes as the counter block.
func abyssChunkToken(path string, size int64) string {
	key, iv := abyssChunkKey(size)
	ciphertext := []byte(path)
	if err := abyssCTR(key, iv, 0, ciphertext); err != nil {
		return ""
	}
	inner := strings.TrimRight(base64.StdEncoding.EncodeToString(ciphertext), "=")
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(inner)), "=")
}

// abyssChunkKey derives the chunk token key for a file size. The player's md5
// receives the size as a JS number, stringifies it to decimal, and then feeds
// the string to a bytesToWords that coerces each character with `<<` — so the
// hashed bytes are the digit values ([3,9,2,...]), not the ASCII codes
// ([51,57,...]). Only non-negative integers reach this path (file sizes), so
// mapping each decimal digit to its value replicates the player exactly.
func abyssChunkKey(size int64) (key, iv []byte) {
	decimal := strconv.FormatInt(size, 10)
	digits := make([]byte, 0, len(decimal))
	for i := 0; i < len(decimal); i++ {
		if c := decimal[i]; c >= '0' && c <= '9' {
			digits = append(digits, c-'0')
		}
	}
	sum := md5.Sum(digits) // #nosec G401 -- see import
	key = []byte(hex.EncodeToString(sum[:]))
	return key, key[:aes.BlockSize]
}

// addToCounter adds n to a big-endian 128-bit counter block, as CTR
// increments it: n goes into the low 64 bits, the carry into the high 64, and
// the whole block wraps at 2^128 exactly like cipher.NewCTR's own counter.
func addToCounter(ctr *[aes.BlockSize]byte, n uint64) {
	lo, carry := bits.Add64(binary.BigEndian.Uint64(ctr[8:]), n, 0)
	binary.BigEndian.PutUint64(ctr[8:], lo)
	binary.BigEndian.PutUint64(ctr[:8], binary.BigEndian.Uint64(ctr[:8])+carry)
}

// decryptAbyssMedia opens the rendition list.
func decryptAbyssMedia(d *abyssDatas) (*abyssMedia, error) {
	ct, err := utf8ToLatin1(d.Media)
	if err != nil {
		return nil, err
	}
	seed := strconv.FormatInt(d.UserID, 10) + ":" + d.Slug + ":" + strconv.FormatInt(d.MD5ID, 10)
	if err := abyssXOR(seed, 0, ct); err != nil {
		return nil, err
	}
	var media abyssMedia
	if err := jsonx.Unmarshal(ct, &media); err != nil {
		return nil, fmt.Errorf("media: %w", err)
	}
	return &media, nil
}

// abyssBetter orders renditions best first: the tallest picture (res_id grows
// with resolution — 2 is 360p, 4 is 720p, 5 is 1080p), and at the same
// height H.264 before AV1, which more machines decode in hardware.
func abyssBetter(a, b abyssSource) bool {
	if a.ResID != b.ResID {
		return a.ResID > b.ResID
	}
	return !strings.EqualFold(a.Codec, "av1") && strings.EqualFold(b.Codec, "av1")
}

// pickAbyssSource returns the best rendition served as a single file (url +
// path on a plain https host name).
func pickAbyssSource(sources []abyssSource) (abyssSource, bool) {
	var files []abyssSource
	for _, s := range sources {
		if s.URL != "" && s.Path != "" && s.Size > abyssEncryptedHead && isAbyssMediaURL(abyssSingleFileURL(s)) {
			files = append(files, s)
		}
	}
	if len(files) == 0 {
		return abyssSource{}, false
	}
	sort.SliceStable(files, func(i, j int) bool { return abyssBetter(files[i], files[j]) })
	return files[0], true
}

func abyssSingleFileURL(s abyssSource) string {
	return strings.TrimRight(s.URL, "/") + "/" + strings.TrimLeft(s.Path, "/")
}

// isAbyssHostname reports whether s is a bare DNS host name, the only thing
// the media list may put between "https://" and a path. A port, user info, a
// path or query, or an IP literal would let that remote list steer requests
// somewhere a CDN host name cannot.
func isAbyssHostname(s string) bool {
	if s == "" || len(s) > 253 || net.ParseIP(s) != nil {
		return false
	}
	for label := range strings.SplitSeq(s, ".") {
		if label == "" || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' {
				return false
			}
		}
	}
	return true
}

// isAbyssMediaURL reports whether raw is an https URL on a bare host name,
// as every media URL the player loads is.
func isAbyssMediaURL(raw string) bool {
	u, err := url.Parse(raw)
	return err == nil && u.Scheme == "https" && u.User == nil && u.Opaque == "" && isAbyssHostname(u.Host)
}

func abyssChunkDomain(domains []string, size int64, sub string) string {
	for _, domain := range domains {
		if sub != "" && strings.Contains(domain, sub) {
			return domain
		}
	}
	// Without a sub match the player falls back to a size-picked domain.
	if len(domains) > 0 && size > 0 {
		return domains[int(uint64(size)%uint64(len(domains)))] // #nosec G115 -- size is positive here
	}
	return ""
}

func pickAbyssFirstData(data []abyssFirstData, src abyssSource) (abyssFirstData, bool) {
	for _, first := range data {
		if first.Size == src.Size && first.ResID == src.ResID && strings.EqualFold(first.Codec, src.Codec) && first.URL != "" && first.PartSize > 0 {
			return first, true
		}
	}
	return abyssFirstData{}, false
}

// abyssFile is one resolved Abyss MP4, as the local proxy serves it.
type abyssFile struct {
	upstream string // url + "/" + path
	size     int64
	seed     string // last path segment: the head's key seed
	origin   string // embed origin, sent as Origin and Referer
	height   int    // the rendition's picture height, 0 when its label gives none
	chunk    *abyssChunkSource

	partsOnce sync.Once
	parts     *abyssPartCache // this file's side of the shared part cache

	urlsOnce sync.Once
	urls     []string // upstream URL of every part; see abyssPartURLs
	urlsErr  error
}

// partsCache lazily creates the chunk part cache.
func (f *abyssFile) partsCache() *abyssPartCache {
	f.partsOnce.Do(func() { f.parts = newAbyssPartCache() })
	return f.parts
}

type abyssChunkSource struct {
	md5ID     int64
	resID     int
	sub       string
	codec     string
	label     string
	chunkURL  string // full base URL (kept for debugging)
	chunkPath string // URL path the tokens encrypt: /mp4/<md5>/<res>/<size>
	chunkHost string
	domains   []string // every host the media list offered, for sibling retries
	firstURL  string
	firstSize int64
	firstSeed string
}

// orderedChunkHosts returns the hosts to try for /sora/ parts: the resolved
// primary first, then every other host the media list offered (deduplicated).
// The token encrypts only the virtual path, so it is valid on each host that
// serves the same file, letting a part that one host refuses (5xx, transient
// reset) be retried elsewhere. It does not rescue a permanently short upload:
// live, every /sora/ host stopped the bad 876 MB rendition at the same byte.
// That case is handled by selectAbyssFile, not here. Hosts are filtered at
// file build time (makeAbyssChunkFile keeps only bare host names), so no
// validation happens here — tests point files at httptest servers directly.
func (c *abyssChunkSource) orderedChunkHosts() []string {
	seen := map[string]bool{}
	var out []string
	add := func(h string) {
		h = strings.TrimSpace(h)
		if h == "" || seen[h] {
			return
		}
		seen[h] = true
		out = append(out, h)
	}
	add(c.chunkHost)
	for _, h := range c.domains {
		add(h)
	}
	return out
}

// resolveAbyss turns an Abyss embed into a local, plain-MP4 URL.
func (c *Client) resolveAbyss(ctx context.Context, embed *url.URL, p Player) (*Stream, error) {
	page, err := c.get(ctx, embed.String(), getOptions{referer: p.referer, layer: "abyss"})
	if err != nil {
		return nil, err
	}
	datas, err := parseAbyssDatas(page)
	if err != nil {
		return nil, netx.NewParserError(SourceName, "abyss", "unreadable embed page", err)
	}
	media, err := decryptAbyssMedia(datas)
	if err != nil {
		return nil, netx.NewDecryptError(SourceName, "abyss", "could not open the media list", err)
	}
	files := makeAbyssFiles(embed, datas, media)
	if len(files) == 0 {
		return nil, fmt.Errorf("abyss %s: no supported MP4 rendition", datas.Slug)
	}
	file, err := c.selectAbyssFile(ctx, files)
	if err != nil {
		return nil, err
	}
	local, err := sharedStreamProxy().serve(file, c.proxyClient())
	if err != nil {
		return nil, err
	}
	return &Stream{URL: local, Host: embed.Host, Height: file.height}, nil
}

// makeAbyssFile returns the best Abyss rendition as one file: the first, and
// until a tail probe says otherwise the only, candidate of makeAbyssFiles.
func makeAbyssFile(embed *url.URL, datas *abyssDatas, media *abyssMedia) (*abyssFile, bool) {
	files := makeAbyssFiles(embed, datas, media)
	if len(files) == 0 {
		return nil, false
	}
	return files[0], true
}

// makeAbyssFiles returns every usable rendition as a file, best first, in the
// order the single-file-or-chunked quality rule always used: tallest picture,
// H.264 over AV1 at a tie, a single file over a chunked rendition at a tie.
// resolveAbyss probes them in order rather than prompting the first, because
// the panel lists renditions the CDN cannot fully serve — live: a 1080p H.264
// whose final 64 KiB never comes back, stalling a download at 99% while a
// sibling 1080p of the same episode serves whole.
func makeAbyssFiles(embed *url.URL, datas *abyssDatas, media *abyssMedia) []*abyssFile {
	// The CDN only serves media to requests carrying the embed's Origin and
	// Referer, so every file carries the embed origin like the player does.
	origin := embed.Scheme + "://" + embed.Host
	single, hasSingle := pickAbyssSource(media.MP4.Sources)

	var choices []abyssSource
	for _, src := range media.MP4.Sources {
		if src.Size <= abyssEncryptedHead || src.Sub == "" || src.ResID <= 0 {
			continue
		}
		if !isAbyssHostname(abyssChunkDomain(media.MP4.Domains, src.Size, src.Sub)) {
			continue
		}
		choices = append(choices, src)
	}
	sort.SliceStable(choices, func(i, j int) bool { return abyssBetter(choices[i], choices[j]) })

	chunked := make([]*abyssFile, 0, len(choices))
	for _, src := range choices {
		chunked = append(chunked, makeAbyssChunkFile(media, datas, src, origin))
	}

	// The best picture wins whichever way it is served; at a tie the single
	// file does, being one plain request per range. The loser still becomes a
	// candidate, so a first choice the CDN serves badly yields to the rest.
	out := chunked
	if hasSingle {
		file := makeAbyssSingleFile(single, origin)
		if len(choices) == 0 || !abyssBetter(choices[0], single) {
			out = append([]*abyssFile{file}, out...)
		} else {
			out = append(out, file)
		}
	}
	return out
}

func makeAbyssSingleFile(single abyssSource, origin string) *abyssFile {
	upstream := abyssSingleFileURL(single)
	return &abyssFile{
		upstream: upstream,
		size:     single.Size,
		seed:     upstream[strings.LastIndex(upstream, "/")+1:],
		origin:   origin,
		height:   labelHeight(single.Label),
	}
}

func makeAbyssChunkFile(media *abyssMedia, datas *abyssDatas, src abyssSource, origin string) *abyssFile {
	// The encrypted head is optional: without a fristData entry — or when its
	// partSize is not a whole number of parts, which the player refuses to
	// split — every part goes through the chunk tokens instead.
	var firstURL, firstSeed string
	var firstSize int64
	if first, ok := pickAbyssFirstData(media.MP4.FristDatas, src); ok {
		part := min(src.Size, int64(2<<20))
		if u, err := url.Parse(first.URL); err == nil && isAbyssMediaURL(first.URL) &&
			part > 0 && first.PartSize%part == 0 {
			firstURL, firstSize = first.URL, first.PartSize
			firstSeed = u.Path[strings.LastIndex(u.Path, "/")+1:]
		}
	}
	domain := abyssChunkDomain(media.MP4.Domains, src.Size, src.Sub)
	chunkPath := "/mp4/" + strconv.FormatInt(datas.MD5ID, 10) + "/" + strconv.Itoa(src.ResID) + "/" + strconv.FormatInt(src.Size, 10)
	chunkURL := "https://" + domain + chunkPath
	// Keep every offered host so a flaky tail chunk can fail over to a
	// sibling domain (see orderedChunkHosts). Only bare host names ever
	// reach the network; anything else is dropped here.
	var domains []string
	for _, d := range media.MP4.Domains {
		if isAbyssHostname(d) {
			domains = append(domains, d)
		}
	}
	return &abyssFile{
		size:   src.Size,
		origin: origin,
		height: labelHeight(src.Label),
		chunk: &abyssChunkSource{
			md5ID: datas.MD5ID, resID: src.ResID, sub: src.Sub,
			codec: src.Codec, label: src.Label,
			chunkURL: chunkURL, chunkPath: chunkPath, chunkHost: domain,
			domains:  domains,
			firstURL: firstURL, firstSize: firstSize,
			firstSeed: firstSeed,
		},
	}
}

// probeAbyss checks the CDN serves the file before mpv is pointed at it, so a
// dead rendition fails over to the next player instead of a blank window. For
// chunked files it opens both the encrypted head and the first token part, so
// a broken token scheme is caught here rather than minutes into playback. The
// two are fetched at once: each is a whole 2 MiB part, and one after the
// other they doubled how long a stream took to resolve — which a batch pays
// once per episode before its first download starts.
func (c *Client) probeAbyss(ctx context.Context, f *abyssFile) error {
	if f.chunk != nil {
		var head []byte
		var headErr, partErr error
		var wg sync.WaitGroup
		wg.Go(func() { head, headErr = readAbyssRange(ctx, c.proxyClient(), f, 0, 15) })
		if next := min(f.chunk.firstSize, f.size-1); next < f.size-1 {
			// The last part may be shorter than the 16 bytes read here.
			wg.Go(func() { _, partErr = readAbyssRange(ctx, c.proxyClient(), f, next, min(next+15, f.size-1)) })
		}
		wg.Wait()
		if headErr != nil {
			return headErr
		}
		if len(head) < 8 || string(head[4:8]) != "ftyp" {
			return fmt.Errorf("abyss: decrypted chunk header is not an MP4")
		}
		return partErr
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, f.upstream, http.NoBody)
	if err != nil {
		return err
	}
	f.decorate(req, c.userAgent)
	req.Header.Set("Range", "bytes=0-15")
	resp, err := c.proxyClient().Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return netx.NewHTTPStatusError(SourceName, "abyss", resp.StatusCode)
	}
	return nil
}

// decorate adds the headers the CDN requires: the embed's Origin and Referer.
func (f *abyssFile) decorate(req *http.Request, ua string) {
	req.Header.Set("User-Agent", ua)
	req.Header.Set("Origin", f.origin)
	req.Header.Set("Referer", f.origin+"/")
	req.Header.Set("Accept", "*/*")
}

// abyssTailProbeTimeout bounds the final-chunk check. A rendition the CDN
// cannot fully serve accepts the request, sends what it has and only then lets
// the connection reset — live: ~6.5s before "unexpected EOF" — so the check
// needs enough rope to observe the reset rather than hang forever.
const abyssTailProbeTimeout = 12 * time.Second

// selectAbyssFile probes candidates best first and returns the first a reader
// can actually stream: its head and first part come back, and for a chunked
// rendition so does its final chunk. A rendition whose tail the CDN never
// serves is skipped instead of being handed to mpv or the downloader, where it
// would stall at 99% retrying the same part forever.
func (c *Client) selectAbyssFile(ctx context.Context, files []*abyssFile) (*abyssFile, error) {
	var lastErr error
	for _, f := range files {
		if err := c.probeAbyss(ctx, f); err != nil {
			lastErr = err
			continue
		}
		if f.chunk != nil {
			if err := c.probeAbyssTail(ctx, f); err != nil {
				util.Debug("StartFlix rendition tail unserved", "height", f.height, "size", f.size, "err", err)
				lastErr = err
				continue
			}
		}
		return f, nil
	}
	if lastErr == nil {
		lastErr = errors.New("abyss: no usably served rendition")
	}
	return nil, lastErr
}

// probeAbyssTail fetches the last part's closing window and requires it whole.
// That window is where a truncated upload shows: the origin answers with the
// final part's declared length but stops one 64 KiB block short, so a request
// that must return every byte it asked for fails here in seconds instead of
// mid-download. Nothing is cached; a healthy part is fetched again when read.
func (c *Client) probeAbyssTail(ctx context.Context, f *abyssFile) error {
	part := f.abyssPartSize()
	if part <= 0 {
		return nil
	}
	last := (f.size - 1) / part
	lo, hi := f.abyssPartBounds(last)
	if hi < lo {
		return nil
	}
	want := hi - lo + 1
	win := min(int64(abyssEncryptedHead), want)
	urls, err := f.abyssPartURLs()
	if err != nil {
		return err
	}
	if last >= int64(len(urls)) {
		return fmt.Errorf("part %d out of range", last)
	}
	var from, through int64
	if f.chunk.firstSize > 0 && lo < f.chunk.firstSize {
		// .fd parts take absolute file offsets; token parts are part-relative.
		from, through = hi-win+1, hi
	} else {
		from, through = want-win, want-1
	}
	pctx, cancel := context.WithTimeout(ctx, abyssTailProbeTimeout)
	defer cancel()
	if _, err := abyssPartRequest(pctx, c.proxyClient(), urls[last], from, through, win, f.origin); err != nil {
		return fmt.Errorf("abyss: final chunk of the %d-byte rendition: %w", f.size, err)
	}
	return nil
}
