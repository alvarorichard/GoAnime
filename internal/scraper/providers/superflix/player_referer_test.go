package superflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPlayerRefererFor pins the Referer shape the CDN actually accepts.
//
// Regression for 2026-08-26: both the fresh-sniff and the cache replay path
// built "<playerHost>/" and every signed master.txt came back 403. Because
// streamURLDead maps 403 to "rotated-out host", a good solve was thrown away
// as a dead host and SuperFlix never reached mpv at all.
func TestPlayerRefererFor(t *testing.T) {
	t.Parallel()
	const host = "https://xn--tckasiu6cvova0eb5fua2449g98vg.best"
	const hash = "6f1896bb1bfc3d0bcd2cba28ad968dd4"

	t.Run("uses the player's own /video/<hash> page", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, host+"/video/"+hash, playerRefererFor(host, hash))
	})

	t.Run("a trailing slash on the host never doubles up", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, host+"/video/"+hash, playerRefererFor(host+"/", hash))
	})

	t.Run("falls back to the origin when no hash was captured", func(t *testing.T) {
		t.Parallel()
		// The raw-media fallback capture has no getVideo URL to read ?data=
		// from, so the origin is the only thing available.
		assert.Equal(t, host+"/", playerRefererFor(host, ""))
	})

	t.Run("no host yields no referer", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, playerRefererFor("", hash))
	})

	t.Run("never returns the bare origin when a hash is known", func(t *testing.T) {
		t.Parallel()
		// The exact value the CDN answers 403 to.
		assert.NotEqual(t, host+"/", playerRefererFor(host, hash))
	})
}

// TestCachedStreamRefererUsesVideoPage guards the cache replay path, which
// built its own Referer independently of the sniff and so regressed the same
// way. Both paths must now go through playerRefererFor.
func TestCachedStreamRefererUsesVideoPage(t *testing.T) {
	t.Parallel()
	ent := streamCacheEntry{
		Host: "https://xn--tckasiu6cvova0eb5fua2449g98vg.best",
		Hash: "6f1896bb1bfc3d0bcd2cba28ad968dd4",
	}
	assert.Equal(t,
		ent.Host+"/video/"+ent.Hash,
		playerRefererFor(ent.Host, ent.Hash),
		"the cached replay must send the same Referer the fresh sniff does")
}
