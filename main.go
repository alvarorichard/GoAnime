package main

import (
	"flag"
	"fmt"
	"github.com/alvarorichard/Goanime/api"
	"github.com/alvarorichard/Goanime/player"
	"github.com/manifoldco/promptui"
	"log"
	"os"
	"strconv"
	"strings"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug mode")
	flag.Parse()
	api.IsDebug = *debug

	animeName := getUserInput("Enter anime name")
	animeURL, err := api.SearchAnime(treatingAnimeName(animeName))
	if err != nil {
		log.Fatalf("Failed to get anime episodes: %v", err)
		os.Exit(1)
	}

	episodes, err := api.GetAnimeEpisodes(animeURL)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime has no episodes in the server.")
		os.Exit(1)
	}

	series, totalEpisodes, err := api.IsSeries(animeURL)
	if err != nil {
		log.Fatalf("Erro ao verificar se o anime é uma série: %v", err)
	}

	if series {
		fmt.Printf("O anime selecionado é uma série com %d episódios.\n", totalEpisodes)
		selectedEpisodeURL, episodeNumberStr := player.SelectEpisode(episodes)

		// A função extractEpisodeNumber não deve ser chamada para filmes/OVAs.
		// Este ajuste é específico para quando sabemos que é uma série com base na verificação anterior.
		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
		if err != nil {
			log.Fatalf("Error parsing episode number: %v", err)
		}

		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
		if err != nil {
			log.Fatalf("Failed to extract video URL: %v", err)
		}

		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, animeURL, episodeNumberStr)
	} else {
		fmt.Println("O anime selecionado é um filme/OVA. Iniciando a reprodução direta...")
		// Para filmes/OVAs, utilizamos o primeiro episódio diretamente sem tentar parsear um número de episódio.
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalf("Failed to extract video URL: %v", err)
		}

		// Assume-se 1 como o número de episódio padrão para filmes/OVAs.
		player.HandleDownloadAndPlay(videoURL, episodes, 1, animeURL, episodes[0].Number)
	}
}

func getUserInput(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	result, err := prompt.Run()
	if err != nil {
		log.Fatalf("Error acquiring user input: %v", err)
	}
	return result
}

func treatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	return strings.ReplaceAll(loweredName, " ", "-")
}
