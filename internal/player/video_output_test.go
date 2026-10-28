package player

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVideoOutputArgLeadsWithGPUNext pins the --vo chain per OS: gpu-next
// first (the only GPU path a Mac has since mpv 0.39 is Vulkan on MoltenVK,
// and mpv itself defaults to it since 0.41), gpu right after it for older mpv,
// then each OS's last resorts.
func TestVideoOutputArgLeadsWithGPUNext(t *testing.T) {
	for goos, want := range map[string]string{
		"darwin":  "--vo=gpu-next,gpu,libmpv",
		"windows": "--vo=gpu-next,gpu,direct3d,sdl",
		"linux":   "--vo=gpu-next,gpu",
		"freebsd": "--vo=gpu-next,gpu",
	} {
		t.Run(goos, func(t *testing.T) {
			got := videoOutputArg(goos)
			assert.Equal(t, want, got)
			assert.Equal(t, []string{"gpu-next", "gpu"}, videoOutputs(got)[:2])
		})
	}
	assert.Equal(t, defaultVideoOutputArg(), filterMPVArgs([]string{defaultVideoOutputArg()})[0], "the arg whitelist lets the chain through")
}

// TestVideoOutputChainNamesExistInMPV checks the chain against the installed
// mpv's own list, so a misspelt output (which mpv skips silently) is caught.
// The last resorts are optional: not every build has them, and mpv skips the
// missing ones.
func TestVideoOutputChainNamesExistInMPV(t *testing.T) {
	mpv, err := exec.LookPath("mpv")
	if err != nil {
		t.Skip("mpv is not installed")
	}
	out, err := exec.Command(mpv, "--no-config", "--vo=help").Output() // #nosec G204 -- the mpv found on PATH, fixed args
	require.NoError(t, err)

	known := map[string]bool{}
	for _, line := range strings.Split(string(out), "\n") {
		if fields := strings.Fields(line); len(fields) > 0 && strings.HasPrefix(line, "  ") {
			known[fields[0]] = true
		}
	}
	require.True(t, known["gpu"], "mpv --vo=help listed no gpu output:\n%s", out)
	if !known["gpu-next"] {
		t.Skip("this mpv predates gpu-next (0.35); it skips the name and plays on gpu")
	}
	optional := map[string]bool{"libmpv": true, "direct3d": true, "sdl": true}
	for _, vo := range videoOutputs(defaultVideoOutputArg()) {
		assert.True(t, known[vo] || optional[vo], "mpv has no video output %q", vo)
	}
}

func videoOutputs(arg string) []string {
	return strings.Split(strings.TrimPrefix(arg, "--vo="), ",")
}
