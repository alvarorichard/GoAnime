//package main
//
//import (
//	"encoding/json"
//	"errors"
//	"fmt"
//	"github.com/PuerkitoBio/goquery"
//	"io"
//	"log"
//	"net/http"
//	"os"
//	"os/exec"
//	"path/filepath"
//	"regexp"
//	"strconv"
//	"strings"
//	"sync"
//	"syscall"
//	"time"
//
//	// Ajuste de acordo com seu projeto
//	"github.com/alvarorichard/Goanime/internal/api"
//	"github.com/alvarorichard/Goanime/internal/util"
//)
//
//// ----------------------------------------------------------------------------
//// Estruturas e tipos auxiliares
//// ----------------------------------------------------------------------------
//
//// VideoData e VideoResponse são exemplos de JSON para sites
//type VideoData struct {
//	Src   string `json:"src"`
//	Label string `json:"label"`
//}
//type VideoResponse struct {
//	Data []VideoData `json:"data"`
//}
//
//type GUIAnime struct {
//	Name string
//	URL  string
//}
//type Episode struct {
//	Number string
//	Title  string
//	URL    string
//}
//
//// DownloadProgressFunc permite atualizar a GUI sobre o progresso (0..N bytes já baixados).
//type DownloadProgressFunc func(received, total int64)
//
//// ----------------------------------------------------------------------------
//// 1. Download sem TUI/PromptUI/BubbleTea
//// ----------------------------------------------------------------------------
//
//// DownloadVideo baixa um vídeo de `url` em `numThreads` partes e salva em `destPath`.
//// O callback `onProgress` (se não for nil) recebe updates de quantos bytes já foram baixados, e o total.
//func DownloadVideo(url, destPath string, numThreads int, onProgress DownloadProgressFunc) error {
//	// Cria um client com timeout seguro
//	httpClient := &http.Client{
//		Transport: api.SafeTransport(10 * time.Second),
//	}
//
//	// 1) Descobre tamanho total do arquivo
//	contentLength, err := getContentLength(url, httpClient)
//	if err != nil {
//		return fmt.Errorf("erro ao obter tamanho do conteúdo: %w", err)
//	}
//	if contentLength <= 0 {
//		return fmt.Errorf("conteúdo inválido ou zero bytes")
//	}
//
//	// 2) Garante que a pasta de destino exista
//	if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
//		return fmt.Errorf("falha ao criar pasta de destino: %w", err)
//	}
//
//	// 3) Calcula o tamanho de cada chunk
//	chunkSize := contentLength / int64(numThreads)
//
//	// Vamos acompanhar quantos bytes já baixamos
//	var received int64
//	var mu sync.Mutex
//	updateProgress := func(n int64) {
//		mu.Lock()
//		defer mu.Unlock()
//		received += n
//		if onProgress != nil {
//			onProgress(received, contentLength)
//		}
//	}
//
//	// 4) Inicia downloads em paralelo
//	var wg sync.WaitGroup
//	for i := 0; i < numThreads; i++ {
//		from := int64(i) * chunkSize
//		to := from + chunkSize - 1
//		if i == numThreads-1 {
//			to = contentLength - 1 // Último chunk vai até o fim
//		}
//
//		wg.Add(1)
//		go func(partIndex int, from, to int64) {
//			defer wg.Done()
//			if err := downloadPart(url, from, to, partIndex, httpClient, destPath, updateProgress); err != nil {
//				log.Printf("Erro ao baixar parte %d: %v", partIndex, err)
//			}
//		}(i, from, to)
//	}
//	wg.Wait()
//
//	// 5) Combina as partes em um só arquivo
//	if err := combineParts(destPath, numThreads); err != nil {
//		return fmt.Errorf("erro ao combinar partes: %w", err)
//	}
//
//	return nil
//}
//
//// getContentLength faz um HEAD (ou GET range=0-0) para obter o tamanho do arquivo.
//func getContentLength(url string, client *http.Client) (int64, error) {
//	req, err := http.NewRequest("HEAD", url, nil)
//	if err != nil {
//		return 0, err
//	}
//
//	resp, err := client.Do(req)
//	if err != nil || resp.StatusCode == http.StatusMethodNotAllowed || resp.StatusCode == http.StatusNotImplemented {
//		req.Method = "GET"
//		req.Header.Set("Range", "bytes=0-0")
//		resp, err = client.Do(req)
//		if err != nil {
//			return 0, err
//		}
//	}
//	defer resp.Body.Close()
//
//	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
//		return 0, fmt.Errorf("status %d (servidor não suporta partial content?)", resp.StatusCode)
//	}
//	cl := resp.Header.Get("Content-Length")
//	if cl == "" {
//		return 0, fmt.Errorf("Content-Length ausente")
//	}
//	length, err := strconv.ParseInt(cl, 10, 64)
//	if err != nil {
//		return 0, err
//	}
//	return length, nil
//}
//
//// downloadPart faz o download de [from..to] bytes e salva em arquivo temporário .partX
//func downloadPart(
//	url string,
//	from, to int64,
//	partIndex int,
//	client *http.Client,
//	destPath string,
//	onData func(n int64),
//) error {
//	req, err := http.NewRequest("GET", url, nil)
//	if err != nil {
//		return err
//	}
//	req.Header.Add("Range", fmt.Sprintf("bytes=%d-%d", from, to))
//
//	resp, err := client.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//
//	partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), partIndex)
//	partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
//
//	f, err := os.Create(partFilePath)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//
//	buf := make([]byte, 32*1024)
//	for {
//		n, err := resp.Body.Read(buf)
//		if n > 0 {
//			if _, werr := f.Write(buf[:n]); werr != nil {
//				return werr
//			}
//			if onData != nil {
//				onData(int64(n))
//			}
//		}
//		if err == io.EOF {
//			break
//		}
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// combineParts concatena todos os .partX no destino final.
//func combineParts(destPath string, numThreads int) error {
//	out, err := os.Create(destPath)
//	if err != nil {
//		return err
//	}
//	defer out.Close()
//
//	for i := 0; i < numThreads; i++ {
//		partFileName := fmt.Sprintf("%s.part%d", filepath.Base(destPath), i)
//		partFilePath := filepath.Join(filepath.Dir(destPath), partFileName)
//
//		p, err := os.Open(partFilePath)
//		if err != nil {
//			return err
//		}
//		if _, err := io.Copy(out, p); err != nil {
//			p.Close()
//			return err
//		}
//		p.Close()
//		os.Remove(partFilePath)
//	}
//	return nil
//}
//
//// ----------------------------------------------------------------------------
//// 2. Reprodução via MPV (sem interação terminal)
//// ----------------------------------------------------------------------------
//
//// PlayVideoGUI inicia o MPV sem prompt, com socket IPC (se quiser), sem Bubble Tea.
////func PlayVideoGUI(videoURL string) error {
////	// Geramos um sufixo random para o socket
////	randomBytes := make([]byte, 4)
////	_, err := rand.Read(randomBytes)
////	if err != nil {
////		return fmt.Errorf("failed to generate random for socket: %w", err)
////	}
////	randomSuffix := fmt.Sprintf("%x", randomBytes)
////
////	var socketPath string
////	if runtime.GOOS == "windows" {
////		// Exemplo de pipe no Windows
////		socketPath = fmt.Sprintf(`\\.\pipe\goanime_mpvsocket_%s`, randomSuffix)
////	} else {
////		// Em sistemas UNIX-like
////		socketPath = fmt.Sprintf("/tmp/goanime_mpvsocket_%s", randomSuffix)
////	}
////
////	// Monta os argumentos básicos para o MPV
////	args := []string{
////		"--no-terminal",
////		"--quiet",
////		fmt.Sprintf("--input-ipc-server=%s", socketPath),
////		videoURL,
////	}
////	args = append(args, extraArgs...) // Se precisar passar args adicionais
////
////	cmd := exec.Command("mpv", args...)
////	cmd.Stdout = os.Stdout
////	cmd.Stderr = os.Stderr
////
////	if err := cmd.Start(); err != nil {
////		return fmt.Errorf("falha ao iniciar mpv: %w", err)
////	}
////	log.Printf("MPV iniciado (PID: %d) com URL: %s", cmd.Process.Pid, videoURL)
////
////	// Espera o MPV encerrar
////	if err := cmd.Wait(); err != nil {
////		return err
////	}
////	log.Println("MPV encerrado.")
////	return nil
////}
//
//func PlayVideoGUI(finalURL string) error {
//	log.Printf("Iniciando MPV com URL: %s", finalURL)
//
//	args := []string{
//		"--no-terminal",
//		"--force-window",
//		"--quiet",
//		finalURL,
//	}
//
//	cmd := exec.Command("mpv", args...)
//	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
//	cmd.Stdout = os.Stdout
//	cmd.Stderr = os.Stderr
//
//	if err := cmd.Start(); err != nil {
//		return fmt.Errorf("falha ao iniciar mpv: %w", err)
//	}
//	log.Printf("MPV iniciado (PID=%d)", cmd.Process.Pid)
//
//	if err := cmd.Wait(); err != nil {
//		return fmt.Errorf("MPV retornou erro: %w", err)
//	}
//	log.Println("MPV finalizado.")
//	return nil
//}
//
//// ----------------------------------------------------------------------------
//// 3. Exemplo de scraping: GetVideoURLForEpisode
//// ----------------------------------------------------------------------------
//
//// Se seu site exigir scraping real, você ajusta aqui (ex.: extrair .mp4).
//
//func extractVideoURL(url string) (string, error) {
//
//	if util.IsDebug {
//		log.Printf("Extraindo URL de vídeo da página: %s", url)
//	}
//
//	response, err := api.SafeGet(url)
//	if err != nil {
//		return "", errors.New(fmt.Sprintf("failed to fetch URL: %+v", err))
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//			log.Printf("Failed to close response body: %v\n", err)
//		}
//	}(response.Body)
//
//	doc, err := goquery.NewDocumentFromReader(response.Body)
//	if err != nil {
//		return "", errors.New(fmt.Sprintf("failed to parse HTML: %+v", err))
//	}
//
//	videoElements := doc.Find("video")
//	if videoElements.Length() == 0 {
//		videoElements = doc.Find("div")
//	}
//
//	if videoElements.Length() == 0 {
//		return "", errors.New("no video elements found in the HTML")
//	}
//
//	videoSrc, exists := videoElements.Attr("data-video-src")
//	if !exists || videoSrc == "" {
//		urlBody, err := fetchContent(url)
//		if err != nil {
//			return "", err
//		}
//		videoSrc, err = findBloggerLink(urlBody)
//		if err != nil {
//			return "", err
//		}
//	}
//
//	return videoSrc, nil
//}
//func GetVideoURLForEpisode(episodeURL string) (string, error) {
//
//	if util.IsDebug {
//		log.Printf("Tentando extrair URL de vídeo para o episódio: %s", episodeURL)
//	}
//	videoURL, err := extractVideoURL(episodeURL)
//	if err != nil {
//		return "", err
//	}
//	return extractActualVideoURL(videoURL)
//}
//
//func findBloggerLink(content string) (string, error) {
//	pattern := `https://www\.blogger\.com/video\.g\?token=([A-Za-z0-9_-]+)`
//
//	re := regexp.MustCompile(pattern)
//	matches := re.FindStringSubmatch(content)
//
//	if len(matches) > 0 {
//		return matches[0], nil
//	} else {
//		return "", errors.New("no blogger video link found in the content")
//	}
//}
//
//func extractActualVideoURL(videoSrc string) (string, error) {
//	if strings.Contains(videoSrc, "blogger.com") {
//		return videoSrc, nil
//	}
//	response, err := api.SafeGet(videoSrc)
//	if err != nil {
//		return "", errors.New(fmt.Sprintf("failed to fetch video source: %+v", err))
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//			log.Printf("Failed to close response body: %v\n", err)
//		}
//	}(response.Body)
//
//	if response.StatusCode != http.StatusOK {
//		return "", errors.New(fmt.Sprintf("request failed with status: %s", response.Status))
//	}
//
//	body, err := io.ReadAll(response.Body)
//	if err != nil {
//		return "", errors.New(fmt.Sprintf("failed to read response body: %+v", err))
//	}
//
//	var videoResponse VideoResponse
//	if err := json.Unmarshal(body, &videoResponse); err != nil {
//		return "", errors.New(fmt.Sprintf("failed to unmarshal JSON response: %+v", err))
//	}
//
//	if len(videoResponse.Data) == 0 {
//		return "", errors.New("no video data found in the response")
//	}
//
//	highestQualityVideoURL := selectHighestQualityVideo(videoResponse.Data)
//	if highestQualityVideoURL == "" {
//		return "", errors.New("no suitable video quality found")
//	}
//
//	return highestQualityVideoURL, nil
//}
//
//func fetchContent(url string) (string, error) {
//	resp, err := api.SafeGet(url)
//	if err != nil {
//		return "", err
//	}
//	defer func(Body io.ReadCloser) {
//		err := Body.Close()
//		if err != nil {
//			log.Printf("Failed to close response body: %v\n", err)
//		}
//	}(resp.Body)
//
//	body, err := io.ReadAll(resp.Body)
//	if err != nil {
//		return "", err
//	}
//
//	return string(body), nil
//}
//
//func selectHighestQualityVideo(videos []VideoData) string {
//	var highestQuality int
//	var highestQualityURL string
//	for _, video := range videos {
//		qualityValue, _ := strconv.Atoi(strings.TrimRight(video.Label, "p"))
//		if qualityValue > highestQuality {
//			highestQuality = qualityValue
//			highestQualityURL = video.Src
//		}
//	}
//	return highestQualityURL
//}
