package main

import (
	"fmt"
	"github.com/alvarorichard/Goanime/api"
	"github.com/alvarorichard/Goanime/player"
	"github.com/alvarorichard/Goanime/util"
	"log"
	"strconv"
)

func main() {

	animeName := util.FlagParser()
	animeURL, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Failed to get anime episodes:", util.ErrorHandler(err))
	}

	episodes, err := api.GetAnimeEpisodes(animeURL)
	if err != nil || len(episodes) == 0 {
		if util.IsDebug {
			log.Fatalln("The selected anime has no episodes in the server:", util.ErrorHandler(err))
		}
		log.Fatalln("The selected anime has no episodes in the server.")
	}

	series, totalEpisodes, err := api.IsSeries(animeURL)
	if err != nil {
		log.Fatalln("Erro ao verificar se o anime é uma série:", util.ErrorHandler(err))
	}

	if series {
		fmt.Printf("O anime selecionado é uma série com %d episódios.\n", totalEpisodes)
		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)

		// A função extractEpisodeNumber não deve ser chamada para filmes/OVAs.
		// Este ajuste é específico para quando sabemos que é uma série com base na verificação anterior.
		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
		if err != nil {
			log.Fatalln("Error parsing episode number:", util.ErrorHandler(err))
		}

		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, animeURL, episodeNumberStr)
	} else {
		fmt.Println("O anime selecionado é um filme/OVA. Iniciando a reprodução direta...")
		// Para filmes/OVAs, utilizamos o primeiro episódio diretamente sem tentar parsear um número de episódio.
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		// Assume-se 1 como o número de episódio padrão para filmes/OVAs.
		player.HandleDownloadAndPlay(videoURL, episodes, 1, animeURL, episodes[0].Number)
	}
}
