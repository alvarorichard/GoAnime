package player

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadMPVLogTail(t *testing.T) {
	assert.Empty(t, readMPVLogTail(""))
	assert.Empty(t, readMPVLogTail(filepath.Join(t.TempDir(), "missing.log")))

	var b strings.Builder
	for i := 1; i <= 40; i++ {
		b.WriteString("line\n")
	}
	b.WriteString("[vo] Error: no usable video output\n")
	p := filepath.Join(t.TempDir(), "mpv.log")
	require.NoError(t, os.WriteFile(p, []byte(b.String()), 0o600))

	tail := readMPVLogTail(p)
	assert.Len(t, strings.Split(tail, "\n"), 15)
	assert.Contains(t, tail, "no usable video output")
}

func TestJoinMPVDiagnostics(t *testing.T) {
	assert.Equal(t, "", joinMPVDiagnostics("", ""))
	assert.Equal(t, "log", joinMPVDiagnostics("", "log"))
	assert.Equal(t, "err\nlog", joinMPVDiagnostics("err\n", "log"))
}
