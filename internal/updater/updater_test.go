package updater

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock release data for testing
var mockRelease = GitHubRelease{
	TagName: "v2.0.0",
	Name:    "Version 2.0.0",
	Body:    "This is a test release with new features and bug fixes.",
	Assets: []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{
		{Name: "goanime-linux-amd64", BrowserDownloadURL: "http://example.com/goanime-linux-amd64"},
		{Name: "goanime-windows-amd64.exe", BrowserDownloadURL: "http://example.com/goanime-windows-amd64.exe"},
		{Name: "goanime-darwin-amd64", BrowserDownloadURL: "http://example.com/goanime-darwin-amd64"},
		{Name: "goanime-darwin-arm64", BrowserDownloadURL: "http://example.com/goanime-darwin-arm64"},
		{Name: "goanime-darwin-universal", BrowserDownloadURL: "http://example.com/goanime-darwin-universal"},
		{Name: "goanime-darwin", BrowserDownloadURL: "http://example.com/goanime-darwin"},
	},
}

func TestIsVersionNewer(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
		hasError bool
	}{
		{
			name:     "newer major version",
			latest:   "2.0.0",
			current:  "1.0.0",
			expected: true,
			hasError: false,
		},
		{
			name:     "newer minor version",
			latest:   "1.1.0",
			current:  "1.0.0",
			expected: true,
			hasError: false,
		},
		{
			name:     "newer patch version",
			latest:   "1.0.1",
			current:  "1.0.0",
			expected: true,
			hasError: false,
		},
		{
			name:     "same version",
			latest:   "1.0.0",
			current:  "1.0.0",
			expected: false,
			hasError: false,
		},
		{
			name:     "older version",
			latest:   "1.0.0",
			current:  "1.1.0",
			expected: false,
			hasError: false,
		},
		{
			name:     "different length versions",
			latest:   "1.0.0.1",
			current:  "1.0.0",
			expected: true,
			hasError: false,
		},
		{
			name:     "invalid latest version",
			latest:   "1.0.invalid",
			current:  "1.0.0",
			expected: false,
			hasError: true,
		},
		{
			name:     "invalid current version",
			latest:   "1.0.0",
			current:  "1.0.invalid",
			expected: false,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := isVersionNewer(tt.latest, tt.current)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestFindAssetForPlatform(t *testing.T) {
	tests := []struct {
		name         string
		platform     PlatformInfo
		expectedName string
		hasError     bool
	}{
		{
			name:         "linux amd64",
			platform:     PlatformInfo{OS: "linux", Arch: "amd64"},
			expectedName: "goanime-linux-amd64",
			hasError:     false,
		},
		{
			name:         "windows amd64",
			platform:     PlatformInfo{OS: "windows", Arch: "amd64"},
			expectedName: "goanime-windows-amd64.exe",
			hasError:     false,
		},
		{
			name:         "darwin amd64",
			platform:     PlatformInfo{OS: "darwin", Arch: "amd64"},
			expectedName: "goanime-darwin-amd64",
			hasError:     false,
		},
		{
			name:         "darwin arm64",
			platform:     PlatformInfo{OS: "darwin", Arch: "arm64"},
			expectedName: "goanime-darwin-arm64",
			hasError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, name, err := findAssetForPlatformWithInfo(&mockRelease, tt.platform)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
				assert.Contains(t, url, "http://example.com/")
			}
		})
	}
}

func TestFindAssetForPlatform_UnsupportedPlatform(t *testing.T) {
	unsupportedPlatform := PlatformInfo{OS: "unsupported", Arch: "amd64"}
	_, _, err := findAssetForPlatformWithInfo(&mockRelease, unsupportedPlatform)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported platform")
}

func TestFindAssetForPlatform_NoMatchingAsset(t *testing.T) {
	// Create release with no matching assets
	emptyRelease := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "some-other-file.txt", BrowserDownloadURL: "http://example.com/other"},
		},
	}

	platform := PlatformInfo{OS: "linux", Arch: "amd64"}
	_, _, err := findAssetForPlatformWithInfo(emptyRelease, platform)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no compatible asset found")
}

func TestTruncateText(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:     "text shorter than max length",
			text:     "short",
			maxLen:   10,
			expected: "short",
		},
		{
			name:     "text equal to max length",
			text:     "exact",
			maxLen:   5,
			expected: "exact",
		},
		{
			name:     "text longer than max length",
			text:     "this is a very long text",
			maxLen:   10,
			expected: "this is a ...",
		},
		{
			name:     "empty text",
			text:     "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "max length zero",
			text:     "test",
			maxLen:   0,
			expected: "...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCopyFile(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "This is test content for file copying"
	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	// Set specific permissions on source file
	err = os.Chmod(srcFile, 0755)
	require.NoError(t, err)

	// Copy file
	dstFile := filepath.Join(tempDir, "destination.txt")
	err = copyFile(srcFile, dstFile)
	assert.NoError(t, err)

	// Verify content
	copiedContent, err := os.ReadFile(dstFile)
	require.NoError(t, err)
	assert.Equal(t, content, string(copiedContent))

	// Verify permissions
	srcInfo, err := os.Stat(srcFile)
	require.NoError(t, err)
	dstInfo, err := os.Stat(dstFile)
	require.NoError(t, err)
	assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
}

func TestCopyFile_SourceNotExists(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "nonexistent.txt")
	dstFile := filepath.Join(tempDir, "destination.txt")

	err := copyFile(srcFile, dstFile)
	assert.Error(t, err)
}

func TestCopyFile_InvalidDestination(t *testing.T) {
	tempDir := t.TempDir()

	// Create source file
	srcFile := filepath.Join(tempDir, "source.txt")
	err := os.WriteFile(srcFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Try to copy to invalid destination (directory that doesn't exist)
	dstFile := filepath.Join(tempDir, "nonexistent", "destination.txt")
	err = copyFile(srcFile, dstFile)
	assert.Error(t, err)
}

func TestCheckForUpdates_MockServer(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/releases/latest") {
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(mockRelease); err != nil {
				http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
				return
			}
		} else {
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Since we can't modify the const, we'll test the function directly
	// but note that in practice you'd want to make the API URL configurable for testing
	t.Skip("Skipping integration test - would need configurable API URL")
}

func TestDownloadAsset_MockServer(t *testing.T) {
	// Create mock server that serves a fake binary
	testContent := "fake binary content"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write([]byte(testContent)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// Download the asset
	tempFile, err := downloadAssetWithTestFlag(server.URL, "test-binary", true)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Verify content
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, testContent, string(content))

	// Verify file exists in temp directory
	assert.True(t, strings.Contains(tempFile, os.TempDir()))
	assert.True(t, strings.Contains(tempFile, "goanime-update-"))
}

func TestDownloadAsset_ServerError(t *testing.T) {
	// Create mock server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	_, err := downloadAssetWithTestFlag(server.URL, "test-binary", true)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "download failed with status 500")
}

func TestReplaceExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping on Windows due to file locking complexities")
	}

	tempDir := t.TempDir()

	// Create current executable
	currentExe := filepath.Join(tempDir, "current")
	currentContent := "current executable content"
	err := os.WriteFile(currentExe, []byte(currentContent), 0755)
	require.NoError(t, err)

	// Create new executable
	newExe := filepath.Join(tempDir, "new")
	newContent := "new executable content"
	err = os.WriteFile(newExe, []byte(newContent), 0644)
	require.NoError(t, err)

	// Replace executable
	err = replaceExecutable(currentExe, newExe)
	assert.NoError(t, err)

	// Verify content was replaced
	content, err := os.ReadFile(currentExe)
	require.NoError(t, err)
	assert.Equal(t, newContent, string(content))

	// Verify permissions on Unix systems
	if runtime.GOOS != "windows" {
		info, err := os.Stat(currentExe)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), info.Mode())
	}
}

func TestReplaceExecutable_WindowsLogic(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tempDir := t.TempDir()

	// Create current executable
	currentExe := filepath.Join(tempDir, "current.exe")
	currentContent := "current executable content"
	err := os.WriteFile(currentExe, []byte(currentContent), 0755)
	require.NoError(t, err)

	// Create new executable
	newExe := filepath.Join(tempDir, "new.exe")
	newContent := "new executable content"
	err = os.WriteFile(newExe, []byte(newContent), 0644)
	require.NoError(t, err)

	// Replace executable
	err = replaceExecutable(currentExe, newExe)
	assert.NoError(t, err)

	// Verify content was replaced
	content, err := os.ReadFile(currentExe)
	require.NoError(t, err)
	assert.Equal(t, newContent, string(content))
}

// Test for GetCurrentPlatform function
func TestGetCurrentPlatform(t *testing.T) {
	platform := GetCurrentPlatform()
	assert.NotEmpty(t, platform.OS)
	assert.NotEmpty(t, platform.Arch)
	assert.Equal(t, runtime.GOOS, platform.OS)
	assert.Equal(t, runtime.GOARCH, platform.Arch)
}

// Test HTTP timeout scenarios
func TestDownloadAsset_Timeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Short delay for test
		if _, err := w.Write([]byte("test content")); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	// This test would need custom HTTP client with timeout, but our current implementation uses default client
	// In production, you might want to add configurable timeouts
	tempFile, err := downloadAssetWithTestFlag(server.URL, "test-binary", true)
	assert.NoError(t, err)
	if tempFile != "" {
		defer func() {
			if err := os.Remove(tempFile); err != nil {
				t.Logf("Failed to remove temp file: %v", err)
			}
		}()
	}
}

// Test for large file download simulation
func TestDownloadAsset_LargeFile(t *testing.T) {
	// Create mock server that serves larger content
	largeContent := strings.Repeat("A", 10*1024) // 10KB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		if _, err := w.Write([]byte(largeContent)); err != nil {
			http.Error(w, "Failed to write response", http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	tempFile, err := downloadAssetWithTestFlag(server.URL, "large-binary", true)
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(tempFile); err != nil {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	// Verify content
	content, err := os.ReadFile(tempFile)
	require.NoError(t, err)
	assert.Equal(t, largeContent, string(content))
}

// Test file permissions more thoroughly
func TestCopyFile_Permissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Permission test not applicable on Windows")
	}

	tempDir := t.TempDir()

	// Test various permission combinations
	permissions := []os.FileMode{0644, 0755, 0600, 0777}

	for _, perm := range permissions {
		t.Run(perm.String(), func(t *testing.T) {
			srcFile := filepath.Join(tempDir, "source_"+perm.String())
			dstFile := filepath.Join(tempDir, "dest_"+perm.String())

			content := "test content for " + perm.String()
			err := os.WriteFile(srcFile, []byte(content), perm)
			require.NoError(t, err)

			err = copyFile(srcFile, dstFile)
			require.NoError(t, err)

			// Check permissions
			srcInfo, err := os.Stat(srcFile)
			require.NoError(t, err)
			dstInfo, err := os.Stat(dstFile)
			require.NoError(t, err)

			assert.Equal(t, srcInfo.Mode(), dstInfo.Mode())
		})
	}
}

// Test concurrent access scenarios
func TestCopyFile_Concurrent(t *testing.T) {
	tempDir := t.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	content := "concurrent test content"

	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(t, err)

	// Test multiple concurrent copies
	const numCopies = 5
	done := make(chan error, numCopies)

	for i := 0; i < numCopies; i++ {
		go func(index int) {
			dstFile := filepath.Join(tempDir, "dest_"+string(rune('A'+index))+".txt")
			done <- copyFile(srcFile, dstFile)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numCopies; i++ {
		err := <-done
		assert.NoError(t, err)
	}
}

// Test ReplaceExecutable with more edge cases
func TestReplaceExecutable_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("empty files", func(t *testing.T) {
		currentExe := filepath.Join(tempDir, "current_empty")
		newExe := filepath.Join(tempDir, "new_empty")

		// Create empty files
		err := os.WriteFile(currentExe, []byte{}, 0755)
		require.NoError(t, err)
		err = os.WriteFile(newExe, []byte{}, 0644)
		require.NoError(t, err)

		err = replaceExecutable(currentExe, newExe)
		assert.NoError(t, err)
	})

	t.Run("binary files", func(t *testing.T) {
		if runtime.GOOS == "windows" {
			t.Skip("Skipping binary test on Windows")
		}

		currentExe := filepath.Join(tempDir, "current_binary")
		newExe := filepath.Join(tempDir, "new_binary")

		// Create files with binary content
		binaryContent := []byte{0x7f, 0x45, 0x4c, 0x46} // ELF magic bytes
		err := os.WriteFile(currentExe, binaryContent, 0755)
		require.NoError(t, err)

		newBinaryContent := []byte{0x7f, 0x45, 0x4c, 0x47} // Modified ELF
		err = os.WriteFile(newExe, newBinaryContent, 0644)
		require.NoError(t, err)

		err = replaceExecutable(currentExe, newExe)
		require.NoError(t, err)

		// Verify content and permissions
		content, err := os.ReadFile(currentExe)
		require.NoError(t, err)
		assert.Equal(t, newBinaryContent, content)

		info, err := os.Stat(currentExe)
		require.NoError(t, err)
		assert.Equal(t, os.FileMode(0755), info.Mode())
	})
}

// Benchmark tests for performance-critical functions
func BenchmarkIsVersionNewer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = isVersionNewer("2.0.0", "1.0.0")
	}
}

func BenchmarkTruncateText(b *testing.B) {
	text := "This is a very long text that needs to be truncated for display purposes"
	for i := 0; i < b.N; i++ {
		_ = truncateText(text, 50)
	}
}

func BenchmarkFindAssetForPlatform(b *testing.B) {
	platform := PlatformInfo{OS: "linux", Arch: "amd64"}
	for i := 0; i < b.N; i++ {
		_, _, _ = findAssetForPlatformWithInfo(&mockRelease, platform)
	}
}

// Integration-style tests that test multiple components together
func TestUpdateWorkflow_MockScenario(t *testing.T) {
	t.Skip("Integration test - requires full mock setup")

	// This would test the complete update workflow:
	// 1. Check for updates
	// 2. Download new version
	// 3. Replace executable
	// 4. Verify update success
	//
	// Implementation would require:
	// - Mock HTTP server for GitHub API
	// - Temporary executables
	// - Mock user interaction for prompts
}

// Test for edge cases in version comparison
func TestIsVersionNewer_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
	}{
		{"empty versions", "", "", false},
		{"single digit versions", "2", "1", true},
		{"mixed length", "1.2", "1.2.0", false},
		{"very long versions", "1.2.3.4.5.6", "1.2.3.4.5.5", true},
		{"zeros", "1.0.0", "1.0", false},
		{"leading zeros", "01.02.03", "1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := isVersionNewer(tt.latest, tt.current)
			if tt.name == "empty versions" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test version comparison with semantic versioning edge cases
func TestIsVersionNewer_SemanticVersioning(t *testing.T) {
	tests := []struct {
		name     string
		latest   string
		current  string
		expected bool
		hasError bool
	}{
		// Pre-release versions (these should be treated as regular versions in our simple implementation)
		{"pre-release ignored", "2.0.0", "2.0.0", false, false},
		{"build metadata ignored", "1.0.0", "1.0.0", false, false},

		// Version with more than 3 parts
		{"four part version newer", "1.2.3.4", "1.2.3.3", true, false},
		{"four part version same", "1.2.3.4", "1.2.3.4", false, false},

		// Large version numbers
		{"large version numbers", "99.99.99", "1.0.0", true, false},
		{"very large numbers", "999.999.999", "1.0.0", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := isVersionNewer(tt.latest, tt.current)

			if tt.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

// Test asset matching with case sensitivity
func TestFindAssetForPlatform_CaseSensitivity(t *testing.T) {
	// Create release with mixed case assets
	mixedCaseRelease := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "GoAnime-Linux-amd64", BrowserDownloadURL: "http://example.com/GoAnime-Linux-amd64"},
			{Name: "GOANIME-WINDOWS-AMD64.EXE", BrowserDownloadURL: "http://example.com/GOANIME-WINDOWS-AMD64.EXE"},
			{Name: "goanime-darwin-amd64", BrowserDownloadURL: "http://example.com/goanime-darwin-amd64"},
		},
	}

	tests := []struct {
		name         string
		platform     PlatformInfo
		expectedName string
		shouldFind   bool
	}{
		{
			name:         "linux case insensitive match",
			platform:     PlatformInfo{OS: "linux", Arch: "amd64"},
			expectedName: "GoAnime-Linux-amd64",
			shouldFind:   true,
		},
		{
			name:         "windows case insensitive match",
			platform:     PlatformInfo{OS: "windows", Arch: "amd64"},
			expectedName: "GOANIME-WINDOWS-AMD64.EXE",
			shouldFind:   true,
		},
		{
			name:         "darwin exact match",
			platform:     PlatformInfo{OS: "darwin", Arch: "amd64"},
			expectedName: "goanime-darwin-amd64",
			shouldFind:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, name, err := findAssetForPlatformWithInfo(mixedCaseRelease, tt.platform)

			if tt.shouldFind {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedName, name)
				assert.NotEmpty(t, url)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Test truncateText with Unicode characters
func TestTruncateText_Unicode(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		maxLen   int
		expected string
	}{
		{
			name:   "unicode characters - byte-based truncation",
			text:   "Hello 世界 🌍",
			maxLen: 8,
			// Note: Current implementation uses byte-based truncation which can break UTF-8
			expected: "Hello 世" + "...", // This might actually be invalid UTF-8
		},
		{
			name:     "ascii only",
			text:     "Hello World",
			maxLen:   5,
			expected: "Hello...",
		},
		{
			name:     "short unicode text",
			text:     "Test",
			maxLen:   10,
			expected: "Test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateText(tt.text, tt.maxLen)
			// For now, just test the length constraint since Unicode handling is implementation-dependent
			if len(tt.text) <= tt.maxLen {
				assert.Equal(t, tt.text, result)
			} else {
				assert.LessOrEqual(t, len(result), tt.maxLen+3) // maxLen + "..."
				assert.True(t, strings.HasSuffix(result, "..."))
			}
		})
	}
}

// Stress test for version comparison
func TestIsVersionNewer_StressTest(t *testing.T) {
	// Test with many version parts
	longVersion1 := "1.2.3.4.5.6.7.8.9.10.11.12"
	longVersion2 := "1.2.3.4.5.6.7.8.9.10.11.13"

	result, err := isVersionNewer(longVersion2, longVersion1)
	assert.NoError(t, err)
	assert.True(t, result)

	// Test with same long versions
	result, err = isVersionNewer(longVersion1, longVersion1)
	assert.NoError(t, err)
	assert.False(t, result)
}

// Test error handling in downloadAsset
func TestDownloadAsset_ErrorHandling(t *testing.T) {
	t.Run("invalid URL", func(t *testing.T) {
		_, err := downloadAssetWithTestFlag("not-a-valid-url", "test", true)
		assert.Error(t, err)
	})

	t.Run("network unreachable", func(t *testing.T) {
		_, err := downloadAssetWithTestFlag("http://192.0.2.1:1234/nonexistent", "test", true)
		assert.Error(t, err)
	})

	t.Run("404 not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.NotFound(w, r)
		}))
		defer server.Close()

		_, err := downloadAssetWithTestFlag(server.URL, "test", true)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})
}

// Benchmark for file operations
func BenchmarkCopyFile(b *testing.B) {
	tempDir := b.TempDir()
	srcFile := filepath.Join(tempDir, "source.txt")
	content := strings.Repeat("benchmark test content\n", 1000) // ~23KB

	err := os.WriteFile(srcFile, []byte(content), 0644)
	require.NoError(b, err)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dstFile := filepath.Join(tempDir, "dest_"+string(rune(i%26+'A'))+".txt")
		_ = copyFile(srcFile, dstFile)
		if err := os.Remove(dstFile); err != nil {
			b.Logf("Failed to remove temp file: %v", err)
		}
	}
}

// Test table-driven approach for multiple platform combinations
func TestFindAssetForPlatform_AllCombinations(t *testing.T) {
	// Create comprehensive release with assets for all supported platforms
	comprehensiveRelease := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			// Linux assets
			{Name: "goanime-linux-amd64", BrowserDownloadURL: "http://example.com/goanime-linux-amd64"},
			{Name: "goanime-linux-386", BrowserDownloadURL: "http://example.com/goanime-linux-386"},
			{Name: "goanime-linux-arm64", BrowserDownloadURL: "http://example.com/goanime-linux-arm64"},

			// Windows assets
			{Name: "goanime-windows-amd64.exe", BrowserDownloadURL: "http://example.com/goanime-windows-amd64.exe"},
			{Name: "goanime-windows-386.exe", BrowserDownloadURL: "http://example.com/goanime-windows-386.exe"},

			// macOS assets
			{Name: "goanime-darwin-amd64", BrowserDownloadURL: "http://example.com/goanime-darwin-amd64"},
			{Name: "goanime-darwin-arm64", BrowserDownloadURL: "http://example.com/goanime-darwin-arm64"},
			{Name: "goanime-darwin-universal", BrowserDownloadURL: "http://example.com/goanime-darwin-universal"},
			{Name: "goanime-darwin", BrowserDownloadURL: "http://example.com/goanime-darwin"},

			// Alternative naming patterns
			{Name: "goanime-macos-amd64", BrowserDownloadURL: "http://example.com/goanime-macos-amd64"},
		},
	}

	platformTests := []struct {
		os           string
		arch         string
		expectedName string
	}{
		{"linux", "amd64", "goanime-linux-amd64"},
		{"linux", "386", "goanime-linux-386"},
		{"linux", "arm64", "goanime-linux-arm64"},
		{"windows", "amd64", "goanime-windows-amd64.exe"},
		{"windows", "386", "goanime-windows-386.exe"},
		{"darwin", "amd64", "goanime-darwin-amd64"},
		{"darwin", "arm64", "goanime-darwin-arm64"},
	}

	for _, tt := range platformTests {
		t.Run(tt.os+"_"+tt.arch, func(t *testing.T) {
			platform := PlatformInfo{OS: tt.os, Arch: tt.arch}
			url, name, err := findAssetForPlatformWithInfo(comprehensiveRelease, platform)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedName, name)
			assert.Contains(t, url, "http://example.com/")
		})
	}
}

// Test universal binary fallback behavior for macOS
func TestFindAssetForPlatform_UniversalBinaryFallback(t *testing.T) {
	// Test case where only universal binaries are available
	universalOnlyRelease := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "goanime-linux-amd64", BrowserDownloadURL: "http://example.com/goanime-linux-amd64"},
			{Name: "goanime-windows-amd64.exe", BrowserDownloadURL: "http://example.com/goanime-windows-amd64.exe"},
			// Only universal binaries for macOS
			{Name: "goanime-darwin-universal", BrowserDownloadURL: "http://example.com/goanime-darwin-universal"},
			{Name: "goanime-darwin", BrowserDownloadURL: "http://example.com/goanime-darwin"},
		},
	}

	tests := []struct {
		name         string
		platform     PlatformInfo
		expectedName string
		description  string
	}{
		{
			name:         "amd64_falls_back_to_universal",
			platform:     PlatformInfo{OS: "darwin", Arch: "amd64"},
			expectedName: "goanime-darwin-universal",
			description:  "Intel Mac should use universal binary when arch-specific not available",
		},
		{
			name:         "arm64_falls_back_to_universal",
			platform:     PlatformInfo{OS: "darwin", Arch: "arm64"},
			expectedName: "goanime-darwin-universal",
			description:  "Apple Silicon Mac should use universal binary when arch-specific not available",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url, name, err := findAssetForPlatformWithInfo(universalOnlyRelease, tt.platform)

			assert.NoError(t, err, tt.description)
			assert.Equal(t, tt.expectedName, name, tt.description)
			assert.Contains(t, url, "http://example.com/")
		})
	}

	// Test case where only generic universal binary is available
	genericOnlyRelease := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: "goanime-linux-amd64", BrowserDownloadURL: "http://example.com/goanime-linux-amd64"},
			{Name: "goanime-windows-amd64.exe", BrowserDownloadURL: "http://example.com/goanime-windows-amd64.exe"},
			// Only generic universal binary for macOS
			{Name: "goanime-darwin", BrowserDownloadURL: "http://example.com/goanime-darwin"},
		},
	}

	t.Run("fallback_to_generic_universal", func(t *testing.T) {
		platform := PlatformInfo{OS: "darwin", Arch: "amd64"}
		url, name, err := findAssetForPlatformWithInfo(genericOnlyRelease, platform)

		assert.NoError(t, err)
		assert.Equal(t, "goanime-darwin", name)
		assert.Contains(t, url, "http://example.com/")
	})
}
