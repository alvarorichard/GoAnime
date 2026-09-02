package util

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sfEnvKeys are the env vars the --sf-* flags drive.
var sfEnvKeys = []string{
	"GOANIME_SF_HEADLESS",
	"GOANIME_SF_BUNDLED",
	"GOANIME_SF_CHROME_CHANNEL",
	"GOANIME_SF_MASK",
	"GOANIME_SF_OFFSCREEN",
}

// clearSFEnv blanks every SuperFlix env var for the duration of the test
// (t.Setenv restores the originals on cleanup), isolating the assertions from
// whatever the developer happens to have exported.
func clearSFEnv(t *testing.T) {
	t.Helper()
	for _, k := range sfEnvKeys {
		t.Setenv(k, "")
	}
}

func TestFlagParser_SuperFlixFlagsSetEnv(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	clearSFEnv(t)

	os.Args = []string{
		"goanime",
		"--sf-headless",
		"--sf-bundled",
		"--sf-browser", "msedge",
		"--sf-mask",
		"naruto",
	}

	name, err := FlagParser()
	require.NoError(t, err)
	assert.Equal(t, "naruto", name)

	assert.Equal(t, "1", os.Getenv("GOANIME_SF_HEADLESS"))
	assert.Equal(t, "1", os.Getenv("GOANIME_SF_BUNDLED"))
	assert.Equal(t, "msedge", os.Getenv("GOANIME_SF_CHROME_CHANNEL"))
	assert.Equal(t, "1", os.Getenv("GOANIME_SF_MASK"))
}

// TestFlagParser_SFWindowOptsOutOfHiding covers the opt-out from the hidden
// default. --sf-window has to write an explicitly falsey value, not just leave
// the var unset: unset means "use the default", which is now hiding.
func TestFlagParser_SFWindowOptsOutOfHiding(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	clearSFEnv(t)

	os.Args = []string{"goanime", "--sf-window", "naruto"}

	name, err := FlagParser()
	require.NoError(t, err)
	assert.Equal(t, "naruto", name)
	assert.Equal(t, "0", os.Getenv("GOANIME_SF_OFFSCREEN"),
		"--sf-window must actively disable hiding, not merely omit the flag")
}

// --sf-offscreen still works even though it now names the default, so commands
// and scripts written before the switch keep behaving the same.
func TestFlagParser_SFOffscreenStillAccepted(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	clearSFEnv(t)

	os.Args = []string{"goanime", "--sf-offscreen", "naruto"}

	_, err := FlagParser()
	require.NoError(t, err)
	assert.Equal(t, "1", os.Getenv("GOANIME_SF_OFFSCREEN"))
}

// Passing both is contradictory; showing the window is the safer resolution,
// since a user who asks to see it can always be shown it, while wrongly hiding
// it can strand them on a captcha.
func TestFlagParser_SFWindowWinsOverSFOffscreen(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	clearSFEnv(t)

	os.Args = []string{"goanime", "--sf-offscreen", "--sf-window", "naruto"}

	_, err := FlagParser()
	require.NoError(t, err)
	assert.Equal(t, "0", os.Getenv("GOANIME_SF_OFFSCREEN"),
		"a contradictory pair must resolve to showing the window")
}

func TestFlagParser_SuperFlixFlagsAbsentLeaveEnvUntouched(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	clearSFEnv(t)

	os.Args = []string{"goanime", "naruto"}

	_, err := FlagParser()
	require.NoError(t, err)

	// No --sf-* flags: an unset flag must never clobber the env (so a manually
	// exported value survives). We blanked them above, so they must stay blank.
	// GOANIME_SF_OFFSCREEN included: blank there means "use the default"
	// (hiding), which is exactly what no flags should produce.
	for _, k := range sfEnvKeys {
		assert.Empty(t, os.Getenv(k), "%s must remain unset when its flag is absent", k)
	}
}
