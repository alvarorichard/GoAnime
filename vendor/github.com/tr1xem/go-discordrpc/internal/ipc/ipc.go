package ipc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

type SocketInterface interface {
	Open() error
	Close() error
	Send(opcode int, payload string) (string, error)
}

// GetIpcPath returns the best IPC socket path for the current environment.
func GetIpcPath() string {
	candidates := []string{
		"/run/user/1000/snap.discord",
		"/run/user/1000/.flatpak/com.discordapp.Discord/xdg-run",
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	envVars := []string{"XDG_RUNTIME_DIR", "TMPDIR", "TMP", "TEMP"}
	for _, v := range envVars {
		if path, exists := os.LookupEnv(v); exists {
			return path
		}
	}

	return "/tmp"
}

// sendToConn is a helper for both UnixSocket and WindowsSocket
func sendToConn(conn ConnInterface, opcode int, payload string) (string, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.LittleEndian, int32(opcode)); err != nil {
		return "", fmt.Errorf("failed to write opcode: %w", err)
	}
	if err := binary.Write(buf, binary.LittleEndian, int32(len(payload))); err != nil {
		return "", fmt.Errorf("failed to write payload length: %w", err)
	}
	if _, err := buf.Write([]byte(payload)); err != nil {
		return "", fmt.Errorf("failed to write payload: %w", err)
	}

	if err := conn.WriteBytes(buf.Bytes()); err != nil {
		return "", fmt.Errorf("failed to send to socket: %w", err)
	}

	data, err := conn.ReadBytes()
	if err != nil {
		return "", fmt.Errorf("failed to read from socket: %w", err)
	}
	if len(data) <= 8 {
		return "", nil
	}
	return string(data[8:]), nil
}

// DefaultSocket returns a new platform-specific socket instance
func DefaultSocket() SocketInterface {
	return NewSocket()
}

// ConnInterface abstracts net.Conn / npipe.Conn for platform-independent reading/writing
type ConnInterface interface {
	WriteBytes([]byte) error
	ReadBytes() ([]byte, error)
}
