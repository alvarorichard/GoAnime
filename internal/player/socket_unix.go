//go:build !windows

// Arquivo específico para sistemas Unix (Linux, macOS) que implementa
// a conexão com o socket do MPV utilizando sockets Unix padrão.

package player

import (
	"net"
)

// dialMPVSocket cria uma conexão com o socket do MPV em sistemas Unix.
// Em sistemas Unix (Linux, macOS), utilizamos sockets Unix padrão
// através da função net.Dial com o tipo "unix".
func dialMPVSocket(socketPath string) (net.Conn, error) {
	// Unix-like system uses Unix sockets
	return net.Dial("unix", socketPath)
}
