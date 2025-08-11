package discord

import (
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/rich-go/client"
)

// DiscordClientID é o ID do cliente Discord para Rich Presence
const DiscordClientID = "1302721937717334128"

// Manager manages the global Discord Rich Presence state
type Manager struct {
	isEnabled     bool
	clientID      string
	initTime      time.Time
	isInitialized bool
}

// NewManager creates a new instance of the Discord manager
func NewManager() *Manager {
	return &Manager{
		isEnabled:     false,
		clientID:      DiscordClientID,
		isInitialized: false,
	}
}

// Initialize inicializa o Discord Rich Presence
func (m *Manager) Initialize() error {
	if m.isInitialized {
		util.Debug("Discord Rich Presence already initialized")
		return nil
	}

	m.initTime = time.Now()

	if err := client.Login(m.clientID); err != nil {
		// Discord is optional; only log when debug is enabled
		if util.IsDebug {
			util.Debugf("Discord Rich Presence not available: %v", err)
		}
		m.isEnabled = false
		return err
	}

	m.isEnabled = true
	m.isInitialized = true

	util.Debugf("[PERF] Discord Rich Presence initialized in %v", time.Since(m.initTime))

	return nil
}

// Shutdown desliga o Discord Rich Presence
func (m *Manager) Shutdown() {
	if m.isInitialized && m.isEnabled {
		client.Logout()
		m.isEnabled = false
		m.isInitialized = false
		util.Debug("Discord Rich Presence shutdown completed")
	}
}

// IsEnabled returns whether Discord Rich Presence is enabled
func (m *Manager) IsEnabled() bool {
	return m.isEnabled
}

// IsInitialized returns whether Discord Rich Presence was initialized
func (m *Manager) IsInitialized() bool {
	return m.isInitialized
}

// GetClientID returns the Discord client ID
func (m *Manager) GetClientID() string {
	return m.clientID
}

// SetClientID sets a new Discord client ID (must be called before Initialize)
func (m *Manager) SetClientID(clientID string) {
	if !m.isInitialized {
		m.clientID = clientID
	} else {
		util.Debug("Cannot change client ID after initialization")
	}
}

// GetInitializationTime returns the initialization time
func (m *Manager) GetInitializationTime() time.Duration {
	if m.isInitialized {
		return time.Since(m.initTime)
	}
	return 0
}
