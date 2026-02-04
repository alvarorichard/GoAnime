package discord

import (
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Manager manages the global Discord Rich Presence state
type Manager struct {
	isEnabled     bool
	clientID      string
	initTime      time.Time
	isInitialized bool
	initMutex     sync.RWMutex
}

// NewManager creates a new instance of the Discord manager
func NewManager() *Manager {
	return &Manager{
		isEnabled:     false,
		clientID:      DiscordClientID,
		isInitialized: false,
	}
}

// Initialize initializes the Discord Rich Presence (non-blocking, runs in background)
func (m *Manager) Initialize() error {
	m.initMutex.Lock()
	if m.isInitialized {
		m.initMutex.Unlock()
		util.Debug("Discord Rich Presence already initialized")
		return nil
	}
	m.initMutex.Unlock()

	m.initTime = time.Now()

	// Run Discord init in background so it doesn't block app startup
	go func() {
		if err := LoginClient(); err != nil {
			if util.IsDebug {
				util.Debugf("Discord Rich Presence not available: %v", err)
			}
			m.initMutex.Lock()
			m.isEnabled = false
			m.isInitialized = true
			m.initMutex.Unlock()
			return
		}

		m.initMutex.Lock()
		m.isEnabled = true
		m.isInitialized = true
		m.initMutex.Unlock()

		util.Debugf("[PERF] Discord Rich Presence initialized in %v", time.Since(m.initTime))
	}()

	return nil
}

// Shutdown shuts down the Discord Rich Presence
func (m *Manager) Shutdown() {
	m.initMutex.Lock()
	defer m.initMutex.Unlock()

	if m.isEnabled {
		_ = LogoutClient()
		m.isEnabled = false
		util.Debug("Discord Rich Presence shutdown completed")
	}
}

// IsEnabled returns whether Discord Rich Presence is enabled
func (m *Manager) IsEnabled() bool {
	m.initMutex.RLock()
	defer m.initMutex.RUnlock()
	return m.isEnabled
}

// IsInitialized returns whether Discord Rich Presence was initialized
func (m *Manager) IsInitialized() bool {
	m.initMutex.RLock()
	defer m.initMutex.RUnlock()
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
