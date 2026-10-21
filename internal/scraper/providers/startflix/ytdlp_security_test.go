package startflix

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestYtdlpArgsRejectsNonHTTPPlayers(t *testing.T) {
	t.Parallel()
	for _, player := range []string{
		"--exec=touch injected", "file:///tmp/video.mp4", "/tmp/video.mp4",
		"https:///no-host", "https://%broken", "",
	} {
		t.Run(player, func(t *testing.T) {
			t.Parallel()
			require.Nil(t, ytdlpArgs(player, ""))
		})
	}
}

func TestYtdlpArgsTreatsScrapedValuesAsData(t *testing.T) {
	t.Parallel()
	player := "https://video.test/embed/1?x=--exec=command"
	referer := "--exec=command"
	args := ytdlpArgs(player, referer)
	require.NotNil(t, args)
	require.Contains(t, args, "--referer="+referer)
	require.NotContains(t, args, referer, "the referer must not become an option")
	require.Equal(t, []string{"--", player}, args[len(args)-2:])
}

func TestYtdlpArgsNormalizesProtocolRelativePlayers(t *testing.T) {
	t.Parallel()
	player := "//video.test/e/1?x=--exec=command"
	referer := "--exec=command"
	args := ytdlpArgs(player, referer)
	require.NotNil(t, args, "protocol-relative players must reach the fallback")
	require.Contains(t, args, "--referer="+referer)
	require.NotContains(t, args, referer, "the referer must not become an option")
	require.Equal(t, []string{"--", "https:" + player}, args[len(args)-2:])
}
