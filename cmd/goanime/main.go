package main

import (
	"flag"
	"log"
	"time"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/playback"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/version"
	// Importa o pacote de notice para registrar avisos de tracking
)

func main() {
	startAll := time.Now()

	// Define all flags in one place
	versionFlag := flag.Bool("version", false, "show version information")
	debugFlag := flag.Bool("debug", false, "enable debug mode")
	helpFlag := flag.Bool("help", false, "show help message")
	altHelpFlag := flag.Bool("h", false, "show help message")

	flag.Parse()

	if *versionFlag || version.HasVersionArg() {
		version.ShowVersion()
		return
	}

	if *helpFlag || *altHelpFlag {
		util.Helper()
		return
	}

	// Set debug mode
	util.SetDebugMode(*debugFlag)
	if *debugFlag {
		log.Println("--- Debug mode is enabled ---")
	}

	animeName, err := util.GetAnimeName()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	tracking.HandleTrackingNotice()
	if util.IsDebug {
		log.Printf("[PERF] starting Goanime v%s", version.Version)
	}

	discordManager := discord.NewManager()
	if err := discordManager.Initialize(); err != nil {
		if util.IsDebug {
			log.Println("Failed to initialize Discord Rich Presence:", err)
		}
	} else {
		defer discordManager.Shutdown()
	}

	anime := appflow.SearchAnime(animeName)
	appflow.FetchAnimeDetails(anime)
	episodes := appflow.GetAnimeEpisodes(anime.URL)

	if util.IsDebug {
		log.Printf("[PERF] Full boot in %v", time.Since(startAll))
	}

	series, totalEpisodes := playback.CheckIfSeries(anime.URL)
	if series {
		playback.HandleSeries(anime, episodes, totalEpisodes, discordManager.IsEnabled())
	} else {
		playback.HandleMovie(anime, episodes, discordManager.IsEnabled())
	}
}
