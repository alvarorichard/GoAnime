package superflix

import (
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
)

// TestSuperFlixGetVideoSniff_Live is a live recon harness: it drives the
// SuperFlix embed through Turnstile and logs the first getVideo request (and any
// server-chooser UI it finds). Diagnostic only — no assertions. Requires a real
// browser + network.
func TestSuperFlixGetVideoSniff_Live(t *testing.T) {
	skipInCI(t) // launches a real browser + hits the live site; recon-only, never in CI
	if testing.Short() {
		t.Skip("skipping live browser recon in -short")
	}
	s := &cfBrowserSolver{}
	bctx, err := s.init()
	if err != nil {
		t.Fatalf("init: %v", err)
	}
	t.Cleanup(s.Close)
	page, _ := bctx.NewPage()
	moveWindow(page, 60, 60)
	page.OnPopup(func(p playwright.Page) { go func() { _ = p.Close() }() })

	gv := regexp.MustCompile(`(?i)/player/index\.php\?.*do=getVideo`)
	var mu sync.Mutex
	var hit string
	page.OnResponse(func(r playwright.Response) {
		if gv.MatchString(r.URL()) {
			mu.Lock()
			if hit == "" {
				hit = r.URL()
			}
			mu.Unlock()
		}
	})

	page.Goto("https://"+SuperFlixEmbedHost+"/", playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded, Timeout: playwright.Float(40000)})
	page.Evaluate(`(src)=>{document.body.innerHTML='<iframe src="'+src+'" allow="autoplay; encrypted-media; fullscreen" style="position:fixed;inset:0;width:100%;height:100%;border:0"></iframe>'}`, "https://"+SuperFlixEmbedHost+"/serie/76479/5/8")

	// wait past turnstile, dump chooser, then auto-click
	dumped := false
	for i := 0; i < 30; i++ {
		mu.Lock()
		got := hit
		mu.Unlock()
		if got != "" {
			t.Logf(">>> getVideo FIRED after %ds: %s", i*2, got)
			return
		}
		if i == 6 && !dumped {
			dumped = true
			for _, fr := range page.Frames() {
				v, e := fr.Evaluate(`()=>{const o=[];document.querySelectorAll('button,a,div,li').forEach(el=>{const t=(el.textContent||'').trim();if(/servidor|dublado|legendado|principal|player/i.test(t)&&t.length<50)o.push(el.tagName+':"'+t+'" cls='+(''+el.className).slice(0,40))});return [...new Set(o)].slice(0,15)}`)
				if e == nil {
					if a, ok := v.([]interface{}); ok && len(a) > 0 {
						t.Logf("FRAME %s", fr.URL())
						for _, x := range a {
							t.Logf("   %v", x)
						}
					}
				}
			}
		}
		triggerPlay(page)
		time.Sleep(2 * time.Second)
	}
	t.Logf("NO getVideo in 60s")
}
