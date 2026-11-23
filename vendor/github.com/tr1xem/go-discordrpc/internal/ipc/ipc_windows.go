//go:build windows
// +build windows

package ipc

import (
	"fmt"
	"time"

	npipe "gopkg.in/natefinch/npipe.v2"
)

// WindowsSocket implements SocketInterface for Windows
type WindowsSocket struct {
	conn *npipe.PipeConn
}

func NewSocket() *WindowsSocket {
	return &WindowsSocket{}
}

func (w *WindowsSocket) Open() error {
	const pipeName = `\\.\pipe\discord-ipc-0`
	const timeout = 2 * time.Second
	sock, err := npipe.DialTimeout(pipeName, timeout)
	if err != nil {
		return err
	}
	w.conn = sock
	return nil
}

func (w *WindowsSocket) Close() error {
	if w.conn != nil {
		err := w.conn.Close()
		w.conn = nil
		return err
	}
	return nil
}

func (w *WindowsSocket) Send(opcode int, payload string) (string, error) {
	if w.conn == nil {
		return "", fmt.Errorf("socket not connected")
	}
	return sendToConn(w, opcode, payload)
}

// ConnInterface methods
func (w *WindowsSocket) WriteBytes(data []byte) error {
	_, err := w.conn.Write(data)
	return err
}

func (w *WindowsSocket) ReadBytes() ([]byte, error) {
	buf := make([]byte, 512)
	n, err := w.conn.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:n], nil
}
