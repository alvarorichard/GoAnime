//go:build !windows
// +build !windows

package ipc

import (
	"fmt"
	"net"
	"time"
)

// UnixSocket implements SocketInterface for Unix
type UnixSocket struct {
	conn net.Conn
}

func NewSocket() *UnixSocket {
	return &UnixSocket{}
}

func (u *UnixSocket) Open() error {
	path := GetIpcPath() + "/discord-ipc-0"
	sock, err := net.DialTimeout("unix", path, 2*time.Second)
	if err != nil {
		return err
	}
	u.conn = sock
	return nil
}

func (u *UnixSocket) Close() error {
	if u.conn != nil {
		err := u.conn.Close()
		u.conn = nil
		return err
	}
	return nil
}

func (u *UnixSocket) Send(opcode int, payload string) (string, error) {
	if u.conn == nil {
		return "", fmt.Errorf("socket not connected")
	}
	return sendToConn(u, opcode, payload)
}

// ConnInterface methods
func (u *UnixSocket) WriteBytes(data []byte) error {
	_, err := u.conn.Write(data)
	return err
}

func (u *UnixSocket) ReadBytes() ([]byte, error) {
	buf := make([]byte, 512)
	n, err := u.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
