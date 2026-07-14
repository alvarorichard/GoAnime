package superflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// restrictedShellHTML mirrors the page users get stuck on (captured from the
// filme/121390 screenshot): a static "Acesso Restrito / Visualização Externa"
// card whose only real content is a copy-paste embed iframe.
const restrictedShellHTML = `<!doctype html><html lang="pt-br"><head><title>Embed | Blades</title></head>
<body><div class="card"><span>ACESSO RESTRITO</span><h1>Visualização Externa</h1>
<p>Este conteúdo é protegido. Use o código ao lado para incorporar…</p>
<pre>&lt;iframe src="https://superflixapi.pro/filme/121390?embed_expires=1815362844&amp;embed_sig=deadbeef"&gt;</pre>
</div></body></html>`

func TestIsRestrictedEmbedPage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		html string
		want bool
	}{
		{"the real restricted shell", restrictedShellHTML, true},
		{"lowercase acesso restrito", `<body>acesso restrito não, mas "Acesso Restrito" sim</body>`, true},
		{"visualização externa alone", `<h1>Visualização Externa</h1>`, true},
		{"a real player page is not restricted", sfRealPlayerPage, false},
		{"empty", "", false},
		{"an ordinary page", `<html><body>Naruto</body></html>`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isRestrictedEmbedPage([]byte(tt.html)))
		})
	}
}
