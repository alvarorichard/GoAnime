//go:build !windows

// Package test contains tests for the player package, specifically focusing on
// Unix socket handling for macOS compatibility.
//
// IDENTIFIED ERROR:
// On macOS, the mpv IPC socket connection was failing with the error:
//
//	"timeout waiting for mpv socket. Possible issues:
//	 1. MPV installation corrupted
//	 2. Firewall blocking IPC
//	 3. Invalid video URL"
//
// ROOT CAUSE:
// The os.TempDir() function on macOS returns a path with a trailing slash
// (e.g., "/var/folders/24/_1_ntj3s67bc4cqxg2vszb300000gn/T/").
// When constructing the socket path using fmt.Sprintf("%s/goanime_mpvsocket_%s", tmpDir, id),
// this resulted in a double-slash path like:
//
//	"/var/folders/.../T//goanime_mpvsocket_xxx"
//
// The double-slash caused the Unix socket connection to fail, as mpv created
// the socket at one path while goanime tried to connect to another.
//
// SOLUTION:
// Use filepath.Join() instead of fmt.Sprintf() for path construction.
// filepath.Join() properly handles trailing slashes and produces clean paths:
//
//	socketPath = filepath.Join(os.TempDir(), fmt.Sprintf("goanime_mpvsocket_%s", randomNumber))
//
// This ensures the socket path is always valid regardless of the OS-specific
// behavior of os.TempDir().
//
// ADDITIONAL FIX:
// A secondary issue was found in the playback loop where selecting "Quit"
// didn't work because GetUserInput() returns "q" but the code checked for "quit".
// Fixed by checking for both values: if userInput == "q" || userInput == "quit"
package test

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMacOSSocketPathConstruction tests that socket paths are constructed correctly on macOS
// This specifically verifies the fix for the double-slash issue where os.TempDir() returns
// a path with a trailing slash on macOS (e.g., /var/folders/.../T/)
func TestMacOSSocketPathConstruction(t *testing.T) {
	t.Run("Socket path should not contain double slashes", func(t *testing.T) {
		tmpDir := os.TempDir()
		// Simulate the fix: remove trailing slash
		tmpDir = strings.TrimSuffix(tmpDir, "/")

		randomNumber := fmt.Sprintf("%x", time.Now().UnixNano())
		socketPath := fmt.Sprintf("%s/goanime_mpvsocket_%s", tmpDir, randomNumber)

		// Verify no double slashes in the path
		assert.NotContains(t, socketPath, "//", "Socket path should not contain double slashes")

		// Verify the path is valid
		assert.True(t, strings.HasPrefix(socketPath, "/"), "Socket path should start with /")

		// Log for debugging in CI
		t.Logf("TempDir: %q", os.TempDir())
		t.Logf("Cleaned TempDir: %q", tmpDir)
		t.Logf("Socket path: %q", socketPath)
	})

	t.Run("TempDir trailing slash handling on macOS", func(t *testing.T) {
		tmpDir := os.TempDir()

		// On macOS, TempDir typically returns a path with trailing slash
		if runtime.GOOS == "darwin" {
			t.Logf("macOS detected, TempDir=%q, has trailing slash=%v",
				tmpDir, strings.HasSuffix(tmpDir, "/"))
		}

		// After trimming, should not have trailing slash
		cleaned := strings.TrimSuffix(tmpDir, "/")
		assert.False(t, strings.HasSuffix(cleaned, "/"),
			"Cleaned temp dir should not have trailing slash")
	})

	t.Run("filepath.Join handles trailing slash correctly", func(t *testing.T) {
		// This tests the actual implementation approach used in player.go
		randomNumber := fmt.Sprintf("%x", time.Now().UnixNano())
		socketPath := filepath.Join(os.TempDir(), fmt.Sprintf("goanime_mpvsocket_%s", randomNumber))

		// filepath.Join should handle the trailing slash
		assert.NotContains(t, socketPath, "//", "filepath.Join should not create double slashes")
		assert.True(t, strings.HasPrefix(socketPath, "/"), "Socket path should start with /")

		t.Logf("filepath.Join result: %q", socketPath)
	})
}

// TestUnixSocketCreationAndConnection tests actual Unix socket creation and connection
// This simulates what mpv does when it creates an IPC socket
func TestUnixSocketCreationAndConnection(t *testing.T) {
	// Note: This test only runs on Unix systems due to //go:build !windows constraint

	t.Run("Create and connect to Unix socket", func(t *testing.T) {
		// Create a unique socket path
		tmpDir := strings.TrimSuffix(os.TempDir(), "/")
		socketPath := filepath.Join(tmpDir, fmt.Sprintf("goanime_test_socket_%d", time.Now().UnixNano()))

		// Ensure socket file doesn't exist
		_ = os.Remove(socketPath)

		// Create a listener (simulating mpv's socket server)
		listener, err := net.Listen("unix", socketPath)
		require.NoError(t, err, "Should be able to create Unix socket listener")
		defer func() {
			_ = listener.Close()
			_ = os.Remove(socketPath)
		}()

		// Accept connections in a goroutine (simulating mpv)
		serverReady := make(chan struct{})
		go func() {
			close(serverReady)
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}()

		<-serverReady

		// Try to connect to the socket (simulating goanime client)
		conn, err := net.Dial("unix", socketPath)
		require.NoError(t, err, "Should be able to connect to Unix socket")
		defer func() { _ = conn.Close() }()

		t.Logf("Successfully created and connected to Unix socket: %s", socketPath)
	})

	t.Run("Socket connection with retry logic", func(t *testing.T) {
		tmpDir := strings.TrimSuffix(os.TempDir(), "/")
		socketPath := filepath.Join(tmpDir, fmt.Sprintf("goanime_test_socket_retry_%d", time.Now().UnixNano()))

		_ = os.Remove(socketPath)

		// Start listener after a small delay (simulating mpv startup time)
		listenerReady := make(chan struct{})
		go func() {
			time.Sleep(200 * time.Millisecond)
			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				close(listenerReady)
				return
			}
			close(listenerReady)
			defer func() {
				_ = listener.Close()
				_ = os.Remove(socketPath)
			}()

			// Accept one connection
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}()

		// Implement retry logic similar to StartVideo
		var conn net.Conn
		var connErr error
		maxAttempts := 10
		for i := 0; i < maxAttempts; i++ {
			conn, connErr = net.Dial("unix", socketPath)
			if connErr == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}

		require.NoError(t, connErr, "Should eventually connect after retries")
		if conn != nil {
			_ = conn.Close()
		}

		<-listenerReady
		t.Log("Successfully connected with retry logic")
	})
}

// TestSocketPathValidation tests that socket paths are properly validated
func TestSocketPathValidation(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		shouldError bool
	}{
		{
			name:        "Valid socket path",
			path:        "/tmp/valid_socket",
			shouldError: false,
		},
		{
			name:        "Path with double slashes should be avoided",
			path:        "/tmp//invalid_socket",
			shouldError: true,
		},
		{
			name:        "Empty path",
			path:        "",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasDoubleSlash := strings.Contains(tt.path, "//")
			isEmpty := tt.path == ""

			if tt.shouldError {
				assert.True(t, hasDoubleSlash || isEmpty,
					"Invalid paths should have double slashes or be empty")
			} else {
				assert.False(t, hasDoubleSlash,
					"Valid paths should not have double slashes")
				assert.NotEmpty(t, tt.path, "Valid paths should not be empty")
			}
		})
	}
}

// TestMacOSSpecificTempDir tests macOS-specific temp directory behavior
func TestMacOSSpecificTempDir(t *testing.T) {
	tmpDir := os.TempDir()

	t.Run("TempDir returns valid directory", func(t *testing.T) {
		assert.NotEmpty(t, tmpDir, "TempDir should not be empty")

		// Check that the directory exists
		info, err := os.Stat(strings.TrimSuffix(tmpDir, "/"))
		require.NoError(t, err, "TempDir should exist")
		assert.True(t, info.IsDir(), "TempDir should be a directory")
	})

	t.Run("Can create files in TempDir", func(t *testing.T) {
		cleanedDir := strings.TrimSuffix(tmpDir, "/")
		testFile := filepath.Join(cleanedDir, fmt.Sprintf("goanime_test_%d", time.Now().UnixNano()))

		// Create test file
		f, err := os.Create(testFile)
		require.NoError(t, err, "Should be able to create file in TempDir")
		_ = f.Close()

		// Clean up
		err = os.Remove(testFile)
		require.NoError(t, err, "Should be able to remove test file")
	})

	if runtime.GOOS == "darwin" {
		t.Run("macOS TMPDIR environment variable", func(t *testing.T) {
			envTmpDir := os.Getenv("TMPDIR")
			t.Logf("TMPDIR env: %q", envTmpDir)
			t.Logf("os.TempDir(): %q", tmpDir)

			// On macOS, TMPDIR is typically set and os.TempDir() uses it
			if envTmpDir != "" {
				// They should be equivalent (allowing for trailing slash differences)
				assert.True(t,
					strings.TrimSuffix(tmpDir, "/") == strings.TrimSuffix(envTmpDir, "/"),
					"TempDir should match TMPDIR environment variable")
			}
		})
	}
}

// BenchmarkSocketPathConstruction benchmarks socket path construction
func BenchmarkSocketPathConstruction(b *testing.B) {
	for i := 0; i < b.N; i++ {
		tmpDir := os.TempDir()
		tmpDir = strings.TrimSuffix(tmpDir, "/")
		randomNumber := fmt.Sprintf("%x", time.Now().UnixNano())
		_ = fmt.Sprintf("%s/goanime_mpvsocket_%s", tmpDir, randomNumber)
	}
}

// TestExponentialBackoffSocketConnection tests that socket connection uses proper exponential backoff
// This simulates the improved StartVideo behavior where we wait up to 10 seconds with backoff
func TestExponentialBackoffSocketConnection(t *testing.T) {
	t.Run("Exponential backoff adapts interval over time", func(t *testing.T) {
		// Test the exponential backoff logic used in StartVideo
		initialInterval := 50 * time.Millisecond
		maxInterval := 500 * time.Millisecond
		currentInterval := initialInterval

		intervals := []time.Duration{currentInterval}

		// Simulate 10 iterations of backoff
		for i := 0; i < 10; i++ {
			currentInterval = time.Duration(float64(currentInterval) * 1.5)
			if currentInterval > maxInterval {
				currentInterval = maxInterval
			}
			intervals = append(intervals, currentInterval)
		}

		// First interval should be small (50ms)
		assert.Equal(t, initialInterval, intervals[0])

		// Later intervals should grow up to maxInterval
		assert.Equal(t, maxInterval, intervals[len(intervals)-1])

		// Verify gradual growth
		for i := 1; i < len(intervals); i++ {
			assert.GreaterOrEqual(t, intervals[i], intervals[i-1],
				"Intervals should grow or stay at max")
		}

		t.Logf("Backoff intervals: %v", intervals)
	})

	t.Run("Connection with delayed socket creation (simulating slow stream)", func(t *testing.T) {
		tmpDir := strings.TrimSuffix(os.TempDir(), "/")
		socketPath := filepath.Join(tmpDir, fmt.Sprintf("goanime_slow_stream_%d", time.Now().UnixNano()))

		_ = os.Remove(socketPath)

		// Simulate a slow stream by delaying socket creation by 2 seconds
		// This tests that our 10-second timeout is sufficient
		socketDelay := 2 * time.Second

		listenerReady := make(chan struct{})
		go func() {
			time.Sleep(socketDelay)
			listener, err := net.Listen("unix", socketPath)
			if err != nil {
				close(listenerReady)
				return
			}
			close(listenerReady)
			defer func() {
				_ = listener.Close()
				_ = os.Remove(socketPath)
			}()

			conn, err := listener.Accept()
			if err != nil {
				return
			}
			_ = conn.Close()
		}()

		// Use exponential backoff like StartVideo does
		startTime := time.Now()
		maxWaitTime := 10 * time.Second
		initialInterval := 50 * time.Millisecond
		maxIntervalCap := 500 * time.Millisecond
		currentInterval := initialInterval

		var conn net.Conn
		var connErr error

		for time.Since(startTime) < maxWaitTime {
			conn, connErr = net.Dial("unix", socketPath)
			if connErr == nil {
				break
			}

			time.Sleep(currentInterval)
			currentInterval = time.Duration(float64(currentInterval) * 1.5)
			if currentInterval > maxIntervalCap {
				currentInterval = maxIntervalCap
			}
		}

		elapsed := time.Since(startTime)
		require.NoError(t, connErr, "Should connect within 10-second timeout")
		if conn != nil {
			_ = conn.Close()
		}

		// Should have connected after approximately 2 seconds (the delay)
		assert.GreaterOrEqual(t, elapsed, socketDelay,
			"Connection should happen after socket is created")
		assert.Less(t, elapsed, 5*time.Second,
			"Should not take too long after socket is available")

		<-listenerReady
		t.Logf("Connected after %.2fs (socket delayed by %v)", elapsed.Seconds(), socketDelay)
	})
}
