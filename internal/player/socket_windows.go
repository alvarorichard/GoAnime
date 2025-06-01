//go:build windows

// Arquivo específico para Windows que implementa a conexão com o socket do MPV
// utilizando named pipes ao invés de sockets Unix.

package player

import (
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// dialMPVSocket cria uma conexão com o socket do MPV no Windows.
// No Windows, utilizamos named pipes no formato \\.\pipe\NOME_DO_PIPE
// O pacote go-winio é usado apenas em Windows para suporte a named pipes.
func dialMPVSocket(socketPath string) (net.Conn, error) {
	// Windows uses named pipes format
	// Named pipes in Windows need to be in the format \\.\pipe\PIPENAME
	if !strings.HasPrefix(socketPath, `\\.\pipe\`) {
		socketPath = `\\.\pipe\` + filepath.Base(socketPath)
	}

	// Use winio to connect to Windows named pipe
	timeout := 5 * time.Second
	return winio.DialPipe(socketPath, &timeout)
}
