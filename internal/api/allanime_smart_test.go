package api

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizeSmart(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"normal", "Naruto", "Naruto"},
		{"allanime tag", "Naruto [AllAnime]", "Naruto"},
		{"animefire tag", "Naruto [AnimeFire]", "Naruto"},
		{"slash", "a/b", "a_b"},
		{"backslash", "a\\b", "a_b"},
		{"colon", "a:b", "a_b"},
		{"wildcard chars", `a*b?c"d<e>f|g`, "a_b_c_d_e_f_g"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sanitizeSmart(tt.in))
		})
	}
}

func TestSanitizeSmartDest(t *testing.T) {
	t.Parallel()

	home, err := os.UserHomeDir()
	require.NoError(t, err)
	root := filepath.Join(home, ".local", "goanime", "downloads", "anime")

	t.Run("valid path under root", func(t *testing.T) {
		t.Parallel()
		good := filepath.Join(root, "Naruto", "1.mp4")
		got, err := sanitizeSmartDest(good)
		require.NoError(t, err)
		assert.Equal(t, good, got)
	})

	t.Run("empty rejected", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeSmartDest("")
		assert.Error(t, err)
	})

	t.Run("dash prefix rejected", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeSmartDest("-malicious")
		assert.Error(t, err)
	})

	t.Run("control bytes rejected", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeSmartDest("file\x00.mp4")
		assert.Error(t, err)
	})

	t.Run("path escape rejected", func(t *testing.T) {
		t.Parallel()
		_, err := sanitizeSmartDest("/tmp/evil.mp4")
		assert.Error(t, err)
	})
}

func TestValidateSmartRangeInputs(t *testing.T) {
	t.Parallel()

	t.Run("non-allanime rejected", func(t *testing.T) {
		t.Parallel()
		q := "best"
		err := validateSmartRangeInputs(&models.Anime{Source: "AnimeFire"}, 1, 2, &q)
		assert.Error(t, err)
	})

	t.Run("invalid range", func(t *testing.T) {
		t.Parallel()
		q := "best"
		err := validateSmartRangeInputs(&models.Anime{Source: "AllAnime"}, 0, 0, &q)
		assert.Error(t, err)
	})

	t.Run("end before start", func(t *testing.T) {
		t.Parallel()
		q := "best"
		err := validateSmartRangeInputs(&models.Anime{Source: "AllAnime"}, 5, 2, &q)
		assert.Error(t, err)
	})

	t.Run("quality defaults to best", func(t *testing.T) {
		t.Parallel()
		q := ""
		err := validateSmartRangeInputs(&models.Anime{Source: "AllAnime"}, 1, 3, &q)
		require.NoError(t, err)
		assert.Equal(t, "best", q)
	})

	t.Run("explicit quality preserved", func(t *testing.T) {
		t.Parallel()
		q := "720p"
		err := validateSmartRangeInputs(&models.Anime{Source: "AllAnime"}, 1, 1, &q)
		require.NoError(t, err)
		assert.Equal(t, "720p", q)
	})
}

func TestShouldUseYtDlp(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"m3u8", "https://x.com/master.m3u8", true},
		{"wixmp", "https://x.wixmp.com/video", true},
		{"repackager wixmp", "https://repackager.wixmp.com/v", true},
		{"blogger", "https://blogger.com/x", true},
		{"sharepoint", "https://sharepoint.com/x", true},
		{"allanime", "https://allanime.to/x", true},
		{"allmanga", "https://allmanga.x/y", true},
		{"plain mp4", "https://x.com/video.mp4", false},
		{"plain http", "https://example.com/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, shouldUseYtDlp(tt.url))
		})
	}
}

func TestIsUnsafeExtensionError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil", nil, false},
		{"unsafe extension", errors.New("file has unsafe extension"), true},
		{"unusual extension", errors.New("unusual extension detected"), true},
		{"unusual will be skipped", errors.New("file is unusual and will be skipped"), true},
		{"other error", errors.New("network failure"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isUnsafeExtensionError(tt.err))
		})
	}
}

func TestAlreadyDownloaded(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		assert.False(t, alreadyDownloaded(filepath.Join(dir, "nope")))
	})

	t.Run("small file ignored", func(t *testing.T) {
		t.Parallel()
		small := filepath.Join(dir, "small.mp4")
		require.NoError(t, os.WriteFile(small, []byte("tiny"), 0o600))
		assert.False(t, alreadyDownloaded(small))
	})

	t.Run("valid large file", func(t *testing.T) {
		t.Parallel()
		big := filepath.Join(dir, "big.mp4")
		require.NoError(t, os.WriteFile(big, make([]byte, 2048), 0o600))
		assert.True(t, alreadyDownloaded(big))
	})
}

func TestSmartOutputDir(t *testing.T) {
	t.Parallel()
	got, err := smartOutputDir(&models.Anime{Name: "Naruto Shippuden"})
	require.NoError(t, err)
	assert.Contains(t, got, filepath.Join(".local", "goanime", "downloads", "anime", "Naruto Shippuden"))
}

func TestWriteAniSkipSidecar(t *testing.T) {
	t.Parallel()

	t.Run("nil episode no-op", func(t *testing.T) {
		t.Parallel()
		assert.NoError(t, writeAniSkipSidecar("/tmp/x.mp4", nil))
	})

	t.Run("no skip windows no-op", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		video := filepath.Join(dir, "ep.mp4")
		ep := &models.Episode{Number: "1"}
		require.NoError(t, writeAniSkipSidecar(video, ep))
		_, err := os.Stat(filepath.Join(dir, "ep.skips.json"))
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("writes payload", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		video := filepath.Join(dir, "ep.mp4")
		ep := &models.Episode{
			Number:    "1",
			SkipTimes: models.SkipTimes{Op: models.Skip{Start: 0, End: 90}},
		}
		require.NoError(t, writeAniSkipSidecar(video, ep))
		sidecar := filepath.Join(dir, "ep.skips.json")
		data, err := os.ReadFile(sidecar)
		require.NoError(t, err)
		var payload map[string]any
		require.NoError(t, json.Unmarshal(data, &payload))
		assert.Equal(t, "aniskip", payload["format"])
		assert.EqualValues(t, 90, payload["op_end"])
		assert.Equal(t, "AllAnime", payload["source"])
	})
}

func TestWriteAniSkipSidecar_Exported(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	video := filepath.Join(dir, "ep.mp4")
	ep := &models.Episode{Number: "1", SkipTimes: models.SkipTimes{Op: models.Skip{End: 90}}}
	require.NoError(t, WriteAniSkipSidecar(video, ep))
	_, err := os.Stat(filepath.Join(dir, "ep.skips.json"))
	assert.NoError(t, err)
}

func TestResolveStreamURLForEpisode_NilArgs(t *testing.T) {
	t.Parallel()
	_, err := resolveStreamURLForEpisode(nil, nil, "best")
	assert.Error(t, err)

	_, err = resolveStreamURLForEpisode(&models.Episode{}, nil, "best")
	assert.Error(t, err)

	_, err = resolveStreamURLForEpisode(nil, &models.Anime{}, "best")
	assert.Error(t, err)
}
