package discord

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tr1xem/go-discordrpc/client"
)

// stubRPC implements rpcClient for tests.
type stubRPC struct {
	loginErr        error
	logoutErr       error
	setActivityErr  error
	loginCount      int
	logoutCount     int
	setActivityHits int
	lastActivity    client.Activity
	mu              sync.Mutex
}

func (s *stubRPC) Login() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loginCount++
	return s.loginErr
}
func (s *stubRPC) Logout() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logoutCount++
	return s.logoutErr
}
func (s *stubRPC) SetActivity(a client.Activity) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.setActivityHits++
	s.lastActivity = a
	return s.setActivityErr
}

// withStubClient swaps newRPCClient to return a stub and resets globals. Tests
// using this MUST NOT run with t.Parallel() because globals are package-scoped.
func withStubClient(t *testing.T, stub *stubRPC) {
	t.Helper()
	prevFactory := newRPCClient
	clientMutex.Lock()
	prevClient := discordClient
	prevLogged := isLoggedIn
	clientMutex.Unlock()

	newRPCClient = func(string) rpcClient { return stub }
	clientMutex.Lock()
	discordClient = nil
	isLoggedIn = false
	clientMutex.Unlock()

	t.Cleanup(func() {
		newRPCClient = prevFactory
		clientMutex.Lock()
		discordClient = prevClient
		isLoggedIn = prevLogged
		clientMutex.Unlock()
	})
}

func TestLoginClient_Success(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)
	require.NoError(t, LoginClient())
	assert.True(t, IsClientLoggedIn())
	assert.Equal(t, 1, stub.loginCount)
}

func TestLoginClient_IdempotentWhenLoggedIn(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)
	require.NoError(t, LoginClient())
	require.NoError(t, LoginClient())
	assert.Equal(t, 1, stub.loginCount, "second Login must short-circuit")
}

func TestLoginClient_PropagatesError(t *testing.T) {
	stub := &stubRPC{loginErr: errors.New("boom")}
	withStubClient(t, stub)
	err := LoginClient()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "discord login failed")
	assert.False(t, IsClientLoggedIn())
}

func TestLogoutClient_Success(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)
	require.NoError(t, LoginClient())
	require.NoError(t, LogoutClient())
	assert.False(t, IsClientLoggedIn())
	assert.Equal(t, 1, stub.logoutCount)
}

func TestLogoutClient_NoOpWhenNotLoggedIn(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)
	// never login
	require.NoError(t, LogoutClient())
	assert.Equal(t, 0, stub.logoutCount)
}

func TestLogoutClient_PropagatesError(t *testing.T) {
	stub := &stubRPC{logoutErr: errors.New("nope")}
	withStubClient(t, stub)
	require.NoError(t, LoginClient())
	err := LogoutClient()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "discord logout failed")
}

// --- GetCurrentPlaybackPosition ---

func TestGetCurrentPlaybackPosition_Success(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(_ string, _ []any) (any, error) { return 42.5, nil }
	rpu.socketPath = "/x"
	got, err := rpu.GetCurrentPlaybackPosition()
	require.NoError(t, err)
	assert.Equal(t, 42*time.Second, got)
}

func TestGetCurrentPlaybackPosition_MPVError(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(_ string, _ []any) (any, error) { return nil, errors.New("mpv down") }
	_, err := rpu.GetCurrentPlaybackPosition()
	require.Error(t, err)
}

func TestGetCurrentPlaybackPosition_BadType(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(_ string, _ []any) (any, error) { return "not-a-number", nil }
	_, err := rpu.GetCurrentPlaybackPosition()
	require.Error(t, err)
}

// --- Start / Stop ---

func TestStart_FailsLogin_NoTickerStarted(t *testing.T) {
	stub := &stubRPC{loginErr: errors.New("auth fail")}
	withStubClient(t, stub)

	a := &models.Anime{Name: "x", Episodes: []models.Episode{{Number: "1"}}}
	rpu := NewRichPresenceUpdater(a, ptrBool(false), &sync.Mutex{}, 10*time.Millisecond, time.Second,
		"/sock", func(string, []any) (any, error) { return 0.0, nil })
	rpu.Start()
	// no goroutine started: closing done would normally race with goroutine,
	// here it's safe because Start exited early.
	rpu.Stop()
	assert.Equal(t, 0, stub.setActivityHits)
}

func TestStartAndStop_PerformsUpdates(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)

	calls := 0
	mpv := func(_ string, args []any) (any, error) {
		calls++
		prop := args[1].(string)
		switch prop {
		case "time-pos":
			return 10.0, nil
		case "duration":
			return 1200.0, nil
		case "pause":
			return false, nil
		case "speed":
			return 1.0, nil
		}
		return nil, nil
	}
	a := &models.Anime{
		Name:     "Foo",
		Episodes: []models.Episode{{Number: "3"}},
		Details:  models.AniListDetails{Title: models.Title{Romaji: "Foo"}},
	}
	rpu := NewRichPresenceUpdater(a, ptrBool(false), &sync.Mutex{}, 10*time.Millisecond, 1200*time.Second, "/sock", mpv)

	// reset global timing state so updater is allowed to fire
	clientMutex.Lock()
	lastUpdateTime = time.Time{}
	lastForceUpdateTime = time.Time{}
	clientMutex.Unlock()

	rpu.Start()
	// wait long enough for initial + at least one tick
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		stub.mu.Lock()
		hits := stub.setActivityHits
		stub.mu.Unlock()
		if hits >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	rpu.Stop()

	stub.mu.Lock()
	defer stub.mu.Unlock()
	assert.GreaterOrEqual(t, stub.setActivityHits, 1)
	assert.True(t, calls > 0)
}

func TestStop_NilSafeAndDoubleClose(t *testing.T) {
	t.Parallel()
	var nilRPU *RichPresenceUpdater
	assert.NotPanics(t, func() { nilRPU.Stop() })

	r := &RichPresenceUpdater{done: make(chan bool)}
	assert.NotPanics(t, r.Stop)
	// second stop must not panic on already-closed channel
	assert.NotPanics(t, r.Stop)
}

// --- updateDiscordPresence ---

func TestUpdateDiscordPresence_SkipsWhenMutexBusy(t *testing.T) {
	t.Parallel()
	mu := &sync.Mutex{}
	mu.Lock()
	defer mu.Unlock()
	a := &models.Anime{}
	rpu := &RichPresenceUpdater{
		anime:      a,
		isPaused:   ptrBool(false),
		animeMutex: mu,
		mpvSendCommand: func(string, []any) (any, error) {
			t.Fatalf("must not call MPV when mutex is busy")
			return nil, nil
		},
	}
	assert.NotPanics(t, func() { rpu.updateDiscordPresence(true) })
}

func TestUpdateDiscordPresence_SkipsWhenPlaybackStateNil(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return nil, errors.New("x") }
	rpu.updateDiscordPresence(true) // must not panic
}

func TestUpdateDiscordPresence_PublishesActivity(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)
	require.NoError(t, LoginClient())

	clientMutex.Lock()
	lastUpdateTime = time.Time{}
	lastForceUpdateTime = time.Time{}
	clientMutex.Unlock()

	a := &models.Anime{
		Name:      "MovieX",
		MediaType: "movie",
		IMDBID:    "tt1",
		Episodes:  []models.Episode{{Number: "1"}},
	}
	rpu := &RichPresenceUpdater{
		anime:      a,
		isPaused:   ptrBool(false),
		animeMutex: &sync.Mutex{},
		mpvSendCommand: func(_ string, args []any) (any, error) {
			switch args[1].(string) {
			case "time-pos":
				return 5.0, nil
			case "duration":
				return 100.0, nil
			case "pause":
				return false, nil
			case "speed":
				return 1.0, nil
			}
			return nil, nil
		},
	}
	rpu.updateDiscordPresence(true)

	stub.mu.Lock()
	defer stub.mu.Unlock()
	require.Equal(t, 1, stub.setActivityHits)
	assert.Equal(t, "MovieX", stub.lastActivity.Name)
	assert.Equal(t, "Watching a movie", stub.lastActivity.State)
}

// --- getPrecisePlaybackState ---

func TestGetPrecisePlaybackState_AllPropsOK(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(_ string, args []any) (any, error) {
		switch args[1].(string) {
		case "time-pos":
			return 30.0, nil
		case "duration":
			return 600.0, nil
		case "pause":
			return true, nil
		case "speed":
			return 1.5, nil
		}
		return nil, nil
	}
	st := rpu.getPrecisePlaybackState()
	require.NotNil(t, st)
	assert.Equal(t, 30, st.positionSec)
	assert.Equal(t, 600, st.durationSec)
	assert.True(t, st.isPaused)
	assert.InDelta(t, 1.5, st.speed, 0.01)
}

func TestGetPrecisePlaybackState_PositionErrorReturnsNil(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return nil, errors.New("x") }
	assert.Nil(t, rpu.getPrecisePlaybackState())
}

func TestGetPrecisePlaybackState_DurationFromCache(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.episodeDuration = 200 * time.Second
	rpu.mpvSendCommand = func(_ string, args []any) (any, error) {
		switch args[1].(string) {
		case "time-pos":
			return 1.0, nil
		case "pause":
			return false, nil
		case "speed":
			return 1.0, nil
		case "duration":
			t.Fatalf("duration should not be queried when cached")
		}
		return nil, nil
	}
	st := rpu.getPrecisePlaybackState()
	require.NotNil(t, st)
	assert.Equal(t, 200, st.durationSec)
}

func TestGetPrecisePlaybackState_BadPositionType(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(_ string, args []any) (any, error) {
		if args[1].(string) == "time-pos" {
			return "bad", nil
		}
		return nil, nil
	}
	assert.Nil(t, rpu.getPrecisePlaybackState())
}

// --- buildPreciseTimestamps ---

func TestBuildPreciseTimestamps_PausedNoEnd(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	now := time.Now()
	ts := rpu.buildPreciseTimestamps(&playbackState{positionSec: 10, isPaused: true}, now)
	require.NotNil(t, ts)
	assert.NotNil(t, ts.Start)
	assert.Nil(t, ts.End)
}

func TestBuildPreciseTimestamps_PlayingWithDuration(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	now := time.Now()
	ts := rpu.buildPreciseTimestamps(&playbackState{positionSec: 20, durationSec: 120, speed: 1.0}, now)
	require.NotNil(t, ts)
	require.NotNil(t, ts.End)
	// end ≈ now + 100s
	assert.WithinDuration(t, now.Add(100*time.Second), *ts.End, 2*time.Second)
}

func TestBuildPreciseTimestamps_DurationUnknown(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	now := time.Now()
	ts := rpu.buildPreciseTimestamps(&playbackState{positionSec: 20, durationSec: 0, speed: 1.0}, now)
	require.NotNil(t, ts)
	assert.NotNil(t, ts.Start)
	assert.Nil(t, ts.End)
}

func TestBuildPreciseTimestamps_SpeedAccelerates(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	now := time.Now()
	ts := rpu.buildPreciseTimestamps(&playbackState{positionSec: 0, durationSec: 200, speed: 2.0}, now)
	require.NotNil(t, ts.End)
	// remaining 200 / 2 = 100s
	assert.WithinDuration(t, now.Add(100*time.Second), *ts.End, 2*time.Second)
}

// --- FetchDuration ---

func TestFetchDuration_CallsCallback(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return 1200.0, nil }
	var got int
	rpu.FetchDuration("/sock", func(d int) { got = d })
	assert.Equal(t, 1200, got)
}

func TestFetchDuration_UsesFallbackPath(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.socketPath = "/fallback"
	var seenPath string
	rpu.mpvSendCommand = func(p string, _ []any) (any, error) { seenPath = p; return 60.0, nil }
	var got int
	rpu.FetchDuration("", func(d int) { got = d })
	assert.Equal(t, "/fallback", seenPath)
	assert.Equal(t, 60, got)
}

func TestFetchDuration_MPVError(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return nil, errors.New("x") }
	called := false
	rpu.FetchDuration("/s", func(int) { called = true })
	assert.False(t, called)
}

func TestFetchDuration_NilResponse(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return nil, nil }
	called := false
	rpu.FetchDuration("/s", func(int) { called = true })
	assert.False(t, called)
}

func TestFetchDuration_BadType(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return "string", nil }
	called := false
	rpu.FetchDuration("/s", func(int) { called = true })
	assert.False(t, called)
}

func TestFetchDuration_ZeroIgnored(t *testing.T) {
	t.Parallel()
	rpu := mkRPU(&models.Anime{})
	rpu.mpvSendCommand = func(string, []any) (any, error) { return 0.0, nil }
	called := false
	rpu.FetchDuration("/s", func(int) { called = true })
	assert.False(t, called)
}

// --- Initialize ---

func TestInitialize_AlreadyInitialized_NoOp(t *testing.T) {
	t.Parallel()
	m := NewManager()
	m.isInitialized = true
	require.NoError(t, m.Initialize())
}

func TestInitialize_BackgroundLoginFails(t *testing.T) {
	stub := &stubRPC{loginErr: errors.New("offline")}
	withStubClient(t, stub)
	m := NewManager()
	require.NoError(t, m.Initialize())

	require.Eventually(t, m.IsInitialized, 1*time.Second, 10*time.Millisecond)
	assert.False(t, m.IsEnabled())
}

func TestInitialize_BackgroundLoginSucceeds(t *testing.T) {
	stub := &stubRPC{}
	withStubClient(t, stub)
	m := NewManager()
	require.NoError(t, m.Initialize())

	require.Eventually(t, m.IsEnabled, 1*time.Second, 10*time.Millisecond)
	assert.True(t, m.IsInitialized())
}

func ptrBool(b bool) *bool { return &b }
