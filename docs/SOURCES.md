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
| Goyabu        | Anime      | Portuguese (Brazil) | Off by default, needs cfsolve        |
| StartFlix     | Movies, TV | Portuguese (Brazil) | Dubbed and subtitled                 |
| TopCine       | Movies, TV | Portuguese (Brazil) | Catalog only, plays through StartFlix |
| Pobreflix     | Movies, TV | Portuguese (Brazil) | Some pages need cfsolve              |
| Pomfy         | Movies, TV | Portuguese (Brazil) | Search by name or TMDB id            |

## Choosing a source

`--source` limits a search to one source (`hianime`, `animefire`, `goyabu`,
`animesdigital`, `anroll`, `animedrive`, `startflix`, `topcine`, `pobreflix`,
`pomfy`). `movietv` covers the movie and TV sources, `ptbr` every Portuguese one.

Use
`GOANIME_ENABLED_SOURCES` / `GOANIME_DISABLED_SOURCES` to turn sources on or
off, and the `GOANIME_*_URL` variables to point one at a new domain when its
site moves.

## Notes

- **AnimeDrive**: the host is sometimes unreachable from some networks. Set
  `GOANIME_ANIMEDRIVE_URL` to a mirror, or search another source.
- **Pobreflix**: plays public MP4s directly. Pages with a Turnstile-gated
  player need cfsolve; without it GoAnime falls back to the StartFlix panel.
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
