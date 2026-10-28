package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMediaContextRoundTrip(t *testing.T) {
	SetMediaContext("Mukashi Banashi", "Animefire.io", "Legendado")
	ctx := snapshotMediaContext()
	assert.Equal(t, "Mukashi Banashi", ctx.animeTitle)
	assert.Equal(t, "Animefire.io", ctx.source)
	assert.Equal(t, "Legendado", ctx.audio)
	SetMediaContext("", "", "")
}

func TestResumeOverrideOnce(t *testing.T) {
	ClearResumeOverride()
	_, _, ok := takeResumeOverride()
	assert.False(t, ok, "nothing pending")

	ForceResumeNext(372)
	position, fromStart, ok := takeResumeOverride()
	assert.True(t, ok)
	assert.False(t, fromStart)
	assert.Equal(t, 372, position)

	_, _, ok = takeResumeOverride()
	assert.False(t, ok, "consumed once")

	ForceStartNext()
	position, fromStart, ok = takeResumeOverride()
	assert.True(t, ok)
	assert.True(t, fromStart)
	assert.Equal(t, 0, position)

	ClearResumeOverride()
	ForceResumeNext(10)
	ClearResumeOverride()
	_, _, ok = takeResumeOverride()
	assert.False(t, ok, "cleared before use")
}
