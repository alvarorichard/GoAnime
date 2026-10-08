package startflix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5" // #nosec G501 -- md5 is the player's key derivation, not a security choice of ours
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Abyss (abyss.to; on StartFlix's panel as playembedapi.site) is the host most
// titles are offered on. Its embed page carries the video's metadata sealed,
// and the video itself is an MP4 whose first 64 KiB arrive encrypted. Measured
// against the player bundles (iamcdn.net/player-v2 core + sw) on 2026-10-07:
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
//
// Renditions without url/path use the player's chunk token route. The local
// proxy maps byte ranges to those chunks and opens the encrypted .fd prefix,
// so mpv and the downloader still see one ordinary MP4.

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

// utf8ToLatin1 narrows a string of code points ≤ 0xFF back to bytes.
func utf8ToLatin1(s string) ([]byte, error) {
	out := make([]byte, 0, len(s))
	for _, r := range s {
		if r > 0xFF {
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
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	ctr := make([]byte, aes.BlockSize)
	copy(ctr, iv)
	addToCounter(ctr, uint64(off/aes.BlockSize)) // #nosec G115 -- off is never negative
	stream := cipher.NewCTR(block, ctr)
	if skip := int(off % aes.BlockSize); skip > 0 {
		var discard [aes.BlockSize]byte
		stream.XORKeyStream(discard[:skip], discard[:skip])
	}
	stream.XORKeyStream(buf, buf)
	return nil
}

// abyssChunkToken mirrors the player's token builder: AES-CTR encrypts the
// virtual chunk path using the ASCII MD5 hex of the file size as both key
// material and the source of the initial counter, then wraps it in Base64
// twice (without padding).
func abyssChunkToken(path string, size int64) string {
	ciphertext := []byte(path)
	if err := abyssXOR(strconv.FormatInt(size, 10), 0, ciphertext); err != nil {
		return ""
	}
	inner := strings.TrimRight(base64.StdEncoding.EncodeToString(ciphertext), "=")
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(inner)), "=")
}

// addToCounter adds n to a big-endian 128-bit counter, as CTR increments it.
func addToCounter(ctr []byte, n uint64) {
	for i := len(ctr) - 1; i >= 0 && n > 0; i-- {
		sum := uint64(ctr[i]) + (n & 0xFF)
		ctr[i] = byte(sum)
		n = (n >> 8) + (sum >> 8)
	}
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

// pickAbyssSource prefers the tallest H.264 rendition that is a single file
// (url + path); AV1 only when nothing else is offered.
func pickAbyssSource(sources []abyssSource) (abyssSource, bool) {
	var files []abyssSource
	for _, s := range sources {
		if s.URL != "" && s.Path != "" && s.Size > abyssEncryptedHead {
			files = append(files, s)
		}
	}
	if len(files) == 0 {
		return abyssSource{}, false
	}
	sort.SliceStable(files, func(i, j int) bool {
		ai, aj := strings.EqualFold(files[i].Codec, "av1"), strings.EqualFold(files[j].Codec, "av1")
		if ai != aj {
			return !ai
		}
		return files[i].ResID > files[j].ResID
	})
	return files[0], true
}

func abyssChunkDomain(domains []string, sub string) string {
	for _, domain := range domains {
		if sub != "" && strings.Contains(domain, sub) {
			return domain
		}
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
	chunk    *abyssChunkSource
}

type abyssChunkSource struct {
	md5ID     int64
	resID     int
	sub       string
	codec     string
	label     string
	chunkURL  string
	chunkHost string
	firstURL  string
	firstSize int64
	firstSeed string
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
	file, ok := makeAbyssFile(datas, media)
	if !ok {
		return nil, fmt.Errorf("abyss %s: no supported MP4 rendition", datas.Slug)
	}
	if err := c.probeAbyss(ctx, file); err != nil {
		return nil, err
	}
	local, err := sharedStreamProxy().serve(file, c.proxyClient())
	if err != nil {
		return nil, err
	}
	return &Stream{URL: local, Host: embed.Host}, nil
}

func makeAbyssFile(datas *abyssDatas, media *abyssMedia) (*abyssFile, bool) {
	if src, ok := pickAbyssSource(media.MP4.Sources); ok {
		upstream := strings.TrimRight(src.URL, "/") + "/" + strings.TrimLeft(src.Path, "/")
		return &abyssFile{
			upstream: upstream,
			size:     src.Size,
			seed:     upstream[strings.LastIndex(upstream, "/")+1:],
			origin:   "https://player.abyssplayer.com",
		}, true
	}

	var choices []abyssSource
	for _, src := range media.MP4.Sources {
		if src.Size <= abyssEncryptedHead || src.Sub == "" || src.ResID <= 0 {
			continue
		}
		if abyssChunkDomain(media.MP4.Domains, src.Sub) == "" {
			continue
		}
		if _, ok := pickAbyssFirstData(media.MP4.FristDatas, src); !ok {
			continue
		}
		choices = append(choices, src)
	}
	if len(choices) == 0 {
		return nil, false
	}
	sort.SliceStable(choices, func(i, j int) bool {
		ai, aj := strings.EqualFold(choices[i].Codec, "av1"), strings.EqualFold(choices[j].Codec, "av1")
		if ai != aj {
			return !ai
		}
		return choices[i].ResID > choices[j].ResID
	})
	src := choices[0]
	first, _ := pickAbyssFirstData(media.MP4.FristDatas, src)
	domain := abyssChunkDomain(media.MP4.Domains, src.Sub)
	firstURL, err := url.Parse(first.URL)
	if err != nil || firstURL.Scheme != "https" || firstURL.Host == "" {
		return nil, false
	}
	chunkURL := "https://" + domain + "/mp4/" + strconv.FormatInt(datas.MD5ID, 10) + "/" + strconv.Itoa(src.ResID) + "/" + strconv.FormatInt(src.Size, 10)
	return &abyssFile{
		size:   src.Size,
		origin: "https://player.abyssplayer.com",
		chunk: &abyssChunkSource{
			md5ID: datas.MD5ID, resID: src.ResID, sub: src.Sub,
			codec: src.Codec, label: src.Label,
			chunkURL: chunkURL, chunkHost: domain,
			firstURL: first.URL, firstSize: first.PartSize,
			firstSeed: firstURL.Path[strings.LastIndex(firstURL.Path, "/")+1:],
		},
	}, true
}

// probeAbyss checks the CDN serves the file before mpv is pointed at it, so a
// dead rendition fails over to the next player instead of a blank window.
func (c *Client) probeAbyss(ctx context.Context, f *abyssFile) error {
	if f.chunk != nil {
		buf, err := readAbyssRange(ctx, c.proxyClient(), f, 0, 15)
		if err != nil {
			return err
		}
		if len(buf) < 8 || string(buf[4:8]) != "ftyp" {
			return fmt.Errorf("abyss: decrypted chunk header is not an MP4")
		}
		return nil
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
