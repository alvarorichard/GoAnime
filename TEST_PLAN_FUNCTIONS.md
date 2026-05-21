# GoAnime — Mapeamento das 165 Funções a 0% (Push 70%)

> **Estado em 2026-05-18 (pós-FASE 14):** 165 funções a 0% (de 12065 statements totais, 52.8% cobertura).
> **Gerado por:** `go tool cover -func=coverage.out | awk '$NF == "0.0%"'`
> **Regra absoluta (CLAUDE.md):** *cada* função desta lista recebe seu próprio `TestNomeDaFuncao_*`.
>
> Manifestos crus em `.test_manifests/p<N>_*.txt` (para grep/audit).

---

## Resumo por Fase

| Fase | Funcs | Foco | Refactor? |
|---|---:|---|:---:|
| 15 | 57 | `internal/api/`, `internal/util/` — branches + error paths | sim — vars injetáveis |
| 16 | 55 | `internal/playback/`, `internal/handlers/`, `internal/discord/`, `internal/upscaler/`, `internal/updater/` — TUI + IPC | sim — interface wrap + splits |
| 17 | 53 | `internal/scraper/`, `internal/api/providers/`, `internal/downloader/`, `pkg/goanime/...`, misc | parcial |
| **TOTAL** | **165** | | |

**Pós-FASE 17 projetado:** ≥ 70% cobertura, ≤ 50 funções a 0% (apenas TUI puro + `main()` + exemplos).

---

## FASE 15 — API + Util — Branches + Error Paths (57 funções)

### `internal/api/` — 27 funções

**allanime_enhanced.go**
- L14: `GetEpisodeStreamURLEnhanced`
- L75: `GetAllAnimeEpisodeURLDirect`

**allanime_smart.go**
- L21: `DownloadAllAnimeSmartRange`
- L81: `smartDownload`
- L286: `smartDownloadDirect`

**anime.go**
- L28: `GetEpisodeData`
- L33: `GetMovieData`
- L86: `FetchAnimeDetails`
- L116: `SearchAnime`
- L150: `FetchAnimeData`
- L209: `enrichAnimeData`
- L239: `searchAnimeOnPage`
- L288: `FetchAnimeFromAniList`
- L292: `selectAnimeWithGoFuzzyFinder`
- L322: `httpGetWithUA`

**enhanced.go**
- L50: `runWithSpinner`
- L90: `SearchAnimeEnhanced`
- L368: `GetEpisodeStreamURL`
- L481: `DownloadEpisodeEnhanced`
- L509: `DownloadEpisodeRangeEnhanced`
- L561: `downloadFromURL`
- L568: `SearchAnimeWithSource`
- L573: `GetAnimeEpisodesWithSource`
- L578: `GetSuperFlixEpisodes`
- L689: `GetSuperFlixStreamURL`

**episode_providers.go**
- L367: `getKitsuAnimeID`
- L441: `GetEpisodeDataWithFallback`
### `internal/util/` — 30 funções

**httpclient.go**
- L308: `PreWarmConnections`

**logger.go**
- L34: `PrintSavedLocation`
- L44: `getColoredPrefix`
- L59: `GetLogDir`
- L80: `initFileLogger`
- L139: `InitLogger`
- L199: `showDebugBanner`
- L234: `CloseLogFile`
- L263: `GetLogFileWriter`
- L273: `Debug`
- L286: `Info`
- L304: `Error`
- L313: `Fatal`
- L337: `Infof`
- L346: `Warnf`
- L355: `Errorf`

**util.go**
- L300: `RegisterCleanup`
- L307: `RunCleanup`
- L322: `ErrorHandler`
- L334: `Helper`
- L389: `FlagParser`
- L529: `getUserInput`
- L554: `TreatingAnimeName`
- L560: `handleDownloadModeWithSmart`
- L766: `handleUpscaleMode`
- L817: `handleMovieDownloadMode`
- L1116: `DefaultDownloadDir`
- L1127: `DefaultMovieDownloadDir`
- L1163: `FormatPlexMovieDir`
- L1192: `FormatPlexEpisodeDir`

---

## FASE 16 — Playback + Handlers + Discord + Upscaler + Updater (TUI/IPC com refactor) (55 funções)

### `internal/discord/` — 10 funções

**discord.go**
- L70: `LoginClient`
- L91: `LogoutClient`
- L114: `GetCurrentPlaybackPosition`
- L179: `Start`
- L207: `Stop`
- L221: `updateDiscordPresence`
- L344: `getPrecisePlaybackState`
- L397: `buildPreciseTimestamps`
- L511: `FetchDuration`

**init.go**
- L29: `Initialize`
### `internal/handlers/` — 12 funções

**download.go**
- L11: `HandleDownloadRequest`
- L26: `HandleMovieDownloadRequest`

**media.go**
- L48: `SearchMedia`
- L60: `SelectMediaType`
- L108: `GetAnimeStreamURL`
- L113: `InteractiveMediaFlow`
- L156: `handleAnimePlayback`

**playback.go**
- L19: `HandlePlaybackMode`

**update.go**
- L11: `HandleUpdateRequest`

**upscale.go**
- L15: `HandleUpscaleRequest`
- L68: `handleImageUpscale`
- L88: `handleVideoUpscale`
### `internal/playback/` — 14 funções

**common.go**
- L20: `PlayEpisode`

**input.go**
- L17: `GetUserInput`

**movie.go**
- L21: `HandleMovie`
- L172: `createUpdater`
- L190: `getSocketPath`

**series.go**
- L20: `printEpisodeNotFoundMsg`
- L24: `HandleSeries`
- L206: `SelectInitialEpisode`
- L224: `handleUserNavigation`
- L247: `handleUserNavigationEnhanced`
- L258: `handleAllAnimeNavigation`
- L302: `CheckIfSeries`
- L313: `CheckIfSeriesEnhanced`
- L323: `ChangeAnimeLocal`
### `internal/updater/` — 8 funções

**updater.go**
- L43: `CheckForUpdates`
- L86: `PerformUpdate`
- L289: `PromptForUpdate`
- L318: `CheckAndPromptUpdate`
- L343: `CheckForUpdatesQuietly`
- L413: `findAssetForPlatform`
- L544: `downloadAsset`
- L725: `createWindowsUpdateScript`
### `internal/upscaler/` — 11 funções

**anime4k.go**
- L401: `Close`

**shaders.go**
- L103: `InstallShaders`
- L161: `InstallGANShaders`
- L228: `extractZip`

**video.go**
- L277: `extractFrames`
- L297: `upscaleFrames`
- L396: `upscaleSingleFrame`
- L438: `encodeVideo`
- L512: `Init`
- L516: `Update`
- L538: `View`

---

## FASE 17 — Scraper + Providers + Downloader + SDK + Misc (53 funções)

### `cmd/goanime/` — 1 funções

**main.go**
- L18: `main`
### `internal/api/providers/` — 8 funções

**source_providers.go**
- L42: `FetchEpisodes`
- L51: `FetchStreamURL`
- L83: `FetchEpisodes`
- L91: `FetchStreamURL`
- L118: `FetchEpisodes`
- L126: `FetchStreamURL`
- L276: `FetchEpisodes`
- L284: `FetchStreamURL`
### `internal/api/providers/metadata/` — 1 funções

**metadata.go**
- L287: `LookupIMDBID`
### `internal/appflow/` — 2 funções

**anime_data.go**
- L48: `SearchAnimeWithRetry`
- L103: `FetchAnimeDetails`
### `internal/download/` — 2 funções

**workflow.go**
- L18: `HandleDownloadRequest`
- L147: `HandleMovieDownloadRequest`
### `internal/downloader/` — 5 funções

**downloader.go**
- L343: `downloadMultipleWithProgress`
- L710: `downloadWithProgress`
- L838: `downloadHTTPWithProgress`
- L941: `downloadM3U8WithYtDlp`
- L1171: `playEpisode`
### `internal/player/` — 1 funções

**playvideo.go**
- L828: `initDiscordPresence`
### `internal/scraper/` — 13 funções

**animefire.go**
- L199: `GetAnimeEpisodes`
- L268: `parseEpisodes`
- L518: `GetAnimeDetails`

**goyabu.go**
- L788: `sleep`

**media_manager.go**
- L65: `GetAnimeStreamURL`
- L91: `GetScraperManager`

**unified.go**
- L455: `SearchAnimePTBR`
- L558: `GetAnimeEpisodes`
- L580: `GetStreamURL`
- L622: `GetAnimeEpisodes`
- L626: `GetStreamURL`
- L646: `GetAnimeEpisodes`
- L650: `GetStreamURL`
### `internal/tracking/` — 1 funções

**notice.go**
- L8: `HandleTrackingNotice`
### `internal/tui/` — 3 funções

**find.go**
- L44: `Find`

**terminal.go**
- L13: `BubbleTeaProgramOptions`
- L23: `NewProgram`
### `pkg/goanime/` — 5 funções

**client.go**
- L25: `SearchAnime`
- L43: `GetAnimeEpisodes`
- L68: `GetStreamURL`
- L86: `DefaultStreamOptions`
- L105: `GetEpisodeStreamURL`
### `pkg/goanime/examples/episodes/` — 1 funções

**main.go**
- L12: `main`
### `pkg/goanime/examples/search/` — 1 funções

**main.go**
- L11: `main`
### `pkg/goanime/examples/source_specific/` — 1 funções

**main.go**
- L12: `main`
### `pkg/goanime/examples/stream/` — 1 funções

**main.go**
- L12: `main`
### `pkg/goanime/types/` — 7 funções

**anime.go**
- L97: `FromInternalAnime`
- L148: `FromInternalAnimeList`
- L157: `FromInternalEpisode`
- L199: `FromInternalEpisodeList`

**source.go**
- L20: `String`
- L32: `ToScraperType`
- L44: `ParseSource`

---


## Como Usar Este Arquivo

### Durante uma FASE

1. Abra a tabela da fase (ex: FASE 15).
2. Para cada arquivo `.go` listado, abra-o em paralelo com seu `*_test.go`.
3. Para cada função listada, escreva `func TestNomeDaFuncao_Cenario(t *testing.T)`.
4. Cobertura é verificada por package no final da fase.

### Verificação Pós-Fase

```bash
# Funções AINDA a 0% no pacote alvo (deve cair)
go test ./internal/<pkg>/ -coverprofile=cov.out -covermode=atomic
go tool cover -func=cov.out | awk '$NF == "0.0%"'

# Total geral
go test ./... -short -coverprofile=cov.out -covermode=atomic
go tool cover -func=cov.out | awk '$NF == "0.0%"' | wc -l
```

### Regeneração deste Arquivo

```bash
go test ./... -short -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | awk '$NF == "0.0%" {print $1, $2}' > /tmp/zero_funcs.txt
# Rerodar split em FASES (script em este histórico)
```

---

## Exceções Permanentes (NÃO testar)

Per CLAUDE.md, NÃO precisam de teste:

| Categoria | Localização | Funções |
|---|---|---|
| `main()` do CLI | `cmd/goanime/main.go:19` | 1 |
| `main()` de exemplos SDK | `pkg/goanime/examples/*/main.go` | 4 |
| Loops MPV/TUI interativos puros | `playback.HandleSeries`, `playback.HandleMovie`, `handlers.InteractiveMediaFlow`, `handlers.SelectMediaType` | ~10 |
| Bubble Tea `View()`/`tickCmd()` integradas | `player/helper.go`, etc. | ~5 |
| Funções com hardware (FFmpeg GPU/audio device) | `upscaler/anime4k.go:Close` | ~3 |
| `pkg/goanime/types/*` (defs sem lógica) | tipos puros | 7 |

**Total estimado de exceções:** ~30 funções → ficam a 0% mesmo após FASE 17.

**Funções a cobrir efetivamente:** 165 − 30 = ~135 (≈ 81% das funções 0% restantes).

---

## Métricas Esperadas

| Métrica | Pós-14 | Pós-15 | Pós-16 | Pós-17 |
|---|---:|---:|---:|---:|
| Funções 0% | 165 | ~108 | ~53 | **≤ 30** |
| Cobertura % | 52.8 | ~58 | ~64 | **≥ 70** |
| Testes novos | — | +57 | +55 | +53 |

