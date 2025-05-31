package test

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMPVClient é um mock para o cliente MPV
type MockMPVClient struct {
	mock.Mock
}

// MockMPVSendCommand é uma implementação mock de mpvSendCommand
func (m *MockMPVClient) MockMPVSendCommand(socketPath string, args []interface{}) (interface{}, error) {
	callArgs := m.Called(socketPath, args)
	return callArgs.Get(0), callArgs.Error(1)
}

// mockMPVSendCommand é uma função mock simples
func mockMPVSendCommand(socketPath string, args []interface{}) (interface{}, error) {
	// Simula a resposta do MPV para time-pos
	if len(args) >= 2 && args[0] == "get_property" && args[1] == "time-pos" {
		return 630.0, nil // 10 minutos e 30 segundos em segundos
	}
	return nil, fmt.Errorf("mock: unsupported command")
}

// TestNewRichPresenceUpdater testa a função construtora
func TestNewRichPresenceUpdater(t *testing.T) {
	// Arrange
	anime := &models.Anime{
		Name:     "Test Anime",
		URL:      "https://example.com/anime/test",
		ImageURL: "https://example.com/image.jpg",
		Episodes: []models.Episode{
			{Number: "1"},
		},
		AnilistID: 123,
		MalID:     456,
		Details: models.AniListDetails{
			Title: models.Title{
				Romaji:  "Test Anime",
				English: "Test Anime English",
			},
		},
	}
	isPaused := false
	animeMutex := &sync.Mutex{}
	updateFreq := 5 * time.Second
	episodeDuration := 24 * time.Minute
	socketPath := "/tmp/mpvsocket"

	// Act
	updater := discord.NewRichPresenceUpdater(anime, &isPaused, animeMutex, updateFreq, episodeDuration, socketPath, mockMPVSendCommand)

	// Assert
	assert.NotNil(t, updater, "RichPresenceUpdater should not be nil")
}

// TestGetCurrentPlaybackPosition testa a recuperação da posição de reprodução
func TestGetCurrentPlaybackPosition(t *testing.T) {
	t.Run("Should return current playback position", func(t *testing.T) {
		// Arrange
		anime := &models.Anime{
			Name:     "Test Anime",
			Episodes: []models.Episode{{Number: "1"}},
			Details: models.AniListDetails{
				Title: models.Title{Romaji: "Test Anime"},
			},
		}
		isPaused := false
		animeMutex := &sync.Mutex{}
		updateFreq := 5 * time.Second
		episodeDuration := 24 * time.Minute
		socketPath := "/tmp/mpvsocket"

		updater := discord.NewRichPresenceUpdater(anime, &isPaused, animeMutex, updateFreq, episodeDuration, socketPath, mockMPVSendCommand)

		// Act
		position, err := updater.GetCurrentPlaybackPosition()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 10*time.Minute+30*time.Second, position)
		assert.Equal(t, 10.5, position.Minutes())
		assert.Equal(t, 630.0, position.Seconds())
	})

	t.Run("Should handle MPV send command error", func(t *testing.T) {
		// Arrange
		errorMockFunc := func(socketPath string, args []interface{}) (interface{}, error) {
			return nil, fmt.Errorf("connection failed")
		}

		anime := &models.Anime{
			Name:     "Test Anime",
			Episodes: []models.Episode{{Number: "1"}},
			Details: models.AniListDetails{
				Title: models.Title{Romaji: "Test Anime"},
			},
		}
		isPaused := false
		animeMutex := &sync.Mutex{}
		updateFreq := 5 * time.Second
		episodeDuration := 24 * time.Minute
		socketPath := "/tmp/mpvsocket"

		updater := discord.NewRichPresenceUpdater(anime, &isPaused, animeMutex, updateFreq, episodeDuration, socketPath, errorMockFunc)

		// Act
		position, err := updater.GetCurrentPlaybackPosition()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, time.Duration(0), position)
	})
}

// TestStartAndStop testa a funcionalidade de start e stop
func TestStartAndStop(t *testing.T) {
	// Arrange
	anime := &models.Anime{
		Name:     "Test Anime",
		URL:      "https://example.com/anime/test",
		ImageURL: "https://example.com/image.jpg",
		Episodes: []models.Episode{
			{Number: "1"},
		},
		AnilistID: 123,
		MalID:     456,
		Details: models.AniListDetails{
			Title: models.Title{
				Romaji:  "Test Anime",
				English: "Test Anime English",
			},
		},
	}
	isPaused := false
	animeMutex := &sync.Mutex{}
	updateFreq := 100 * time.Millisecond // Intervalo curto para teste
	episodeDuration := 24 * time.Minute
	socketPath := "/tmp/mpvsocket"

	updater := discord.NewRichPresenceUpdater(anime, &isPaused, animeMutex, updateFreq, episodeDuration, socketPath, mockMPVSendCommand)

	t.Run("Start and Stop should complete without errors", func(t *testing.T) {
		// Usamos um timeout para garantir que o teste não trave
		completed := make(chan bool)

		go func() {
			// Inicia o updater em uma goroutine
			updater.Start()

			// Deixa rodar por um tempo curto
			time.Sleep(200 * time.Millisecond)

			// Para
			updater.Stop()

			// Sinaliza conclusão
			completed <- true
		}()

		// Aguarda conclusão ou timeout
		select {
		case <-completed:
			// Teste passou se chegamos aqui
		case <-time.After(1 * time.Second):
			t.Fatal("Test timed out - Stop() may be blocking indefinitely")
		}
	})
}

// TestUpdateDiscordPresence testa a funcionalidade de atualização do Discord presence
func TestUpdateDiscordPresence(t *testing.T) {
	t.Run("Should format time correctly", func(t *testing.T) {
		// Testa a lógica de formatação de tempo usada em updateDiscordPresence
		currentPosition := 5*time.Minute + 30*time.Second
		episodeDuration := 24*time.Minute + 15*time.Second

		// Converte a duração do episódio para formato de minutos e segundos
		totalMinutes := int(episodeDuration.Minutes())
		totalSeconds := int(episodeDuration.Seconds()) % 60

		// Formata a posição atual de reprodução como minutos e segundos
		timeInfo := formatTime(int(currentPosition.Minutes()), int(currentPosition.Seconds())%60,
			totalMinutes, totalSeconds)

		// Asserta que o formato de tempo está correto
		assert.Equal(t, "05:30 / 24:15", timeInfo)
	})
}

// formatTime é uma função auxiliar para formatar tempo similar ao updateDiscordPresence
func formatTime(currentMinutes, currentSeconds, totalMinutes, totalSeconds int) string {
	return fmt.Sprintf("%02d:%02d / %02d:%02d",
		currentMinutes, currentSeconds,
		totalMinutes, totalSeconds,
	)
}

// TestDiscordManager testa o gerenciador Discord
func TestDiscordManager(t *testing.T) {
	t.Run("Should create new manager", func(t *testing.T) {
		manager := discord.NewManager()
		assert.NotNil(t, manager)
		assert.False(t, manager.IsEnabled())
		assert.False(t, manager.IsInitialized())
		assert.Equal(t, discord.DiscordClientID, manager.GetClientID())
	})

	t.Run("Should set client ID before initialization", func(t *testing.T) {
		manager := discord.NewManager()
		newClientID := "test-client-id"
		manager.SetClientID(newClientID)
		assert.Equal(t, newClientID, manager.GetClientID())
	})

	t.Run("Should handle shutdown gracefully", func(t *testing.T) {
		manager := discord.NewManager()
		// Deve ser seguro chamar Shutdown mesmo sem inicializar
		manager.Shutdown()
		assert.False(t, manager.IsEnabled())
		assert.False(t, manager.IsInitialized())
	})
}

// TestRichPresenceIntegration testa a integração geral do RichPresenceUpdater
func TestRichPresenceIntegration(t *testing.T) {
	t.Skip("Skipping integration test as it requires Discord and MPV to be running")

	// Em uma aplicação real, você implementaria um teste de integração mais abrangente
	// que usa instâncias reais ou mock dos sistemas externos
}
