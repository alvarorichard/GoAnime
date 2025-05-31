package discord

import (
	"log"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/rich-go/client"
)

// DiscordClientID é o ID do cliente Discord para Rich Presence
const DiscordClientID = "1302721937717334128"

// Manager gerencia o estado global do Discord Rich Presence
type Manager struct {
	isEnabled     bool
	clientID      string
	initTime      time.Time
	isInitialized bool
}

// NewManager cria uma nova instância do gerenciador Discord
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
		if util.IsDebug {
			log.Println("Discord Rich Presence already initialized")
		}
		return nil
	}

	m.initTime = time.Now()

	if err := client.Login(m.clientID); err != nil {
		if util.IsDebug {
			log.Printf("Failed to initialize Discord Rich Presence: %v", err)
		}
		m.isEnabled = false
		return err
	}

	m.isEnabled = true
	m.isInitialized = true

	if util.IsDebug {
		log.Printf("[PERF] Discord Rich Presence initialized in %v", time.Since(m.initTime))
	}

	return nil
}

// Shutdown desliga o Discord Rich Presence
func (m *Manager) Shutdown() {
	if m.isInitialized && m.isEnabled {
		client.Logout()
		m.isEnabled = false
		m.isInitialized = false
		if util.IsDebug {
			log.Println("Discord Rich Presence shutdown completed")
		}
	}
}

// IsEnabled retorna se o Discord Rich Presence está habilitado
func (m *Manager) IsEnabled() bool {
	return m.isEnabled
}

// IsInitialized retorna se o Discord Rich Presence foi inicializado
func (m *Manager) IsInitialized() bool {
	return m.isInitialized
}

// GetClientID retorna o ID do cliente Discord
func (m *Manager) GetClientID() string {
	return m.clientID
}

// SetClientID define um novo ID do cliente Discord (deve ser chamado antes de Initialize)
func (m *Manager) SetClientID(clientID string) {
	if !m.isInitialized {
		m.clientID = clientID
	} else if util.IsDebug {
		log.Println("Cannot change client ID after initialization")
	}
}

// GetInitializationTime retorna o tempo de inicialização
func (m *Manager) GetInitializationTime() time.Duration {
	if m.isInitialized {
		return time.Since(m.initTime)
	}
	return 0
}
