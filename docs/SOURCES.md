# Sources

GoAnime searches every enabled source in parallel. Sources are third-party
websites and change without notice, so one can stop working until it is
updated. A [daily workflow](../.github/workflows/source-health.yml) checks them.

## List

| Source        | Content    | Language            | Notes                                |
| ------------- | ---------- | ------------------- | ------------------------------------ |
| HiAnime       | Anime      | English             | Subbed by default, dub available     |
| AnimeFire     | Anime      | Portuguese (Brazil) |                                      |
| AnimesDigital | Anime      | Portuguese (Brazil) | Subtitled and dubbed                 |
| AnimeDrive    | Anime      | Portuguese (Brazil) |                                      |
| AnimesROLL    | Anime      | Portuguese (Brazil) | Subtitled and dubbed                 |
| AnimesFHD     | Anime      | Portuguese (Brazil) | Seasons, dub/sub, native MP4/HLS     |
| Goyabu        | Anime      | Portuguese (Brazil) | Off by default, needs cfsolve        |
| StartFlix     | Movies, TV | Portuguese (Brazil) | Dubbed and subtitled                 |
| TopCine       | Movies, TV | Portuguese (Brazil) | Catalog only, plays through StartFlix |
| Pobreflix     | Movies, TV | Portuguese (Brazil) | Some pages need cfsolve              |
| Pomfy         | Movies, TV | Portuguese (Brazil) | Search by name or TMDB id            |
| Pluto TV      | Movies, TV | By region (Brazil: mostly Portuguese) | Free, plays without ads            |

## Choosing a source

`--source` limits a search to one source (`hianime`, `animefire`, `goyabu`,
`animesdigital`, `anroll`, `animesfhd`, `animedrive`, `startflix`, `topcine`, `pobreflix`,
`pomfy`, `plutotv`). `movietv` covers the movie and TV sources, `ptbr` every Portuguese one,
`english` HiAnime and Pluto TV. Pluto TV's catalog has anime too, so it is searched on the
ALL, PT-BR, ENGLISH and MOVIES/TV tabs; `--source hianime` is still HiAnime alone. Pluto TV
says each title's language, so on the PT-BR tab its English titles are left out and on the
ENGLISH tab its Portuguese ones are; ALL and MOVIES/TV show every title.

Use
`GOANIME_ENABLED_SOURCES` / `GOANIME_DISABLED_SOURCES` to turn sources on or
off, and the `GOANIME_*_URL` variables to point one at a new domain when its
site moves.

## Notes

- **AnimeDrive**: the host is sometimes unreachable from some networks. Set
  `GOANIME_ANIMEDRIVE_URL` to a mirror, or search another source.
- **AnimesFHD** (`goanime --source animesfhd "frieren"`): provides Brazilian
  Portuguese subtitled and dubbed anime, including anime movies. Choose a
  season/audio version before the episode list. Only native MP4/HLS sources
  are supported. It uses its public JSON API without a key. Search returns
  up to 50 matches; episode lists follow pagination for the selected version.
  Embedded players are skipped. Native sources are tried in API order, with
  Referer and Portuguese captions passed to playback. The public SDK lists
  the exact series URL supplied by the caller; the season/audio menu belongs
  to the CLI.
- **Pobreflix**: plays public MP4s directly. Pages with a Turnstile-gated
  player need cfsolve; without it GoAnime falls back to the StartFlix panel.
- **Pluto TV** (`goanime --source plutotv "death note"`): the free service,
  read through the same keyless JSON the website uses; there is no account and
  no key. The catalog depends on the country you connect from, and a VPN
  changes it: some countries list nothing on demand, so a search that finds
  nothing there is not a title that does not exist. The debug log (`--debug`)
  names the session's country when a Pluto TV search comes back empty. Movies play at
  once; for a series you choose the season first. Its playlist cuts each title
  into pieces with ad breaks between them; GoAnime plays the title's own
  uninterrupted playlist, taken from the same place, so there are no breaks,
  and `-d`/`-dm` download that playlist. The playlist is handed to mpv from a
  local address (`http://127.0.0.1:<port>/plutotv/...`), because one of the
  service's CDNs refuses it at its own URL; the video itself still comes
  straight from the CDN, and seeking takes about a second. If a title cannot be read that way
  (for instance when it is made of several clips), the service's own playlist
  is played with its breaks rather than not at all. Titles protected by a DRM
  key system are refused with a message. Live channels are not offered.
  `--quality` picks the rung (`best`, `worst`, `720p`, `480p`...).
  Downloads: `goanime -dm --source plutotv "Movie"` for a movie,
  `goanime -dm --source plutotv --type tv "Show" 1 3` for season 1, episode 3
  (`-r` for a range, `-a` for every season). With `--source plutotv` only
  Pluto TV is searched; without it `-dm` searches every movie/TV source.
  `goanime -d --source plutotv "Show" 3` works too; it asks which season.
- **Pomfy**: looks titles up on TMDB (your `TMDB_API_KEY` if set, otherwise a
  public key). A bare TMDB id, such as `550`, also works as a search.
  `GOANIME_POMFY_URL` sets another mirror.
- **TopCine**: catalog only; playback goes through StartFlix.

## Cloudflare (cfsolve)

Goyabu and Pobreflix's player are behind Cloudflare. On networks that do not
challenge them they just work. Elsewhere they need
[cfsolve](https://github.com/chris124567/cfsolve), which release builds
download and verify on first run (Linux and Apple Silicon Macs). Windows and
Intel Macs have no build yet, so Goyabu only works there on unchallenged
networks (`GOANIME_ENABLED_SOURCES=goyabu`).

- `GOANIME_CFSOLVE_AUTODOWNLOAD=0` turns the download off.
- `GOANIME_CFSOLVE_BIN` points at your own binary.
- Developers: see [CFSOLVE.md](CFSOLVE.md).
