package player

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Style definitions
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00BFFF")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true)

	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF69B4")).
			Bold(true)

	commandStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98FB98")).
			Padding(0, 1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#87CEEB")).
			Bold(true).
			Underline(true)

	// New enhanced styles
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")).
			Bold(true)

	actionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#626262")).
			Padding(1, 2)
)

// applySkipTimes aplica os tempos de skip em uma instância do mpv
func applySkipTimes(socketPath string, episode *models.Episode) {
	var opts []string
	if episode.SkipTimes.Op.Start > 0 || episode.SkipTimes.Op.End > 0 {
		opts = append(opts, fmt.Sprintf("skip_op=%d-%d", episode.SkipTimes.Op.Start, episode.SkipTimes.Op.End))
	}
	if episode.SkipTimes.Ed.Start > 0 || episode.SkipTimes.Ed.End > 0 {
		opts = append(opts, fmt.Sprintf("skip_ed=%d-%d", episode.SkipTimes.Ed.Start, episode.SkipTimes.Ed.End))
	}

	if len(opts) > 0 {
		combinedOpts := strings.Join(opts, ",")
		_, cmdErr := mpvSendCommand(socketPath, []interface{}{"set_property", "script-opts", combinedOpts})
		if cmdErr != nil {
			if util.IsDebug {
				log.Printf("Falha ao aplicar skip times: %v. Comando: set_property script-opts %s", cmdErr, combinedOpts)
			}
		} else if util.IsDebug {
			log.Printf("Skip times aplicados com sucesso: %s", combinedOpts)
		}
	} else if util.IsDebug {
		log.Printf("Nenhum skip time disponível para o episódio %s", episode.Number)
	}
}

// promptYesNo solicita confirmação do usuário
func promptYesNo(question string) (bool, error) {
	prompt := promptStyle.Render(question) + " " + infoStyle.Render("(y/n):")
	fmt.Print(prompt + " ")
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("error reading input: %w", err)
	}
	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes" || input == "s" || input == "sim", nil
}

// playVideo reproduz o vídeo e gerencia interações
func playVideo(
	videoURL string,
	episodes []models.Episode,
	currentEpisodeNum int,
	anilistID int,
	updater *discord.RichPresenceUpdater,
) error {
	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)
	if util.IsDebug {
		log.Printf("URL do vídeo: %s", videoURL)
	}

	currentEpisode, err := getCurrentEpisode(episodes, currentEpisodeNum)
	if err != nil {
		return fmt.Errorf("📺 ❌ error getting current episode: %w", err)
	}

	mpvArgs := []string{
		"--hwdec=auto-safe",
		"--vo=gpu",
		"--profile=fast",
		"--cache=yes",
		"--demuxer-max-bytes=300M",
		"--demuxer-readahead-secs=20",
		"--no-config",
		"--video-latency-hacks=yes",
		"--audio-display=no",
	}

	tracker, resumeTime := initTracking(anilistID, currentEpisode, currentEpisodeNum)
	if resumeTime > 0 {
		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", resumeTime))
	}

	skipDataChan := fetchAniSkipAsync(anilistID, currentEpisodeNum, currentEpisode)
	socketPath, err := StartVideo(videoURL, mpvArgs)
	if err != nil {
		return fmt.Errorf("🎥 ❌ failed to start video: %w", err)
	}

	// Display current episode information
	episodeInfo := fmt.Sprintf("📺 Episode %d", currentEpisodeNum)
	episodeInfoBox := boxStyle.Render(
		headerStyle.Render("🎬 Now Playing:") + "\n" +
			infoStyle.Render(episodeInfo),
	)
	fmt.Println("\n" + episodeInfoBox)

	applyAniSkipResults(skipDataChan, socketPath, currentEpisode, currentEpisodeNum)

	if updater != nil {
		initDiscordPresence(updater, socketPath, tracker, anilistID, currentEpisode, currentEpisodeNum)
		defer updater.Stop()
	}

	currentEpisodeIndex := findEpisodeIndex(episodes, currentEpisodeNum)
	if currentEpisodeIndex == -1 {
		return fmt.Errorf("🔍 ❌ episode %d not found in list", currentEpisodeNum)
	}

	preloadNextEpisode(episodes, currentEpisodeIndex)

	stopTracking := startTrackingRoutine(tracker, socketPath, anilistID, currentEpisode, currentEpisodeNum, updater)
	defer close(stopTracking)

	reader := bufio.NewReader(os.Stdin)
	return handleUserInput(
		reader,
		socketPath,
		episodes,
		currentEpisodeIndex,
		anilistID,
		updater,
		stopTracking,
		currentEpisode,
	)
}

// getCurrentEpisode obtém o episódio atual
func getCurrentEpisode(episodes []models.Episode, num int) (*models.Episode, error) {
	if num < 1 || num > len(episodes) {
		return nil, fmt.Errorf("🔢 ❌ invalid episode number: %d", num)
	}
	return &episodes[num-1], nil
}

// initTracking inicializa o sistema de rastreamento
func initTracking(anilistID int, episode *models.Episode, episodeNum int) (*tracking.LocalTracker, int) {
	if !tracking.IsCgoEnabled {
		if util.IsDebug {
			log.Println("Rastreamento desativado: CGO não disponível")
		}
		return nil, 0
	}

	currentUser, err := user.Current()
	if err != nil {
		log.Printf("Falha ao obter usuário atual: %v", err)
		return nil, 0
	}

	var dbPath string
	if runtime.GOOS == "windows" {
		dbPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
	} else {
		dbPath = filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db")
	}

	tracker := tracking.NewLocalTracker(dbPath)
	if tracker == nil {
		return nil, 0
	}

	progress, err := tracker.GetAnime(anilistID, episode.URL)
	if err != nil || progress == nil || progress.EpisodeNumber != episodeNum || progress.PlaybackTime <= 0 {
		return tracker, 0
	}

	// Create a beautiful progress message
	progressMsg := fmt.Sprintf("📺 Episode %d, ⏱️  %d seconds", progress.EpisodeNumber, progress.PlaybackTime)
	styledProgress := boxStyle.Render(
		successStyle.Render("💾 ✓ Saved progress found:") + "\n" +
			infoStyle.Render(progressMsg),
	)
	fmt.Println("\n" + styledProgress)

	if ok, _ := promptYesNo("🔄 Would you like to resume from where you left off?"); ok {
		if util.IsDebug {
			log.Printf("Resuming from saved time: %d seconds", progress.PlaybackTime)
		}
		return tracker, progress.PlaybackTime
	}

	return tracker, 0
}

// fetchAniSkipAsync busca dados do AniSkip em paralelo
func fetchAniSkipAsync(anilistID, episodeNum int, episode *models.Episode) chan error {
	ch := make(chan error, 1)
	go func() {
		err := api.GetAndParseAniSkipData(anilistID, episodeNum, episode)
		ch <- err
	}()
	return ch
}

// applyAniSkipResults aplica os resultados do AniSkip
func applyAniSkipResults(ch chan error, socketPath string, episode *models.Episode, episodeNum int) {
	select {
	case err := <-ch:
		if err == nil {
			applySkipTimes(socketPath, episode)
		} else if util.IsDebug {
			log.Printf("Dados do AniSkip indisponíveis para episódio %d: %v", episodeNum, err)
		}
	case <-time.After(3 * time.Second):
		if util.IsDebug {
			log.Printf("Timeout ao buscar dados do AniSkip para episódio %d", episodeNum)
		}
	}
}

// initDiscordPresence inicia a presença no Discord
func initDiscordPresence(updater *discord.RichPresenceUpdater, socketPath string, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	updater.SetSocketPath(socketPath)
	updater.Start()

	go func() {
		waitForPlaybackStart(socketPath, updater)
		updateEpisodeDuration(socketPath, updater, tracker, anilistID, episode, episodeNum)
	}()
}

// waitForPlaybackStart aguarda o início da reprodução
func waitForPlaybackStart(socketPath string, updater *discord.RichPresenceUpdater) {
	for {
		timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
		if err == nil && timePos != nil && !updater.IsEpisodeStarted() {
			updater.SetEpisodeStarted(true)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// updateEpisodeDuration atualiza a duração do episódio
func updateEpisodeDuration(socketPath string, updater *discord.RichPresenceUpdater, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	for {
		if !updater.IsEpisodeStarted() || updater.GetEpisodeDuration() > 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		durationPos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
		if err != nil || durationPos == nil {
			break
		}

		duration, ok := durationPos.(float64)
		if !ok {
			break
		}

		dur := time.Duration(duration * float64(time.Second))
		if dur < time.Second {
			dur = 24 * time.Minute
		}

		updater.SetEpisodeDuration(dur)

		if tracker != nil && dur > 0 {
			anime := tracking.Anime{
				AnilistID:     anilistID,
				AllanimeID:    episode.URL,
				EpisodeNumber: episodeNum,
				Duration:      int(dur.Seconds()),
				Title:         getEpisodeTitle(episode.Title),
				LastUpdated:   time.Now(),
			}
			if err := tracker.UpdateProgress(anime); err != nil && util.IsDebug {
				log.Printf("Falha ao atualizar rastreamento: %v", err)
			}
		}
		break
	}
}

// getEpisodeTitle obtém o título do episódio
func getEpisodeTitle(title models.TitleDetails) string {
	if title.English != "" {
		return title.English
	}
	if title.Romaji != "" {
		return title.Romaji
	}
	if title.Japanese != "" {
		return title.Japanese
	}
	return "Sem título"
}

// findEpisodeIndex encontra o índice do episódio
func findEpisodeIndex(episodes []models.Episode, num int) int {
	episodeStr := strconv.Itoa(num)
	for i, ep := range episodes {
		if ExtractEpisodeNumber(ep.Number) == episodeStr {
			return i
		}
	}
	return -1
}

// preloadNextEpisode pré-carrega o próximo episódio
func preloadNextEpisode(episodes []models.Episode, currentIndex int) {
	if currentIndex+1 >= len(episodes) {
		return
	}

	go func() {
		_, err := GetVideoURLForEpisode(episodes[currentIndex+1].URL)
		if err != nil && util.IsDebug {
			log.Printf("Erro no pré-carregamento: %v", err)
		}
	}()
}

// startTrackingRoutine inicia rotina de rastreamento
func startTrackingRoutine(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) chan struct{} {
	stopChan := make(chan struct{})
	if tracker == nil {
		return stopChan
	}

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				updateTracking(tracker, socketPath, anilistID, episode, episodeNum, updater)
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// updateTracking atualiza o rastreamento
func updateTracking(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) {
	timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
	if err != nil || timePos == nil {
		return
	}

	position, ok := timePos.(float64)
	if !ok {
		return
	}

	duration := 1440
	if updater != nil {
		duration = int(updater.GetEpisodeDuration().Seconds())
	}

	anime := tracking.Anime{
		AnilistID:     anilistID,
		AllanimeID:    episode.URL,
		EpisodeNumber: episodeNum,
		PlaybackTime:  int(position),
		Duration:      duration,
		Title:         getEpisodeTitle(episode.Title),
		LastUpdated:   time.Now(),
	}

	if err := tracker.UpdateProgress(anime); err != nil && util.IsDebug {
		log.Printf("Erro ao atualizar rastreamento: %v", err)
	}
}

// handleUserInput gerencia entrada do usuário
func handleUserInput(
	reader *bufio.Reader,
	socketPath string,
	episodes []models.Episode,
	currentIndex int,
	anilistID int,
	updater *discord.RichPresenceUpdater,
	stopTracking chan struct{},
	currentEpisode *models.Episode,
) error {
	// Create a beautiful commands menu
	commandsTitle := headerStyle.Render("🎮 Available Commands:")
	commands := []string{
		commandStyle.Render("n") + " 🚀 Next episode",
		commandStyle.Render("p") + " ⬅️  Previous episode",
		commandStyle.Render("e") + " 📋 Select episode",
		commandStyle.Render("q") + " 🚪 Exit",
		commandStyle.Render("s") + " ⏭️  Skip intro",
	}

	commandsBox := boxStyle.Render(
		commandsTitle + "\n" +
			strings.Join(commands, "\n"),
	)
	fmt.Println("\n" + commandsBox)

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			return fmt.Errorf("⌨️ ❌ error reading input: %w", err)
		}

		switch char {
		case 'n':
			nextMsg := actionStyle.Render("🚀 ➡️  Switching to next episode...")
			fmt.Println(nextMsg)
			return playNextEpisode(currentIndex+1, episodes, anilistID, updater, stopTracking, socketPath)
		case 'p':
			prevMsg := actionStyle.Render("⬅️ 🔙 Switching to previous episode...")
			fmt.Println(prevMsg)
			return playPreviousEpisode(currentIndex-1, episodes, anilistID, updater, stopTracking, socketPath)
		case 'q':
			quitMsg := infoStyle.Render("🚪 ✨ Goodbye! Thanks for watching!")
			fmt.Println(quitMsg)
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return nil
		case 'e':
			selectMsg := actionStyle.Render("📋 🔍 Opening episode selector...")
			fmt.Println(selectMsg)
			return selectEpisode(episodes, anilistID, updater, stopTracking, socketPath)
		case 's':
			skipMsg := actionStyle.Render("⏭️ ⚡ Attempting to skip intro...")
			fmt.Println(skipMsg)
			skipIntro(socketPath, currentEpisode)
		default:
			invalidMsg := warningStyle.Render("❌ Invalid command. Use: n, p, e, q or s")
			fmt.Println(invalidMsg)
		}
	}
}

// playNextEpisode reproduz próximo episódio
func playNextEpisode(newIndex int, episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	if newIndex >= len(episodes) {
		msg := infoStyle.Render("🎬 You are on the last episode")
		fmt.Println(msg)
		return nil
	}
	return switchEpisode(newIndex, episodes, anilistID, updater, stopTracking, socketPath)
}

// playPreviousEpisode reproduz episódio anterior
func playPreviousEpisode(newIndex int, episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	if newIndex < 0 {
		msg := infoStyle.Render("🎬 You are on the first episode")
		fmt.Println(msg)
		return nil
	}
	return switchEpisode(newIndex, episodes, anilistID, updater, stopTracking, socketPath)
}

// selectEpisode permite selecionar um episódio
func selectEpisode(episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	selectedURL, selectedNumStr, err := SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		return fmt.Errorf("🔍 ❌ failed to select episode: %w", err)
	}

	for i, ep := range episodes {
		if ep.URL == selectedURL {
			return switchEpisode(i, episodes, anilistID, updater, stopTracking, socketPath)
		}
	}

	return fmt.Errorf("📺 ❌ episode %s not found", selectedNumStr)
}

// switchEpisode alterna entre episódios
func switchEpisode(newIndex int, episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	target := episodes[newIndex]
	targetNum, err := strconv.Atoi(ExtractEpisodeNumber(target.Number))
	if err != nil {
		return fmt.Errorf("🔢 ❌ invalid episode number: %w", err)
	}

	targetURL, err := GetVideoURLForEpisode(target.URL)
	if err != nil {
		return fmt.Errorf("🌐 ❌ failed to get video URL: %w", err)
	}

	if updater != nil {
		updater.Stop()
	}

	close(stopTracking)
	_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})

	var newUpdater *discord.RichPresenceUpdater
	if updater != nil {
		duration := time.Duration(target.Duration) * time.Second
		newUpdater = discord.NewRichPresenceUpdater(
			updater.GetAnime(),
			updater.GetIsPaused(),
			updater.GetAnimeMutex(),
			updater.GetUpdateFreq(),
			duration,
			"",
			MpvSendCommand,
		)
		newUpdater.SetEpisodeStarted(false)
	}

	return playVideo(targetURL, episodes, targetNum, anilistID, newUpdater)
}

// skipIntro pula a introdução
func skipIntro(socketPath string, episode *models.Episode) {
	if episode.SkipTimes.Op.End > 0 {
		_, _ = mpvSendCommand(socketPath, []interface{}{"seek", episode.SkipTimes.Op.End, "absolute"})
		skipMsg := successStyle.Render(fmt.Sprintf("⏭️ ✓ Intro skipped to %ds", episode.SkipTimes.Op.End))
		fmt.Println(skipMsg)
	} else {
		noSkipMsg := warningStyle.Render("⚠️ ❌ Skip intro data not available")
		fmt.Println(noSkipMsg)
	}
}
