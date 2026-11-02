package diag

import (
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactRemovesSecretsAndKeepsTheRest(t *testing.T) {
	t.Parallel()
	cases := []struct{ name, in, want string }{
		{"signed media url",
			"https://cdn.example/m/film.mp4?X-Amz-Algorithm=AWS4&X-Amz-Signature=abcdef123&X-Amz-Credential=AKIA%2Fx&X-Amz-Expires=300",
			"https://cdn.example/m/film.mp4?X-Amz-Algorithm=AWS4&X-Amz-Signature=[redacted]&X-Amz-Credential=[redacted]&X-Amz-Expires=300"},
		{"turnstile proof and tokens", "GET /filme/310?tsv=1793605772350.proof&token=0.the-token", "GET /filme/310?tsv=[redacted]&token=[redacted]"},
		{"a search is not a secret", "https://site.example/api?keyword=naruto&q=one+piece&page=2", "https://site.example/api?keyword=naruto&q=one+piece&page=2"},
		{"bare key is secret, keyword is not", "?key=abc&keyword=abc", "?key=[redacted]&keyword=abc"},
		{"redeflix tickets", `{"launchTicket":"v1.Yy_YeO98Rc-U2YV4.b_V0BtGHM3omELoO_77nnQ.D0V5um0yW_BRj9AEG"}`, `{"launchTicket":"[redacted]"}`},
		{"a ticket loose in a message", "player answered v1.Yy_YeO98Rc-U2YV4.b_V0BtGHM3omELoO_77nnQ.D0V5um0yW", "player answered v1.[redacted]"},
		{"authorization header", "Authorization: Bearer abc.def.ghi-123456", "Authorization: [redacted]"},
		{"cookie header", "Cookie: cf_clearance=Cin6LHlq; session=xyz", "Cookie: [redacted]"},
		{"clearance on its own", "got cf_clearance=Cin6LHlqI4OP6sheBk_flKNn-9 for goyabu", "got cf_clearance=[redacted] for goyabu"},
		{"bearer in prose", "sent bearer abcdefghijklmnop to the api", "sent Bearer [redacted] to the api"},
		{"credentials in a proxy url", "proxy http://alice:s3cret@10.0.0.1:8080/", "proxy http://[redacted]@10.0.0.1:8080/"},
		{"jwt", "id eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N", "id [redacted]"},
		{"a secret in prose, not in a URL", "GoAnime crashed: index out of range with token=LEAKY and sig=abc", "GoAnime crashed: index out of range with token=[redacted] and sig=[redacted]"},
		{"a secret as a log field", `msg="resolving" ticket=v1.AAAAAAAA.BBBBBBBB.CCCCCCCC err=boom`, `msg="resolving" ticket=[redacted] err=boom`},
		{"json members", `{"launchTicket":"v1.x","token":"abc","title":"Todo Poderoso","password": "p"}`, `{"launchTicket":"[redacted]","token":"[redacted]","title":"Todo Poderoso","password": "[redacted]"}`},
		{"ordinary fields are not touched", "source=Pobreflix status=200 elapsed=412ms title=Todo", "source=Pobreflix status=200 elapsed=412ms title=Todo"},
		{"plain text untouched", "Pobreflix moved to a mirror domain from=https://a.gay to=https://a.christmas", "Pobreflix moved to a mirror domain from=https://a.gay to=https://a.christmas"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, Redact(tc.in))
		})
	}
}

func TestRedactIsIdempotent(t *testing.T) {
	t.Parallel()
	in := "https://x/?token=abc&sig=def Cookie: a=b v1.abcdef12.ghijkl34.mnop5678"
	once := Redact(in)
	assert.Equal(t, once, Redact(once))
	assert.NotContains(t, once, "abc&")
}

func TestRedactHome(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "~/.local/share/goanime/logs/x.log", RedactHome("/home/krone/.local/share/goanime/logs/x.log", "/home/krone"))
	assert.Equal(t, `~\AppData\Local\GoAnime`, RedactHome(`C:\Users\krone\AppData\Local\GoAnime`, `C:\Users\krone`))
	assert.Equal(t, `~\AppData`, RedactHome(`c:\users\KRONE\AppData`, `C:\Users\krone`), "Windows paths differ in case")
	assert.Equal(t, "/etc/hosts", RedactHome("/etc/hosts", "/"), "a root home would mangle every path")
	assert.Equal(t, "/usr/bin/mpv", RedactHome("/usr/bin/mpv", ""))
	assert.False(t, strings.Contains(RedactHome("/home/krone/a /home/krone/b", "/home/krone"), "krone"))
}

// Lowercasing can change a string's length in bytes ("İ" is two bytes, its
// lowercase "i̇" is three), so offsets taken from one cannot index the other.
// It panicked inside the log sink, and so inside the crash handler.
func TestRedactHomeDoesNotPanicWhenCaseFoldingChangesLength(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() {
		got := RedactHome("İİİ /home/Zed/a /home/zed/b", "/home/Zed")
		assert.Equal(t, "İİİ ~/a ~/b", got, "both spellings of the home folder go")
	})
	assert.NotPanics(t, func() { RedactHome("ǅǅǅ C:\\Users\\Zed\\x c:\\users\\zed\\y", `C:\Users\Zed`) })
}

// The first parameter of a URL starts right after "?", and the name must not
// swallow the path before it: "sig", "key" and "auth" are secrets only as whole
// names.
func TestRedactCatchesTheFirstQueryParameter(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"GET https://host/path?sig=abc123&x=1":     "GET https://host/path?sig=[redacted]&x=1",
		"https://host/p?key=SECRET":                "https://host/p?key=[redacted]",
		"https://host/p?auth=SECRET&q=naruto":      "https://host/p?auth=[redacted]&q=naruto",
		"https://host/p?sid=SECRET#frag":           "https://host/p?sid=[redacted]#frag",
		"https://host/p?q=naruto&key=SECRET":       "https://host/p?q=naruto&key=[redacted]",
		"https://host/p?keyword=naruto":            "https://host/p?keyword=naruto",
		"https://host/some.key/file?page=2":        "https://host/some.key/file?page=2",
		"url=https://host/p?pwd=SECRET and moving": "url=https://host/p?pwd=[redacted] and moving",
	} {
		assert.Equal(t, want, Redact(in), in)
	}
}

func BenchmarkRedactTypicalLogLine(b *testing.B) {
	line := `17:21:39 DEBUG <api/enhanced.go:340> GoAnime: Provider check source=Goyabu result=challenge ms=1218 status=403 detail="cloudflare challenge (challenge)" url=https://goyabu.io/anime/black-clover?page=2` + "\n"
	b.ReportAllocs()
	for b.Loop() {
		_ = Redact(line)
	}
}

// The shortcut that skips every regular expression for a line with no trigger
// word is only an optimisation: it must never change what Redact returns.
func TestFastPathAgreesWithEveryRule(t *testing.T) {
	t.Parallel()
	fragments := []string{
		"token=", "sig=", "key=", "keyword=", "q=", "?", "&", " ", "=", ":", "@", "/", "https://", "host", "x.example",
		"abc123", "SECRET", "v1.AAAAAAAA.BBBBBBBB.CCCCCCCC", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N",
		"Authorization: ", "Cookie: ", "Bearer ", "cf_clearance=", "X-Amz-Signature=", "password=", "user:pw@", `"token":"`, `"`, "naruto",
		"status=200", "source=Pobreflix", "17:21:36", "İ", "\n",
	}
	rng := rand.New(rand.NewPCG(7, 11))
	for range 20000 {
		var b strings.Builder
		for range 1 + rng.IntN(9) {
			b.WriteString(fragments[rng.IntN(len(fragments))])
		}
		in := b.String()
		want := RedactHome(redactEverything(in), homeDir())
		if got := Redact(in); got != want {
			t.Fatalf("the shortcut changed the answer for %q:\n  Redact    = %q\n  every rule = %q", in, got, want)
		}
	}
}

// Every word a rule can match on must be a trigger, or the shortcut would skip
// a line the rule would have cleaned.
func TestEveryRuleWordIsATrigger(t *testing.T) {
	t.Parallel()
	for _, w := range append(append([]string{}, sensitiveWords...), "key", "sig", "sid", "tsv", "auth", "pass", "pwd") {
		assert.True(t, mayHoldASecret("a?"+w+"=VALUE"), w)
	}
	for _, w := range []string{"Authorization: x", "Cookie: x", "X-Api-Key: x", "Bearer abcdefghij", "v1.abcdefgh.ijklmnop", "eyJabcdefgh.abcdefgh.abcd", "cf_clearance=x", "u:p@h"} {
		assert.True(t, mayHoldASecret(w), w)
	}
	assert.False(t, mayHoldASecret(`17:21:39 DEBUG GoAnime: Provider check source=Goyabu result=ok ms=12 status=200`))
}
