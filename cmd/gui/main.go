//package main
//
//import (
//	"fmt"
//	"log"
//	"os"
//	"os/exec"
//	"strings"
//	"syscall"
//
//	"fyne.io/fyne/v2"
//	"fyne.io/fyne/v2/app"
//	"fyne.io/fyne/v2/container"
//	"fyne.io/fyne/v2/widget"
//
//	"github.com/alvarorichard/Goanime/internal/api"
//)
//
//const (
//	mpvExecutable = "mpv" // Use caminho absoluto se necessário (ex: "/usr/local/bin/mpv")
//	mpvArgs       = "--no-terminal --force-window --script-opts=ytdl_hook-ytdl_path=yt-dlp"
//)
//
//func main() {
//	a := app.New()
//	w := a.NewWindow("GoAnime Player")
//	w.SetFixedSize(true)
//	w.Resize(fyne.NewSize(800, 600))
//
//	// Componentes UI
//	searchEntry := widget.NewEntry()
//	searchEntry.SetPlaceHolder("Digite o nome do anime...")
//
//	statusLabel := widget.NewLabel("Pronto para buscar")
//	playerStatus := widget.NewLabel("")
//
//	var animeData []api.GUIAnime
//	var episodeData []api.Episode
//
//	// Listas
//	animeList := widget.NewList(
//		func() int { return len(animeData) },
//		func() fyne.CanvasObject { return widget.NewLabel("") },
//		func(id widget.ListItemID, obj fyne.CanvasObject) {
//			obj.(*widget.Label).SetText(animeData[id].Name)
//		},
//	)
//
//	episodeList := widget.NewList(
//		func() int { return len(episodeData) },
//		func() fyne.CanvasObject { return widget.NewLabel("") },
//		func(id widget.ListItemID, obj fyne.CanvasObject) {
//			obj.(*widget.Label).SetText(fmt.Sprintf("Episódio %s", episodeData[id].Number))
//		},
//	)
//
//	// Handlers
//	searchHandler := func() {
//		query := strings.TrimSpace(searchEntry.Text)
//		if len(query) < 3 {
//			statusLabel.SetText("Digite pelo menos 3 caracteres")
//			return
//		}
//
//		statusLabel.SetText("Buscando...")
//		go func() {
//			defer func() {
//				if r := recover(); r != nil {
//					log.Printf("Panic in search: %v", r)
//				}
//			}()
//
//			results, err := api.SearchAnimeGUI(query)
//			if err != nil {
//				a.SendNotification(fyne.NewNotification("Erro", "Falha na conexão com o servidor"))
//				return
//			}
//
//			animeData = results
//			animeList.Refresh()
//			statusLabel.SetText(fmt.Sprintf("%d resultados encontrados", len(results)))
//		}()
//	}
//
//	animeList.OnSelected = func(id widget.ListItemID) {
//		if id < 0 || id >= len(animeData) {
//			return
//		}
//
//		statusLabel.SetText("Carregando episódios...")
//		go func() {
//			episodes, err := api.GetAnimeEpisodes(animeData[id].URL)
//			if err != nil {
//				a.SendNotification(fyne.NewNotification("Erro", "Falha ao carregar episódios"))
//				return
//			}
//
//			episodeData = episodes
//			episodeList.Refresh()
//			statusLabel.SetText(fmt.Sprintf("%d episódios carregados", len(episodes)))
//		}()
//	}
//
//	episodeList.OnSelected = func(id widget.ListItemID) {
//		if id < 0 || id >= len(episodeData) {
//			return
//		}
//
//		ep := episodeData[id]
//		playerStatus.SetText(fmt.Sprintf("Iniciando Episódio %s...", ep.Number))
//		log.Printf("Tentando reproduzir: %s", ep.URL)
//
//		go func(url string) {
//			args := append(strings.Split(mpvArgs, " "), url)
//			cmd := exec.Command(mpvExecutable, args...)
//
//			// Configurar saída
//			cmd.Stdout = os.Stdout
//			cmd.Stderr = os.Stderr
//			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
//
//			if err := cmd.Start(); err != nil {
//				log.Printf("Erro MPV: %v", err)
//				a.SendNotification(fyne.NewNotification("Erro", fmt.Sprintf("Falha: %v", err)))
//				return
//			}
//
//			playerStatus.SetText(fmt.Sprintf("Reproduzindo Episódio %s", ep.Number))
//			defer cmd.Wait()
//
//			if err := cmd.Wait(); err != nil {
//				if exiterr, ok := err.(*exec.ExitError); ok {
//					if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
//						log.Printf("Código de saída: %d", status.ExitStatus())
//					}
//				}
//				log.Printf("Erro na reprodução: %v", err)
//				a.SendNotification(fyne.NewNotification("Erro", "Problema na reprodução"))
//			}
//		}(ep.URL)
//	}
//
//	// Layout
//	w.SetContent(container.NewBorder(
//		container.NewVBox(
//			container.NewBorder(nil, nil, nil,
//				widget.NewButton("Buscar", searchHandler),
//				searchEntry,
//			),
//			statusLabel,
//		),
//		playerStatus,
//		nil,
//		nil,
//		container.NewHSplit(
//			animeList,
//			episodeList,
//		),
//	))
//
//	w.ShowAndRun()
//}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/PuerkitoBio/goquery"

	// Ajuste o import para o local do seu projeto (onde ficam SearchAnimeGUI e GetAnimeEpisodes)
	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/util"
)

//
// NENHUM tipo local "GUIAnime" ou "Episode" duplicado!
// Usamos diretamente api.GUIAnime e api.Episode.
//

// ---------------------------------------------------------------------
// Estruturas para parse do JSON do player (caso o site retorne JSON com "data")
// ---------------------------------------------------------------------
type VideoData struct {
	Src   string `json:"src"`
	Label string `json:"label"`
}
type VideoResponse struct {
	Data []VideoData `json:"data"`
}

// ---------------------------------------------------------------------
// 1. Funções de "scraping" para obter link final do episódio
// ---------------------------------------------------------------------

// GetVideoURLForEpisode chama internamente extractVideoURL e extractActualVideoURL.
// Se o site exigir outro scraping, você ajusta aqui.
func GetVideoURLForEpisode(episodeURL string) (string, error) {
	if util.IsDebug {
		log.Printf("[DEBUG] GetVideoURLForEpisode -> %s", episodeURL)
	}
	// 1) Tenta achar <video data-video-src> ou link blogger
	videoURL, err := extractVideoURL(episodeURL)
	if err != nil {
		return "", err
	}
	// 2) Se não for blogger, parse do JSON e pega link final .mp4 etc.
	return extractActualVideoURL(videoURL)
}

// Procura <video data-video-src> ou <div data-video-src> no HTML
func extractVideoURL(url string) (string, error) {
	// Ajuste se quiser outro método HTTP
	resp, err := api.SafeGet(url)
	if err != nil {
		return "", fmt.Errorf("erro http: %w", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("erro parse HTML: %w", err)
	}

	// Tenta <video data-video-src="...">
	videoEl := doc.Find("video[data-video-src]")
	if videoEl.Length() == 0 {
		// Tenta <div data-video-src="...">
		videoEl = doc.Find("div[data-video-src]")
	}
	if videoEl.Length() == 0 {
		return "", errors.New("nenhum data-video-src encontrado no HTML")
	}

	videoSrc, ok := videoEl.Attr("data-video-src")
	if !ok || videoSrc == "" {
		// Se não achou, tenta blogger link no HTML cru
		htmlBody, err2 := fetchContent(url)
		if err2 != nil {
			return "", err2
		}
		videoSrc, err2 = findBloggerLink(htmlBody)
		if err2 != nil {
			return "", err2
		}
	}
	return videoSrc, nil
}

// Se "videoSrc" contiver "blogger.com", retorna direto.
// Caso contrário, faz GET desse link, parseia JSON e seleciona a maior qualidade.
func extractActualVideoURL(videoSrc string) (string, error) {
	if strings.Contains(videoSrc, "blogger.com") {
		// Link direto
		return videoSrc, nil
	}

	resp, err := api.SafeGet(videoSrc)
	if err != nil {
		return "", fmt.Errorf("erro ao buscar videoSrc: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status HTTP %d", resp.StatusCode)
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("falha ao ler corpo: %w", err)
	}

	var vr VideoResponse
	if err := json.Unmarshal(bodyBytes, &vr); err != nil {
		return "", fmt.Errorf("erro unmarshal JSON: %w", err)
	}
	if len(vr.Data) == 0 {
		return "", errors.New("nenhum 'data' disponível no JSON")
	}

	// Seleciona a maior qualidade
	return selectHighestQualityVideo(vr.Data), nil
}

// fetchContent pega todo o HTML da URL (caso precise varrer).
func fetchContent(u string) (string, error) {
	resp, err := api.SafeGet(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// findBloggerLink acha https://www.blogger.com/video.g?token=... no HTML
func findBloggerLink(content string) (string, error) {
	pattern := `https://www\.blogger\.com/video\.g\?token=[A-Za-z0-9_-]+`
	re := regexp.MustCompile(pattern)
	match := re.FindString(content)
	if match == "" {
		return "", errors.New("nenhum blogger link encontrado")
	}
	return match, nil
}

// selectHighestQualityVideo percorre o slice, acha o maior "Label" (ex.: "720p","1080p").
func selectHighestQualityVideo(data []VideoData) string {
	var best int
	var bestURL string
	for _, d := range data {
		q, _ := strconv.Atoi(strings.TrimRight(d.Label, "p"))
		if q > best {
			best = q
			bestURL = d.Src
		}
	}
	return bestURL
}

// ---------------------------------------------------------------------
// 2. Função para abrir MPV e reproduzir sem prompts de terminal
// ---------------------------------------------------------------------
func PlayVideoGUI(finalURL string) error {
	log.Printf("[DEBUG] Abrindo MPV com URL: %s", finalURL)

	args := []string{
		"--no-terminal",
		"--force-window",
		"--quiet",
		finalURL,
	}
	cmd := exec.Command("mpv", args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("falha ao iniciar mpv: %w", err)
	}
	log.Printf("[DEBUG] MPV iniciado (PID=%d)", cmd.Process.Pid)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("MPV encerrou com erro: %w", err)
	}
	log.Println("[DEBUG] MPV finalizado com sucesso.")
	return nil
}

// ---------------------------------------------------------------------
// 3. MAIN: Fyne, sem concorrência e sem tipos duplicados
// ---------------------------------------------------------------------
func main() {
	a := app.New()
	w := a.NewWindow("GoAnime - sem concurrency, usando api.GUIAnime / api.Episode")
	w.Resize(fyne.NewSize(1000, 700))

	searchEntry := widget.NewEntry()
	searchEntry.SetPlaceHolder("Ex: Black Clover, Naruto...")

	statusLabel := widget.NewLabel("Digite algo e clique em Buscar")

	debugConsole := widget.NewMultiLineEntry()
	debugConsole.Disable()

	// Listas
	// Aqui armazenamos:
	var currentAnimes []api.GUIAnime
	var currentEpisodes []api.Episode

	// Precisamos criar a widget.List para animes e episódios
	animeList := widget.NewList(
		// length
		func() int { return len(currentAnimes) },
		// create
		func() fyne.CanvasObject { return widget.NewLabel("") },
		// update
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			if i < 0 || i >= len(currentAnimes) {
				return
			}
			obj.(*widget.Label).SetText(currentAnimes[i].Name)
		},
	)

	episodeList := widget.NewList(
		func() int { return len(currentEpisodes) },
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(i widget.ListItemID, obj fyne.CanvasObject) {
			if i < 0 || i >= len(currentEpisodes) {
				return
			}
			ep := currentEpisodes[i]
			obj.(*widget.Label).SetText(fmt.Sprintf("Ep. %s - %s", ep.Number, ep.Title))
		},
	)

	// Botão de busca (sincrono)
	searchBtn := widget.NewButtonWithIcon("Buscar", theme.SearchIcon(), func() {
		query := strings.TrimSpace(searchEntry.Text)
		if len(query) < 3 {
			statusLabel.SetText("Digite ao menos 3 caracteres!")
			return
		}
		statusLabel.SetText("Buscando... (UI congela até terminar)")

		// Chamamos api.SearchAnimeGUI
		results, err := api.SearchAnimeGUI(query)
		if err != nil {
			log.Printf("Erro na busca: %v", err)
			statusLabel.SetText("Erro na busca (veja log)")
			return
		}
		// results é []api.GUIAnime
		currentAnimes = results

		// Forçamos a lista a recarregar
		animeList.Refresh()
		statusLabel.SetText(fmt.Sprintf("Encontrados %d animes", len(currentAnimes)))
	})

	// Ao selecionar um anime
	animeList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(currentAnimes) {
			return
		}
		selectedAnime := currentAnimes[id]
		statusLabel.SetText("Carregando episódios... (UI congela)")

		eps, err := api.GetAnimeEpisodes(selectedAnime.URL)
		if err != nil {
			log.Printf("Erro ao pegar episódios: %v", err)
			statusLabel.SetText("Erro ao pegar episódios (ver log)")
			return
		}
		currentEpisodes = eps

		episodeList.Refresh()
		statusLabel.SetText(fmt.Sprintf("Carregados %d episódios", len(currentEpisodes)))
	}

	// Ao selecionar um episódio
	episodeList.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(currentEpisodes) {
			return
		}
		ep := currentEpisodes[id]
		statusLabel.SetText(fmt.Sprintf("Extraindo link do Ep. %s... (UI congela)", ep.Number))

		// Sincrono
		finalURL, err := GetVideoURLForEpisode(ep.URL)
		if err != nil {
			log.Printf("Erro ao extrair link: %v", err)
			statusLabel.SetText("Erro ao extrair link (veja log)")
			return
		}

		// Mostra no debugConsole
		debugConsole.SetText(debugConsole.Text + fmt.Sprintf("\n[DEBUG] finalURL: %s", finalURL))

		statusLabel.SetText("Iniciando MPV...")

		// Abre o MPV
		err = PlayVideoGUI(finalURL)
		if err != nil {
			log.Printf("Erro ao reproduzir: %v", err)
			statusLabel.SetText("Erro ao reproduzir (veja log)")
			return
		}
		statusLabel.SetText("Reprodução finalizada.")
	}

	// Layout
	topBar := container.NewBorder(
		nil,
		nil,
		nil,
		searchBtn,
		searchEntry,
	)

	animeScroll := container.NewVScroll(animeList)
	episodeScroll := container.NewVScroll(episodeList)
	listsSplit := container.NewHSplit(animeScroll, episodeScroll)
	listsSplit.SetOffset(0.3)

	bottomBox := container.NewVScroll(debugConsole)

	mainContainer := container.NewBorder(
		container.NewVBox(topBar, statusLabel),
		bottomBox,
		nil,
		nil,
		listsSplit,
	)

	w.SetContent(mainContainer)
	w.ShowAndRun()
}
