<p align="center">
  <b>English</b> | <a href="README_pt-br.md">Português</a>
</p>

<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.png">
    <img src="docs/assets/logo-light.png" alt="GoAnime logo" width="160">
  </picture>
</p>

<h1 align="center">GoAnime</h1>

<p align="center">
  <a href="https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml"><img src="https://github.com/alvarorichard/GoAnime/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/alvarorichard/GoAnime/releases/latest"><img src="https://img.shields.io/github/v/release/alvarorichard/GoAnime" alt="Latest release"></a>
  <a href="https://pkg.go.dev/github.com/alvarorichard/Goanime/pkg/goanime"><img src="https://pkg.go.dev/badge/github.com/alvarorichard/Goanime/pkg/goanime.svg" alt="Go Reference"></a>
  <a href="https://aur.archlinux.org/packages/goanime"><img src="https://img.shields.io/aur/version/goanime" alt="AUR"></a>
  <a href="LICENSE"><img src="https://img.shields.io/github/license/alvarorichard/GoAnime" alt="License"></a>
</p>

GoAnime is a terminal application for searching, streaming and downloading
anime, movies and TV shows in English and Brazilian Portuguese. It searches
several sources at once, plays the selected episode in
[mpv](https://mpv.io/), and keeps track of where you stopped.

<p align="center">
  <img src="docs/demo/demo.gif" alt="Searching for a title in GoAnime, picking it from the results and streaming it" width="800">
</p>

It is a single binary for Linux, macOS and Windows, and mpv is all it needs to
stream. [FFmpeg](https://ffmpeg.org/) is optional: downloads use it to embed
subtitles, and upscaling a video requires it.

## Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Sources](#sources)
- [Configuration](#configuration)
- [Files](#files)
- [Troubleshooting](#troubleshooting)
- [Using GoAnime as a library](#using-goanime-as-a-library)
- [Contributing](#contributing)
- [Community](#community)
- [Acknowledgments](#acknowledgments)
- [Disclaimer](#disclaimer)
- [License](#license)

## Features

- Searches every enabled source in parallel and falls back to another source
  when one fails.
- Subbed and dubbed anime in English and Portuguese, plus Portuguese movies
  and TV shows.
- Streams with quality selection, or downloads a single episode, a range, or
  a whole series.
- Saves downloads with Plex-compatible names (`Anime - S01E01.mp4`).
- Resumes playback and remembers watched episodes.
- Skips the opening from the player menu, using timestamps from AniSkip.
- Discord Rich Presence.
- Anime4K upscaling in real time while you watch, or of video and image files.
- Updates itself with `goanime --update`.

## Installation

GoAnime uses [mpv](https://mpv.io/) for playback. Install mpv first, unless
you use the Windows installer, which includes it.

Prebuilt binaries for every platform are available on the
[releases page](https://github.com/alvarorichard/GoAnime/releases/latest).

### Windows

Download and run `GoAnime-Installer-<version>.exe` from the
[latest release](https://github.com/alvarorichard/GoAnime/releases/latest).
It installs GoAnime and mpv and adds both to your `PATH`.

If you prefer not to use the installer, download
`goanime-windows-amd64.zip`, extract `goanime-windows-amd64.exe`, rename it to
`goanime.exe` and put it in a folder on your `PATH`. The archive does not
include mpv: [install it](https://mpv.io/installation/) separately. GoAnime
finds `mpv.exe` on your `PATH` or in the same folder as `goanime.exe`.

### macOS

```bash
brew install mpv
curl -Lo goanime https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-darwin-universal
chmod +x goanime
sudo mv goanime /usr/local/bin/
```

The binary is universal and runs on both Apple Silicon and Intel Macs. If you
download it with a browser instead of `curl`, macOS quarantines it and refuses
to run it; clear the flag with
`xattr -d com.apple.quarantine /usr/local/bin/goanime`.

### Arch Linux

GoAnime is available in the [AUR](https://aur.archlinux.org/packages/goanime):

```bash
yay -S goanime
```

### Debian, Ubuntu, Fedora and other distributions

Install mpv with your package manager (`sudo apt install mpv` or
`sudo dnf install mpv`), then:

```bash
curl -LO https://github.com/alvarorichard/GoAnime/releases/latest/download/goanime-linux-amd64.tar.gz
tar -xzf goanime-linux-amd64.tar.gz
sudo install goanime-linux-amd64 /usr/local/bin/goanime
```

For ARM64, replace `amd64` with `arm64`.

### From source

With Go 1.27.3 or later:

```bash
go install github.com/alvarorichard/Goanime/cmd/goanime@latest
```

Progress tracking is backed by SQLite and requires CGO, so a C compiler must
be available when building. Without one, GoAnime still builds and runs, but
does not track progress. The release binaries are built with SQLite. See
[docs/BUILD_OPTIONS.md](docs/BUILD_OPTIONS.md) for details.

### Verifying a download

Every release includes `checksums-sha256.txt`. Compute the hash of the file
you downloaded and compare it with the line for that file:

```bash
sha256sum goanime-linux-amd64.tar.gz   # Linux
shasum -a 256 goanime                  # macOS, the goanime-darwin-universal line
Get-FileHash goanime-windows-amd64.zip # Windows, in PowerShell
```

### Updating

| Installed with              | Update with                                       |
| --------------------------- | ------------------------------------------------- |
| Windows installer           | The installer of the new version                  |
| A release binary or archive | `goanime --update`                                |
| AUR                         | Your AUR helper, for example `yay -Syu`           |
| `go install`                | The same `go install` command                     |

`goanime --update` checks the download against the release's
`checksums-sha256.txt` and installs nothing if they do not match. It replaces
the executable in place, so if it lives in a folder you cannot write to, such
as `/usr/local/bin`, run `sudo goanime --update`.

## Usage

Run `goanime` without arguments to open search home. Type a name and press
enter, or press `↓` to pick from the lists below: Continue reopens an
episode where you stopped (`enter` resumes, `r` starts over, `e` lists
episodes, `x` removes it) and Recent repeats a past search with its result
count. `tab` switches scope between ALL, PT-BR, ENGLISH and MOVIES/TV.
From a result, choose a version with `←` `→` and, for a series, an episode.
You can also pass the name directly:

```bash
goanime "one piece"
```

Use spaces in the name, not hyphens.

Anime downloads go to `~/.local/goanime/downloads/anime/` unless you pass
`-o`:

```bash
goanime -d "one piece" 1                  # episode 1
goanime -d -r "naruto" 1-12               # episodes 1 to 12
goanime -d -a "one piece"                 # every episode
goanime -d --quality 720p "frieren" 1     # a specific quality
goanime -d -o ~/Anime "bleach" 10         # a custom directory
```

The directory given to `-o` must be inside your home folder.

Movies and TV shows use `-dm` and go to `~/.local/goanime/downloads/movies/`:

```bash
goanime -dm "Interstellar"                # a movie
goanime -dm --type tv "Dark" 1 5          # season 1, episode 5
goanime -dm -r "Dark" 1 1-5               # season 1, episodes 1 to 5
goanime -dm -a "Dark"                     # every season
```

Some streams are downloaded with [yt-dlp](https://github.com/yt-dlp/yt-dlp),
which GoAnime fetches on its own the first time it needs it.

Other options:

```bash
goanime --source animefire "jujutsu kaisen"   # search a single source
goanime --upscale --upscale-hq video.mp4      # upscale a file with Anime4K
goanime --update                              # update to the latest release
goanime --debug "naruto"                      # write a debug log
```

Run `goanime -h` for the full list of options.

### Keys

| Key                   | Action                    |
| --------------------- | ------------------------- |
| `↑` `↓` or `k` `j`    | Move                      |
| Type, or `/`          | Filter the list           |
| `g` `G`               | First or last item        |
| `Enter`               | Select                    |
| `Esc`                 | Go back                   |
| `tab`                 | Switch search scope (search home) |
| `←` `→`               | Choose a version (results) |
| `r` `e` `x`           | From start, episodes, remove (Continue list) |
| `d` `D` `A` `m`       | Download one, range, all, options (episodes) |

In a menu that shows a letter beside each option, that letter picks the
option at once; any other key still filters.

The episode list marks the episodes you finished with ✓ and the one you
stopped partway with ◐, and opens where you left off.

### While watching

While mpv plays, the terminal shows the player menu, with how far the episode
has played:

| Key   | Option                           | What it does                                           |
| ----- | -------------------------------- | ------------------------------------------------------ |
| `n`   | Next episode                     | Play the next episode                                  |
| `p`   | Previous episode                 | Play the previous episode                              |
| `e`   | Select episode                   | Pick another episode from the list                     |
| `c`   | Change anime                     | Search for another title                               |
| `s`   | Skip intro                       | Jump past the opening, when AniSkip has its timestamps |
| `a`   | Audio track                      | Switch the audio track, when there is more than one    |
| `t`   | Subtitles                        | Switch or turn off subtitles, when there are any       |
| `Esc` | ← Back                           | Return to the download and playback menu               |
| `q`   | Exit                             | Close mpv and quit                                     |

For a movie, the menu offers Replay movie (`r`), Change movie (`c`), the track
options, ← Back and Exit. When an episode or movie has saved progress, GoAnime
asks whether to resume from where you stopped.

### Real-time upscaling

Real-time Upscale, in the menu before playback, applies
[Anime4K](https://github.com/bloc97/Anime4K) shaders in mpv. The first time,
choose Setup shaders to download them, then pick a mode. Each mode's label
says what it is for, from Performance for a weak GPU to Ultra for SD sources.

## Sources

| Source        | Content    | Language                   | Notes                                         |
| ------------- | ---------- | -------------------------- | --------------------------------------------- |
| HiAnime       | Anime      | English, subbed and dubbed | Subbed by default                             |
| AnimeFire     | Anime      | Portuguese (Brazil)        |                                               |
| AnimesDigital | Anime      | Portuguese (Brazil)        | Subtitled and dubbed                          |
| Goyabu        | Anime      | Portuguese (Brazil)        | Disabled by default, see below                |
| StartFlix     | Movies, TV | Portuguese (Brazil)        | Dubbed and subtitled                          |
| TopCine       | Movies, TV | Portuguese (Brazil)        | Catalog only; playback goes through StartFlix |

`--source` limits a search to one source: `hianime`, `animefire`,
`animesdigital`, `startflix` or `topcine`, or `ptbr` for all the Portuguese
sources.

Goyabu is behind a Cloudflare challenge and is disabled by default. It still
works on networks that Cloudflare does not challenge; enable it with
`GOANIME_ENABLED_SOURCES=goyabu`.

GoAnime depends on third-party websites, and those websites change without
notice. When one does, the corresponding source can stop working until it is
updated. A [scheduled workflow](.github/workflows/source-health.yml) checks
every source daily, so breakage is usually visible there before it is
reported.

## Configuration

GoAnime needs no configuration file. A few environment variables cover less
common needs:

| Variable                      | Description                                                              |
| ----------------------------- | ------------------------------------------------------------------------ |
| `GOANIME_ENABLED_SOURCES`     | Comma-separated sources to enable, such as `goyabu`                      |
| `GOANIME_DISABLED_SOURCES`    | Comma-separated sources to skip                                          |
| `GOANIME_THEME`               | `dark`, `light` or `auto` (default): `auto` asks the terminal once for its background |
| `GOANIME_PLAIN`               | Set to `1` for the plain interface: no colors, no Japanese glosses, ASCII only (same as `--plain`, `NO_COLOR` or `TERM=dumb`) |
| `GOANIME_HIANIME_AUDIO`       | `sub` (default) or `dub`                                                 |
| `GOANIME_STRICT_SOURCE`       | Set to `1` to report an error instead of guessing an unknown source      |
| `GOANIME_ANIMEFIRE_API`       | Address of the AnimeFire API                                             |
| `GOANIME_STARTFLIX_URL`       | Address of StartFlix                                                     |
| `GOANIME_STARTFLIX_PANEL_URL` | Address of the StartFlix video panel                                     |
| `GOANIME_TOPCINE_URL`         | Address of TopCine                                                       |
| `TMDB_API_KEY`                | A [TMDB](https://www.themoviedb.org/settings/api) key for movie, TV and season details |
| `OMDB_API_KEY`                | An [OMDb](https://www.omdbapi.com/apikey.aspx) key to use instead of the shared demo key |

The address variables point a source at a new domain when its site moves,
without waiting for a new release.

## Files

| What            | Linux                                   | macOS                                                     | Windows                                       |
| --------------- | --------------------------------------- | --------------------------------------------------------- | --------------------------------------------- |
| Watch history   | `~/.local/goanime/tracking/progress.db` | `~/.local/goanime/tracking/progress.db`                   | `%LOCALAPPDATA%\GoAnime\tracking\progress.db` |
| Downloads       | `~/.local/goanime/downloads/`           | `~/.local/goanime/downloads/`                             | `%USERPROFILE%\.local\goanime\downloads\`     |
| Debug logs      | `~/.local/share/goanime/logs/`          | `~/Library/Logs/GoAnime/`                                 | `%LOCALAPPDATA%\GoAnime\logs\`                 |
| Anime4K shaders | `~/.config/goanime/anime4k-shaders/`    | `~/Library/Application Support/goanime/anime4k-shaders/`  | `%APPDATA%\goanime\anime4k-shaders\`           |
| yt-dlp          | `~/.cache/go-ytdlp/`                    | `~/Library/Caches/go-ytdlp/`                              | `%LOCALAPPDATA%\go-ytdlp\`                      |

On Linux, the shader and yt-dlp folders follow `XDG_CONFIG_HOME` and
`XDG_CACHE_HOME` when they are set. Deleting the watch history file resets
all progress. To uninstall GoAnime, remove the executable (or, with the
Windows installer, uninstall it from Settings) and delete the folders above
that you no longer need.

## Troubleshooting

### AnimeDrive search times out

AnimeDrive can take longer to answer when the site is busy. Its search uses
the full source deadline, without restarting a request at the shorter HTTP
client timeout. If it still times out, try again later or use `--source` to
search another source, and include a debug log when reporting the problem.

### Reporting a problem

Open a
[bug report](https://github.com/alvarorichard/GoAnime/issues/new?template=bug_report.yml)
and attach a debug log to it:

1. Run the same command again with `--debug` added, for example
   `goanime --debug "naruto"`, and repeat what you did until the problem
   happens.
2. When it starts, GoAnime prints the path of the log file. Each run writes a
   new file named `goanime_<date>_<time>.log` in this folder:

   | System  | Folder                         |
   | ------- | ------------------------------ |
   | Windows | `%LOCALAPPDATA%\GoAnime\logs`  |
   | macOS   | `~/Library/Logs/GoAnime`       |
   | Linux   | `~/.local/share/goanime/logs`  |

3. Attach the newest file to the issue, together with the output of
   `goanime --version`, your operating system and the command you ran.

The log lists the titles you searched and paths on your computer, which can
include your user name. Read it before attaching if that matters to you.

### mpv is not found

Make sure `mpv --version` works in the same terminal you run GoAnime from.
On Windows, open a new terminal after installing so the updated `PATH` is
picked up.

## Using GoAnime as a library

The search and scraping code is exposed as a public Go package:

```go
client := goanime.NewClient()
results, err := client.SearchAnime("One Piece", nil)
```

Search covers every source. Episode lists and stream URLs are available for
AnimeFire results only for now. [docs/LIBRARY_USAGE.md](docs/LIBRARY_USAGE.md)
walks through the API, and [pkg/goanime/examples](pkg/goanime/examples) has
complete programs.

## Contributing

Bug reports, fixes and new sources are welcome.

Most of the maintenance work in GoAnime is keeping sources working as the
sites behind them change, and that is also the easiest way to start
contributing: the [source health workflow](https://github.com/alvarorichard/GoAnime/actions/workflows/source-health.yml)
shows which sources are failing, and each source is a self-contained package
under `internal/scraper/providers`. To add a new one, follow
[docs/ADDING_A_SOURCE.md](docs/ADDING_A_SOURCE.md), which uses Goyabu as the
reference implementation.

To build and test:

```bash
git clone https://github.com/alvarorichard/GoAnime.git
cd GoAnime
git checkout dev
go build ./cmd/goanime
go test -short ./...
```

Development happens on the `dev` branch. Create your branch from `dev` and
open pull requests against `dev`; `main` only receives releases.
[CONTRIBUTING.md](CONTRIBUTING.md) lists the checks to run before opening a
pull request, and [docs/Development.md](docs/Development.md) covers code style
and the rest of the workflow.

## Community

- [Discord](https://discord.gg/FbQuf78D9G) for questions, feedback and release
  announcements.
- [GoAnime Mobile](https://github.com/alvarorichard/goanime-mobile), the
  Android version.
- [goanime.online](https://goanime.online/), the project website, with an
  overview of GoAnime and installation instructions.

## Acknowledgments

GoAnime is built on the work of these projects:

- [mpv](https://mpv.io/) plays every stream.
- [AniSkip](https://github.com/aniskip) provides the opening timestamps.
- [AniList](https://anilist.co/) and [Jikan](https://jikan.moe/), an
  unofficial MyAnimeList API, provide anime metadata.
- [TMDB](https://www.themoviedb.org/) and [OMDb](https://www.omdbapi.com/)
  provide movie and TV metadata.
- [Anime4K](https://github.com/bloc97/Anime4K) provides the upscaling
  algorithm and shaders.
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) handles some downloads.
- [Bubble Tea](https://github.com/charmbracelet/bubbletea),
  [Huh](https://github.com/charmbracelet/huh) and
  [Lip Gloss](https://github.com/charmbracelet/lipgloss) build the terminal
  interface.
- [ani-cli](https://github.com/pystardust/ani-cli) inspired the project.

## Disclaimer

GoAnime does not host, upload or distribute any media. It retrieves content
that third-party websites make publicly available and is not affiliated with
any of them. Users are responsible for complying with the laws of their
country.

## License

GoAnime is released under the [MIT License](LICENSE).
