//go:build !windows

package player

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mpvSockCounter keeps mock socket paths unique across parallel subtests.
var mpvSockCounter atomic.Uint64

// startMockMPVSocket spins up a unix-socket listener that mimics mpv's JSON
// IPC protocol. handler is invoked once per accepted connection; the bytes it
// returns are written back to the client (a trailing newline is appended).
//
// macOS limits unix-socket paths to ~104 bytes, so the socket lives under
// /tmp (a short symlink) rather than t.TempDir() (which on darwin expands to
// a long /var/folders path).
func startMockMPVSocket(t *testing.T, handler func(req map[string]any) []byte) string {
	t.Helper()
	n := mpvSockCounter.Add(1)
	sock := filepath.Join(string(filepath.Separator), "tmp", fmt.Sprintf("goanime_mpv_%d_%d.sock", os.Getpid(), n))
	require.LessOrEqual(t, len(sock), 100, "socket path too long for unix-socket limit")

	ln, err := net.Listen("unix", sock)
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = ln.Close()
		_ = os.Remove(sock)
	})

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer func() { _ = c.Close() }()
				buf := make([]byte, 4096)
				n, err := c.Read(buf)
				if err != nil {
					return
				}
				var req map[string]any
				_ = json.Unmarshal(bytes.TrimSpace(buf[:n]), &req)
				resp := handler(req)
				if len(resp) == 0 {
					return
				}
				_, _ = c.Write(append(resp, '\n'))
			}(conn)
		}
	}()
	return sock
}

// mpvOK returns a `{"data":<v>,"error":"success"}` payload, or
// `{"error":"success"}` when v is nil (set_property style).
func mpvOK(v any) []byte {
	resp := map[string]any{"error": "success"}
	if v != nil {
		resp["data"] = v
	}
	b, _ := json.Marshal(resp)
	return b
}

func TestMpvSendCommand_RoundTripReturnsData(t *testing.T) {
	t.Parallel()
	var received map[string]any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		received = req
		return mpvOK(123.0)
	})

	got, err := mpvSendCommand(sock, []any{"get_property", "time-pos"})
	require.NoError(t, err)
	assert.Equal(t, 123.0, got)

	cmd, ok := received["command"].([]any)
	require.True(t, ok, "expected command field in IPC payload")
	assert.Equal(t, []any{"get_property", "time-pos"}, cmd)
}

func TestMpvSendCommand_SuccessWithoutData(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte {
		return []byte(`{"error":"success"}`)
	})

	got, err := mpvSendCommand(sock, []any{"set_property", "pause", true})
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestMpvSendCommand_SkipsPropertyUnavailable(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte {
		// Two JSONs separated by newline; first must be skipped, second has
		// no data → surfaces the "no data field" sentinel error.
		return []byte(`{"error":"property unavailable"}` + "\n" + `{"error":"some-other"}`)
	})

	_, err := mpvSendCommand(sock, []any{"get_property", "speed"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no data field")
}

func TestMpvSendCommand_DialError(t *testing.T) {
	t.Parallel()
	_, err := mpvSendCommand("/tmp/goanime_mpv_does_not_exist.sock", []any{"ping"})
	require.Error(t, err)
}

func TestMpvSendCommand_PublicWrapper(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK("hello") })
	got, err := MpvSendCommand(sock, []any{"get_property", "filename"})
	require.NoError(t, err)
	assert.Equal(t, "hello", got)
}

func TestToggleSubtitle(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	require.NoError(t, ToggleSubtitle(sock))
	assert.Equal(t, []any{"cycle", "sub-visibility"}, seen)
}

func TestSetPlaybackSpeed(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	require.NoError(t, SetPlaybackSpeed(sock, 1.5))
	require.Len(t, seen, 3)
	assert.Equal(t, "set_property", seen[0])
	assert.Equal(t, "speed", seen[1])
	assert.InDelta(t, 1.5, seen[2].(float64), 1e-9)
}

func TestCycleAudioTrack(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	require.NoError(t, CycleAudioTrack(sock))
	assert.Equal(t, []any{"cycle", "aid"}, seen)
}

func TestCycleSubtitleTrack(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	require.NoError(t, CycleSubtitleTrack(sock))
	assert.Equal(t, []any{"cycle", "sid"}, seen)
}

func TestSetAudioTrack(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	require.NoError(t, SetAudioTrack(sock, 2))
	require.Len(t, seen, 3)
	assert.Equal(t, "set_property", seen[0])
	assert.Equal(t, "aid", seen[1])
	assert.InDelta(t, float64(2), seen[2].(float64), 1e-9)
}

func TestSetSubtitleTrack(t *testing.T) {
	t.Parallel()
	var seen []any
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		seen, _ = req["command"].([]any)
		return mpvOK(nil)
	})
	require.NoError(t, SetSubtitleTrack(sock, 3))
	require.Len(t, seen, 3)
	assert.Equal(t, "sid", seen[1])
	assert.InDelta(t, float64(3), seen[2].(float64), 1e-9)
}

func TestGetPlaybackStats_AggregatesProperties(t *testing.T) {
	t.Parallel()
	values := map[string]any{
		"time-pos": 42.5,
		"duration": 1440.0,
		"speed":    1.0,
		"volume":   80.0,
		"pause":    false,
		"filename": "ep01.mkv",
	}
	var calls atomic.Int32
	sock := startMockMPVSocket(t, func(req map[string]any) []byte {
		calls.Add(1)
		cmd, _ := req["command"].([]any)
		require.Len(t, cmd, 2)
		prop, _ := cmd[1].(string)
		v, ok := values[prop]
		if !ok {
			return []byte(`{"error":"property unavailable"}`)
		}
		return mpvOK(v)
	})

	stats, err := GetPlaybackStats(sock)
	require.NoError(t, err)
	assert.Equal(t, int32(6), calls.Load(), "must request all 6 properties")
	assert.Equal(t, "ep01.mkv", stats["filename"])
	assert.InDelta(t, 42.5, stats["time-pos"].(float64), 1e-9)
	assert.Equal(t, false, stats["pause"])
}

func TestGetAudioTracks_FiltersByType(t *testing.T) {
	t.Parallel()
	tracks := []any{
		map[string]any{"id": 1.0, "type": "audio", "lang": "ja"},
		map[string]any{"id": 2.0, "type": "audio", "lang": "en"},
		map[string]any{"id": 3.0, "type": "sub", "lang": "en"},
		map[string]any{"id": 4.0, "type": "video"},
	}
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(tracks) })

	got, err := GetAudioTracks(sock)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "ja", got[0]["lang"])
	assert.Equal(t, "en", got[1]["lang"])
}

func TestGetAudioTracks_RejectsBadShape(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK("not-an-array") })
	_, err := GetAudioTracks(sock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected track-list format")
}

func TestGetSubtitleTracks_FiltersByType(t *testing.T) {
	t.Parallel()
	tracks := []any{
		map[string]any{"id": 1.0, "type": "audio"},
		map[string]any{"id": 2.0, "type": "sub", "lang": "pt"},
		map[string]any{"id": 3.0, "type": "sub", "lang": "es"},
	}
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(tracks) })

	got, err := GetSubtitleTracks(sock)
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "pt", got[0]["lang"])
}

func TestGetSubtitleTracks_RejectsBadShape(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(42.0) })
	_, err := GetSubtitleTracks(sock)
	require.Error(t, err)
}

func TestGetCurrentAudioTrack_NumericID(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(5.0) })
	got, err := GetCurrentAudioTrack(sock)
	require.NoError(t, err)
	assert.Equal(t, 5, got)
}

func TestGetCurrentAudioTrack_UnexpectedType(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK("no") })
	_, err := GetCurrentAudioTrack(sock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected aid format")
}

func TestGetCurrentSubtitleTrack_NumericID(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(2.0) })
	got, err := GetCurrentSubtitleTrack(sock)
	require.NoError(t, err)
	assert.Equal(t, 2, got)
}

func TestGetCurrentSubtitleTrack_StringMeansNoSub(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK("no") })
	got, err := GetCurrentSubtitleTrack(sock)
	require.NoError(t, err)
	assert.Equal(t, 0, got)
}

func TestGetCurrentSubtitleTrack_UnexpectedType(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(true) })
	_, err := GetCurrentSubtitleTrack(sock)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected sid format")
}

func TestDialMPVSocket_ConnectsToExistingListener(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(nil) })
	conn, err := dialMPVSocket(sock)
	require.NoError(t, err)
	require.NoError(t, conn.Close())
}

func TestDialMPVSocket_ErrorOnMissingSocket(t *testing.T) {
	t.Parallel()
	_, err := dialMPVSocket("/tmp/goanime_mpv_does_not_exist.sock")
	require.Error(t, err)
}
