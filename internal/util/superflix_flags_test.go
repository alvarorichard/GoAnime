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

func TestFlagParser_SuperFlixFlagsAbsentLeaveEnvUntouched(t *testing.T) {
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	clearSFEnv(t)

	os.Args = []string{"goanime", "naruto"}

	_, err := FlagParser()
	require.NoError(t, err)

	// No --sf-* flags: an unset flag must never clobber the env (so a manually
	// exported value survives). We blanked them above, so they must stay blank.
	for _, k := range sfEnvKeys {
		assert.Empty(t, os.Getenv(k), "%s must remain unset when its flag is absent", k)
	}
}
