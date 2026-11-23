package client

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"

	"github.com/tr1xem/go-discordrpc/internal/ipc"
)

// Client represents a single Discord RPC connection
type Client struct {
	ClientID string
	loggedIn bool
	socket   ipc.SocketInterface
}

// NewClient creates a new Discord RPC client instance
func NewClient(clientID string) *Client {
	var sock ipc.SocketInterface

	sock = ipc.DefaultSocket()

	return &Client{
		ClientID: clientID,
		socket:   sock,
	}
}

// Login performs a handshake with Discord and opens the IPC socket.
func (c *Client) Login() error {
	if c.loggedIn {
		return nil
	}

	payload, err := json.Marshal(Handshake{"1", c.ClientID})
	if err != nil {
		return fmt.Errorf("failed to marshal handshake: %w", err)
	}

	if err := c.socket.Open(); err != nil {
		return fmt.Errorf("failed to open IPC socket: %w", err)
	}

	// Send handshake and handle response
	response, err := c.socket.Send(0, string(payload))
	if err != nil {
		return fmt.Errorf("failed to send handshake: %w", err)
	}
	if len(response) == 0 {
		return fmt.Errorf("empty response from Discord handshake")
	}

	c.loggedIn = true
	return nil
}

// Logout closes the IPC socket and marks the client as logged out.
func (c *Client) Logout() error {
	if !c.loggedIn {
		return nil
	}

	if err := c.socket.Close(); err != nil {
		return fmt.Errorf("failed to close IPC socket: %w", err)
	}

	c.loggedIn = false
	return nil
}

// SetActivity updates the Discord Rich Presence activity.
func (c *Client) SetActivity(activity Activity) error {
	if !c.loggedIn {
		return fmt.Errorf("client is not logged in")
	}

	nonce, err := generateNonce()
	if err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	payload, err := json.Marshal(Frame{
		"SET_ACTIVITY",
		Args{
			os.Getpid(),
			mapActivity(&activity),
		},
		nonce,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal activity frame: %w", err)
	}

	// Send activity frame and handle response
	response, err := c.socket.Send(1, string(payload))
	if err != nil {
		return fmt.Errorf("failed to send activity frame: %w", err)
	}
	if len(response) == 0 {
		return fmt.Errorf("empty response from Discord activity")
	}

	return nil
}

// generateNonce creates a unique nonce for Discord RPC requests.
func generateNonce() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	// set version bits for UUIDv4
	buf[6] = (buf[6] & 0x0f) | 0x40

	return fmt.Sprintf("%x-%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:]), nil
}
