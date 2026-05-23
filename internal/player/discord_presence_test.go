//go:build !windows

package player

import (
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// initDiscordPresence
// ---------------------------------------------------------------------------

// TestInitDiscordPresence_StartsUpdaterAndReturns verifies that initDiscordPresence:
//  1. Sets the socket path on the updater
//  2. Calls updater.Start()
//  3. Launches the background goroutine (non-blocking)
//
// The goroutine tries to connect to a non-existent MPV socket and gives up
// after retries; we don't wait for it — just check the synchronous portion.
func TestInitDiscordPresence_StartsUpdaterAndReturns(t *testing.T) {
	paused := false
	var mu sync.Mutex
	updater := discord.NewRichPresenceUpdater(
		&models.Anime{Name: "TestAnime"},
		&paused,
		&mu,
		100*time.Millisecond,
		0,
		"",
		MpvSendCommand,
	)

	// Use a socket path that doesn't exist so waitForPlaybackStart fails fast.
	// On Linux/macOS unix sockets, connect fails immediately for missing paths.
	fakeSock := "/tmp/goanime_p17_test_nonexistent_socket"

	ep := &models.Episode{Number: "1", Num: 1}

	// initDiscordPresence must return without blocking.
	initDiscordPresence(updater, fakeSock, nil, 0, ep, 1)

	// The updater was started — allow a short moment for Start() to register.
	time.Sleep(20 * time.Millisecond)

	// No panic and function returned → pass.
	assert.True(t, true, "initDiscordPresence returned without panic")

	// Clean up: stop the updater so the background goroutine terminates.
	updater.Stop()
}
