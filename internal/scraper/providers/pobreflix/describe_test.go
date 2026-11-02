package pobreflix

import (
	"strings"
	"testing"
)

func TestDescribePageNeverQuotesTheBody(t *testing.T) {
	secret := "v1.TOPSECRETTICKET"
	body := []byte(`<html><head><title>Todo   Poderoso - RedeFlixApi</title></head><script>var t="` + secret + `"</script></html>`)
	got := describePage(body)
	if !strings.Contains(got, `title "Todo Poderoso - RedeFlixApi"`) || !strings.Contains(got, "bytes") {
		t.Fatalf("describePage = %q; want size and the collapsed title", got)
	}
	if strings.Contains(got, secret) {
		t.Fatalf("describePage leaked page content: %q", got)
	}
	if got := describePage([]byte("no markup")); !strings.Contains(got, `title "none"`) {
		t.Fatalf("describePage without a title = %q", got)
	}
}
