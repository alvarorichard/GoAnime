package superflix

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reported as "o captcha não está resolvendo sozinho" (2026-09-21). It was not:
// the debug log shows the gate clearing normally and handing back a
// `?cfv=<JWT>` URL with scope "embed". What failed is the step after it.
//
//	15:15:14  solve starts, window stays minimized throughout
//	15:16:12  CF solve result … hasALL_EPISODES=false hasCSRF_TOKEN=false
//	15:16:12  ERROR SuperFlix didn't show an episode list for this title
//
// The post-gate "Visualização Externa / Acesso Restrito" page carries the real
// player behind an embed URL, printed in its "Embed Code" box. SuperFlix changed
// that snippet to a BARE url:
//
//	<iframe src="https://superflixapi.quest/serie/1405" allow="autoplay *; …">
//
// Both extraction patterns required ?cfv=, so extractSuperFlixEmbedURL returned
// "" and readEmbeddedPlayer — the only thing that can reach the player from
// there — never ran. Every SuperFlix title failed, and the ~58s the gate spent
// beforehand made it look like the captcha was stuck.
//
// The token was never what made the read work: loading the URL in a genuine
// cross-origin iframe is, because that makes the request Sec-Fetch-Site:
// cross-site and the server mints a fresh cfv itself.
//
// The fixture is the real page, captured from a live solve.

func restrictedPageFixture(t *testing.T) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join("testdata", "restricted_embed_2026_09_21.html"))
	require.NoError(t, err, "fixture missing; it is the captured restricted page")
	return string(body)
}

func TestExtractEmbedURL_ReadsTheBareEmbedCode(t *testing.T) {
	t.Parallel()
	page := restrictedPageFixture(t)

	// The page must still be recognised as the restricted shell, or the solver
	// would not even look for an embed URL in it.
	require.True(t, isRestrictedEmbedPage([]byte(page)))
	require.False(t, isRealPlayerHTML(page), "this page has no player markers; that is the point")

	got := extractSuperFlixEmbedURL(page)

	assert.Equal(t, "https://superflixapi.quest/serie/1405", got,
		"without this the embed read never runs and the title reports no episodes")
}

// The cfv-bearing forms are still preferred: they identify the embed box
// unambiguously, where a bare URL is a best-effort last resort.
func TestExtractEmbedURL_PrefersACfvTokenWhenOneIsPresent(t *testing.T) {
	t.Parallel()
	const withToken = `<iframe src="https://superflixapi.quest/serie/1405?cfv=abc.def.ghi"></iframe>` +
		`<iframe src="https://superflixapi.quest/serie/9999"></iframe>`

	assert.Equal(t, "https://superflixapi.quest/serie/1405?cfv=abc.def.ghi",
		extractSuperFlixEmbedURL(withToken))
}

// The snippet is rendered as escaped text inside the box, and appears
// JSON-escaped elsewhere on the same page. Both forms have to decode.
func TestExtractEmbedURL_HandlesEscapedMarkup(t *testing.T) {
	t.Parallel()
	for name, page := range map[string]string{
		"html entities": `<div>&lt;iframe src="https://superflixapi.quest/filme/550" allow="autoplay *"&gt;&lt;/iframe&gt;</div>`,
		"json escaped":  `{"embed":"<iframe src=\"https:\/\/superflixapi.quest\/filme\/550\" allow=\"autoplay *\"><\/iframe>"}`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, "https://superflixapi.quest/filme/550", extractSuperFlixEmbedURL(page))
		})
	}
}

// The bare-URL pattern must not reach outside the SuperFlix family. The
// restricted page also links a third-party "watch via partner" site, and
// loading THAT in the player iframe would send the user somewhere else.
func TestExtractEmbedURL_RefusesOffFamilyHosts(t *testing.T) {
	t.Parallel()
	const page = `<a href="https://vizer.forum/serie/dexter/1/1">partner</a>` +
		`<iframe src="https://evil.example.com/serie/1405"></iframe>`

	assert.Empty(t, extractSuperFlixEmbedURL(page),
		"only a superflixapi.* host may be loaded as the embed")
}

func TestExtractEmbedURL_EmptyWhenThereIsNothingToFind(t *testing.T) {
	t.Parallel()
	assert.Empty(t, extractSuperFlixEmbedURL(""))
	assert.Empty(t, extractSuperFlixEmbedURL(`<html><body>nothing here</body></html>`))
}
