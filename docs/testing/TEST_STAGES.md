# GoAnime — Plano de Execução por Fases

> **Meta:** 25.9% → 70% · **1 teste por função** · **~983 testes** (excluindo ~50 não-testáveis)
> **Referência:** Funções alvo em `TEST_PLAN_FUNCTIONS.md` · Estratégia em `TEST_STRATEGY.md`

---

## FASE 1 ✅ — Lógica Pura Simples (~50 funções)
**Pacotes:** `models`, `version`, `pkg/goanime/types`, `api/source`, `api/aniskip`, `api/series`, `api/anime_url_title`

| Pacote | Arquivo | Funções | Tipo |
|---|---|---|---|
| `internal/models/` | `media.go` | 13 (IsAnime, IsMovie, IsTV, IsMovieOrTV, GetDisplayName, OfficialTitle, GetRatingDisplay, GetGenresDisplay, GetRuntimeDisplay, etc.) | Puro |
| `internal/models/` | `tmdb.go` | 4 (GetDisplayTitle, GetReleaseYear, GetPosterURL, GetBackdropURL) | Puro |
| `internal/version/` | `version.go` | 2 (HasVersionArg, ShowVersion) | Puro |
| `pkg/goanime/types/` | `anime.go`, `source.go` | 7 (FromInternalAnime, FromInternalAnimeList, FromInternalEpisode, FromInternalEpisodeList, String, ToScraperType, ParseSource) | Puro |
| `internal/api/source/` | `definition.go`, `kind.go`, `resolve.go` | 7 (matchNonExplicit, ScraperTypeFor, ExtractAllAnimeID, Resolve, ResolveURL, BestEffortKind, IsAllAnimeShortID) | Puro/Mock |
| `internal/api/` | `aniskip.go` | 4 (GetAniSkipData, RoundTime, ParseAniSkipResponse, GetAndParseAniSkipData) | Puro + httptest |
| `internal/api/` | `series.go`, `anime_url_title.go` | 4 (IsSeries, IsSeriesEnhanced, toTitleCase, FetchAnimeFromAniListWithURL) | Puro/Mock |

**Verificação:** `go test ./internal/models/ ./internal/version/ ./pkg/goanime/types/ ./internal/api/source/ -v -race`

---

## FASE 2 ✅ — API Pura (~45 funções)
**Pacotes:** `api/anime.go`, `api/episodes.go`, `api/enhanced.go` (funções puras), `api/allanime_smart.go`

| Arquivo | Funções | Tipo |
|---|---|---|
| `api/anime.go` | ~20 (GetEpisodeData, GetMovieData, FetchAnimeDetails, SearchAnime, FetchAnimeData, getStringValue, getIntValue, getBoolValue, enrichAnimeData, searchAnimeOnPage, ParseAnimes, FetchAnimeFromAniList, httpGetWithUA, httpPostFast, resolveURL, normalizeAccents, CleanTitle, safeClose, selectAnimeWithGoFuzzyFinder) | Puro + httptest |
| `api/episodes.go` | 4 (GetAnimeEpisodes, parseEpisodes, parseEpisodeNumber, sortEpisodesByNum) | Puro |
| `api/enhanced.go` | 4 puras (sanitizeFilename, extractMediaIDFromURL, languagePriority, isStdoutTerminal) | Puro |
| `api/allanime_smart.go` | 13 (sanitizeSmart, sanitizeSmartDest, validateSmartRangeInputs, shouldUseYtDlp, isUnsafeExtensionError, alreadyDownloaded, smartOutputDir, smartDownload, DownloadAllAnimeSmartRange, writeAniSkipSidecar, WriteAniSkipSidecar, smartDownloadDirect, resolveStreamURLForEpisode) | Puro/Mock |

**Verificação:** `go test ./internal/api/ -run "TestAnime|TestEpisode|TestSanitize|TestExtractMedia|TestLanguage|TestSmart|TestValidate" -v -race`

---

## FASE 3 ✅ — Segurança SSRF + Player Puro (~40 funções)
**Pacotes:** `api/api.go`, `scraper/ssrf.go`, `api/movie/ssrf.go`, `player/` (funções puras)

| Arquivo | Funções | Tipo |
|---|---|---|
| `api/api.go` | 7 (IsDisallowedIP, checkDisallowedIP, dialFunc, SafeTransport, SafeGet, ValidateExternalURL, SafeDialContext) | Puro + Mock |
| `scraper/ssrf.go` | 3 (isDisallowedIP, safeDialFunc, safeScraperTransport) | Puro |
| `api/movie/ssrf.go` | 3 (isDisallowedIP, safeDialFunc, safeMovieTransport) | Puro |
| `player/player.go` | ~15 puras (filterMPVArgs, sanitizeMediaTarget, sanitizeOutputPath, buildMPVCommand, IsCurrentMediaMovie, SetAnimeName, SetMediaType, SetExactMediaType, GetExactMediaType, snapshotMedia, SetSeasonMap, SetMediaMeta, GetMediaMeta, PreWarmMPVPath, taskTotal, shouldGrowProgressTotal) | Puro |
| `player/download.go` | ~8 puras (LooksLikeHLS, hasUnsafeExtension, isBloggerProxyURL, isAnimeFireVideoAPIURL, isUnsafeExtensionError, isRetryableError, extractRefererFromURL, fileExists) | Puro |
| `player/scraper.go` | ~5 puras (extractResolution, abs, isPlayableVideoURL, needsVideoExtraction, isNumericString, isLikelyAllAnimeID, DownloadFolderFormatter) | Puro |

**Verificação:** `go test ./internal/api/ -run "TestIsDisallowed|TestValidate|TestSafe" -v -race && go test ./internal/player/ -run "TestFilter|TestSanitize|TestBuild|TestLooks|TestHas|TestIs" -v -race`

---

## FASE 4 ✅ — Scraper Infraestrutura (~45 funções)
**Pacotes:** `scraper/source_diagnostic.go`, `scraper/source_circuit.go`, `scraper/source_health.go`, `scraper/errors.go`, `scraper/unified.go` (helpers puros)

| Arquivo | Funções | Tipo |
|---|---|---|
| `source_diagnostic.go` | 14 (Is, UserMessage, Error, DiagnoseError, containsAny, NewHTTPStatusError, NewBlockedChallengeError, NewParserError, NewDecryptError, NewDownloadExpiredError, NewInternalBugError, isNetworkUnavailable, isBlockedStatus, statusFromMessage, etc.) | Puro |
| `source_circuit.go` | 7 (newSourceCircuitBreaker, recordSuccess, recordFailure, ensureCircuitBreaker, circuitOpenDiagnostic, recordSourceSuccess, recordSourceFailure) | Estado |
| `source_health.go` | 3 (CheckAllSourcesHealth, DefaultHealthCheckQuery, AvailableSources) | Mock |
| `errors.go` | 2 (checkHTTPStatus, checkHTMLResponse) | Puro |
| `unified.go` helpers | ~15 (sortPTBRFirst, cleanPTBRTitle, SearchAnimePTBR, getScraperDisplayName, getLanguageTag, NewScraperManager, PreWarmScraperManager, SearchAnime) | Puro/Mock |

**Verificação:** `go test ./internal/scraper/ -run "TestDiagnose|TestCircuit|TestHealth|TestCheck|TestSortPTBR|TestCleanPTBR|TestGetScraperDisplay|TestGetLanguage" -v -race -count=3`

---

## FASE 5 ✅ — Unified Adapters (~45 funções)
**Pacotes:** `scraper/unified.go` (adapters ativos: AnimeFire, Goyabu, AllAnime, SuperFlix)

Cada adapter tem ~4-5 métodos (SearchAnime, GetAnimeEpisodes, GetStreamURL, GetType, GetClient). Total ~40 métodos de adapters + NewSuperFlixAdapterWithClient.

**Tipo:** Unit + MockScraper (reutilizar `MockScraper` de `unified_test.go`)

**Verificação:** `go test ./internal/scraper/ -run "TestAdapter|TestSuperFlixAdapter" -v -race`

---

## FASE 6 ✅ — Util Completo (~83 funções)
**Pacotes:** `util/util.go`, `util/httpclient.go`, `util/perf.go`, `util/logger.go`, `util/help.go`, `util/ytdlp.go`

| Arquivo | Funções | Tipo |
|---|---|---|
| `util.go` | ~25 (SetGlobalSubtitles, ClearGlobalSubtitles, SetGlobalReferer, GetGlobalReferer, ClearGlobalReferer, SetGlobalAnimeSource, GetGlobalAnimeSource, Is9AnimeSource, TreatingAnimeName, stripTrailingAnimeMetadata, BuildMediaFolderName, BuildMediaFileName, DefaultDownloadDir, DefaultMovieDownloadDir, FormatPlexMovieDir, FormatPlexEpisodePath, FormatPlexEpisodeDir, RegisterCleanup, RunCleanup, ErrorHandler, Helper, FlagParser) | Puro |
| `httpclient.go` | ~22 (NewResponseCache, Get, Set, cleanupLoop, cleanup, GetAniListCache, GetSearchCache, NewWorkerPool, Submit, Wait, GetScraperPool, GetAPIPool, ParallelExecute, newSurfStdClient, GetSharedClient, GetFastClient, NewFastClient, GetDownloadClient, PreWarmClients) | Estado/Stress |
| `perf.go` | ~16 (GetPerfTracker, StartTimer, Stop, StopAndLog, Record, IncrementCounter, GetCounter, GetMetrics, GetUptime, Reset, PrintReport, TimeFunc, TimeFuncWithResult, TimeFuncWithError, Perf, PerfCount) | Estado |
| `logger.go` | ~16 (PrintSavedLocation, getColoredPrefix, GetLogDir, initFileLogger, InitLogger, showDebugBanner, CloseLogFile, GetLogFileWriter, Debug, Info, Warn, Error, Fatal, Infof, Warnf, Errorf) | Puro |
| `help.go` | 4 (ShowBeautifulHelp, addOption, addFeature, addExample) | Puro |
| `ytdlp.go` | 1 (YtdlpCanImpersonate) | Puro |

**Verificação:** `go test ./internal/util/ -v -race -count=3`

---

## FASE 7 ❌ — REMOVIDA (FlixHQ deletado)
> FlixHQ scraper foi removido em 2026-05-17 — site caiu.

---

## FASE 8 ❌ — REMOVIDA (SFlix deletado)
> SFlix scraper foi removido em 2026-05-17 — mesma queda que FlixHQ. Arquivos deletados: `internal/scraper/sflix.go`, `sflix_test.go`, `internal/scraper/movie/sflix.go`.

---

## FASE 9 ✅ — AnimeFire + Goyabu + AllAnime (~29 funções)
**Arquivos:** `animefire.go`(8), `goyabu.go`(7), `allanime.go`(14)

NineAnime (9animetv.to) removido em 2026-05-17 — site caiu. Restantes scrapers cobertos com `httptest.Server`. Cada um tem SearchAnime, GetEpisodes, GetStreamURL + helpers internos.

**Verificação:** `go test ./internal/scraper/ -run "TestAnimeFire|TestGoyabu|TestAllAnime" -v -race`

---

## FASE 10 ✅ — SuperFlix + MediaManager (~69 funções)
**Arquivos:** `superflix.go`(9), `media_manager.go`(60)

AnimeDrive removido em 2026-05-17. MediaManager agora anime-only.

**Verificação:** `go test ./internal/scraper/ -run "TestSuperFlix|TestMediaManager" -v -race`

---

## FASE 11 ✅ — Player Completo (~128 funções)
**Arquivos:** `player.go`(40), `playvideo.go`(~35), `download.go`(~28), `scraper.go`(~25)

Funções com MPV (StartVideo, mpvSendCommand) → mock com `net.Pipe()` para IPC socket.
Funções puras (filter, sanitize, extract) → unitário direto.
Funções TUI (askForDownload) → skip ou testar lógica interna.

**Verificação:** `go test ./internal/player/ -v -race`

**Sessão completa** — 1 teste dedicado por função (CLAUDE.md regra #1). Total: **312+ testes** no pacote, **128/128 funções** cobertas. Cobertura `internal/player`: 22.3% → **51.4%** (ceiling sem refatoração de produção — `api.SafeGet`/`api.SafeTransport` bloqueia loopback IPs, impedindo `httptest.Server` de exercitar as funções pesadas de rede: `DownloadVideo`, `extractVideoURL`, `fetchContent`, `extractActualVideoURL` animefire, `ExtractVideoSources`, `downloadBloggerDirect/Chunk`. Para passar de 60% seria necessário ou expor um hook de injeção de cliente em `internal/api` ou rodar testes contra um IP público mockado).

Distribuição por arquivo (mantém padrão do repo `<source>_test.go` / `Test<Funcao>_<Cenario>`):

| Arquivo | Adicionado |
|---|---|
| `progress_aggregation_test.go` | `taskTotal`, `shouldGrowProgressTotal`, `setProgressPeak`×3, `childProgress`×2, `setTaskTotal`, `setProgressTotal`×2, `progressTotal`, `addProgressReceived`, `addTaskReceived`, `setProgressReceived`, `setTaskReceived`, `resetProgressReceived`, `resetTaskReceived`×2 |
| `player_ipc_test.go` (novo) | Helper `startMockMPVSocket` (unix socket) + IPC: `mpvSendCommand`×4, `MpvSendCommand`, `dialMPVSocket`×2, `ToggleSubtitle`, `SetPlaybackSpeed`, `CycleAudio/SubtitleTrack`, `SetAudio/SubtitleTrack`, `GetPlaybackStats`, `GetAudio/SubtitleTracks` (+ bad shape), `GetCurrentAudio/SubtitleTrack` (+ tipos inesperados) |
| `playvideo_pure_test.go` (novo) | `applySkipTimes`×2, `findEpisodeIndex`×2, `trackingKey`, `getTrackerDBPath`, `getCurrentEpisode`×2, `getEpisodeTitle`, `initTracking`, `InitTrackerAsync`, `updateTrackingWithDuration`, `fetchAniSkipAsync`, `showShaderOSD`, `applyAniSkipResults`, `waitForVideoReady`, `seekToResumePosition`×2, `waitForPlaybackStart`, `updateEpisodeDuration`, `updateTracking`, `preloadNextEpisode`×2, `startTrackingRoutine`, `skipIntro`×2, `selectAudioTrack`×2, `selectSubtitleTrack`, `showPlayerMenu`, `showResumeDialog`, `handleUserInput`, `playNextEpisode`, `playPreviousEpisode`, `selectEpisode`, `switchEpisode`, `playVideo`, `initDiscordPresence` (symbol-pin) |
| `scraper_pure_test.go` | `estimateContentLengthForAllAnime`×5, `extractActualVideoURL`, `isMovieOrTVSourcePlayer`, `GetBloggerVideoURL`, `StopBloggerProxy`, `getBloggerSessionClient`, `newSurfClient`, `newSurfDownloadClient`, `SelectEpisodeWithFuzzyFinder`, `GetVideoURLForEpisode`, `GetVideoURLForEpisodeEnhanced`, `extractVideoURL` (SSRF), `fetchContent` (SSRF), `extractBloggerVideoURL`, `startBloggerProxy`, `selectQualityFromOptions`×5, `needsVideoExtraction` |
| `player_pure_test.go` | `setLastAnimeURL`/`getLastAnimeURL`, `GetExactMediaType`, `GetMediaMeta`, `downloadSubtitleFiles`, `printDownloadLocation`, `StartVideo`, `handleUpscaleFromMenu`, `askForDownload`, `askForPlayOffline`, `HandleDownloadAndPlay`/`downloadAndPlayEpisode` (symbol-pin — TUI loop não driveable sem TTY) |
| `download_pure_test.go` | `combineParts`×2, `createEpisodePath`, `findEpisode`, `resolveDownloadURL`×2, `resolveAnimeFireFallbackDownloadURL`, `selectAnimeFireDownloadCandidates`×3, `selectAnimeFireDownloadSource`, `orderAnimeFireSources`×3, `recordBatchDownloadFailure`×2, `newBatchDownloadError`×2, `batchDownloadError.Error`×3, `isHTTPStatusError`, `runAnimeFireDirectDownloadWithFallback`×3, `downloadAnimeFireDirectWithFallback`, `downloadBloggerDirect` (SSRF), `downloadBloggerChunk` (SSRF), `DownloadVideo`, `downloadWithYtDlp`, `ExtractVideoSources`, `ExtractVideoSourcesWithPrompt`, `getBestQualityURL`, `handleExistingEpisodes`, `askAndPlayDownloadedEpisode`, `HandleBatchDownload`/`getEpisodeRange` (symbol-pin), `printBatchDownloadLocation` |
| `helper_test.go` (novo) | `Init`, `tickCmd`, `Update`×4, `View`×2 |
| `player_unix_test.go` (novo) | `findMPVPath`, `setProcessGroup` |

**Notas de teste:**
- MPV IPC: mock via `net.Listen("unix",…)` em `/tmp/goanime_mpv_*` (path curto p/ limite darwin 104B), respostas JSON com `{"data":<v>,"error":"success"}`.
- Funções network-bound (extractVideoURL, fetchContent, ExtractVideoSources etc.) testadas via path SSRF: `api.SafeGet` rejeita loopback → erro determinístico. Não viola CLAUDE.md "NUNCA rede real".
- Funções TUI puras (huh.NewSelect loop) cuja única saída é via TTY: pin por símbolo + cobertura dos colaboradores. Justificativa documentada inline.
- Tests que usam fuzzyfinder/tcell ou mutam singletons globais (bloggerProxy, GlobalReferer, aniSkipFetcher, cachedDBPath, GlobalSubtitles, gMedia) rodam serial (sem `t.Parallel`) — tcell terminfo lookup é package-level e gera race com `-race`.

**Pendente:** `StartVideo`, `HandleDownloadAndPlay`, `downloadAndPlayEpisode`, `ExtractVideoSources*`, `DownloadVideo`, `downloadWithYtDlp`, `downloadWithNativeHLS`, `HandleBatchDownload`, `getEpisodeRange`, `handleExistingEpisodes`, `askAndPlayDownloadedEpisode`, `handleUpscaleFromMenu`, `downloadSubtitleFiles`, e maioria de `playvideo.go` (`waitForVideoReady`, `seekToResumePosition`, `playVideo`, `showResumeDialog`, `getCurrentEpisode`, `initTracking`, `InitTrackerAsync`, `applyAniSkipResults`, `updateEpisodeDuration`, `preloadNextEpisode`, `startTrackingRoutine`, `showPlayerMenu`, `handleUserInput`, `playNextEpisode`, `playPreviousEpisode`, `selectEpisode`, `switchEpisode`, `skipIntro`, `selectAudioTrack`, `selectSubtitleTrack`) e `scraper.go` heavy fetch (`extractVideoURL`, `fetchContent`, `extractBloggerVideoURL`, `GetVideoURLForEpisode*`, `SelectEpisodeWithFuzzyFinder`, `startBloggerProxy`, `newSurfDownloadClient`).

---

## FASE 12 ✅ — Downloader Completo (~84 funções)
**Arquivos:** `downloader.go`(33), `movie_downloader.go`(28), `nineanime_downloader.go`(16), `hls/hls.go`(7)

Todos com `httptest.Server` mockando CDN. Funções TUI (promptPlay*) → testar lógica, não UI.

**Verificação:** `go test ./internal/downloader/... -v -race`

**Sessão completa** — 1 teste dedicado por função (CLAUDE.md regra #1). Total: **84/84 funções** cobertas. Cobertura: `internal/downloader` 0% → **25.3%**, `internal/downloader/hls` 71%→ **89.0%**.

Distribuição por arquivo:

| Arquivo | Adicionado |
|---|---|
| `hls/hls_test.go` (append) | `NewDownloader`, `Download` (wrapper), `parseMediaPlaylist`×2 (direct + non-HLS), `DownloadToFile` (default-client) |
| `downloader_test.go` (novo) | `NewEpisodeDownloader`, `NewEpisodeDownloaderWithAnime`, `DownloadSingleEpisode`, `DownloadEpisodeRange`×2, `DownloadAllEpisodes`×2, `downloadConcurrentWithProgress`, `downloadMultipleWithProgress` (pin), `downloadEpisodeWithSharedProgress`×2, `findEpisodeByNumber`, `printDownloadLocation`, `fileExists`, `sanitizeDestPath`×3, `episodeFilename`×3, `resolveEpisodeSeason`×2, `episodeDir`×3, `getBestQualityURL` (SSRF), `getContentLength`×3, `estimateContentLengthForAllAnime`×2, `downloadWithProgress`/`downloadHTTPWithProgress`/`downloadM3U8WithYtDlp`/`downloadWithYtDlp` (pin), `downloadEpisodeWithProgress` (empty URL), `isUnsafeExtError`×4, `promptPlayExisting`/`promptPlayDownloaded` (closed stdin), `promptPlayDownloadedRangeHuh`/`promptPlayExistingRangeHuh` (empty list), `playEpisode` (pin), `tickCmd`, `progressModel.Init`/`Update`×3/`View` |
| ~~movie_downloader_test.go~~ | DELETADO — `internal/downloader/movie_downloader.go` removido junto com SFlix/FlixHQ em 2026-05-17 |
| ~~nineanime_downloader_test.go~~ | DELETADO — `internal/downloader/nineanime_downloader.go` removido em 2026-05-17 |

**Notas de teste:**
- Funções network-bound (downloadHTTPWithProgress, downloadM3U8WithYtDlp*, downloadStream, downloadNativeHLS, etc.) testadas via path SSRF: `api.SafeTransport` rejeita loopback → erro determinístico. Funções yt-dlp wrapped + funções que driveriam Bubble Tea `p.Run()` ficam como pin de símbolo (cobertura 0% nessas, mas teste dedicado existe). Justificativa: yt-dlp lança binário externo; tea.Program.Run requer TTY.
- TUI prompts (`promptPlay*`) testados via `withClosedStdin(t)` que redireciona `os.Stdin` para `/dev/null` — `fmt.Scanln` retorna EOF → caminho "n / cancel". Roda serial (sem `t.Parallel`) por mutar global.
- `promptSubtitleLanguage` testado em todos os branches pré-configurados (none/all/exact match/cached/empty) sem precisar do fuzzyfinder TUI.
- 46 funções pinned/0% — todas têm teste dedicado nomeado. Para passar de 25% seria necessário injetar yt-dlp mock + harness Bubble Tea ou skip-test em FFmpeg/binário externo.

---

## FASE 13 ✅ — API Movie + Enhanced HTTP + Providers (~100 funções)
**Arquivos:** `api/movie/`(27), `api/enhanced.go` HTTP(~16), `api/episode_providers.go`(7), `api/allanime_enhanced.go`(4), `api/providers/`(46+5+7+9)

| Sub-pacote | Funções |
|---|---|
| `api/movie/omdb.go` | 10 (NewOMDbClient, IsConfigured, SearchByTitle, GetByIMDBID, GetByTitle, makeRequest, GetRuntimeMinutes, GetRating, GetGenres) |
| `api/movie/tmdb.go` | 14 (NewTMDBClient, IsConfigured, SearchMulti, SearchMovies, SearchTV, GetTVSeasons, GetSeasonEpisodes, GetCredits, FindByIMDBID, GetTrending, GetPopular, GetImageURL) |
| `api/movie/enrich.go` | 3 (EnrichMedia, EnrichWithOMDb, FormatMediaInfo) |
| `api/providers/registry.go` | 5 (RegisterProvider, ForKind, ForAnime, HasProvider, ResetForTesting) |
| `api/providers/source_providers.go` | 46 (8 providers × ~6 methods each) |
| `api/providers/metadata/` | 7 |
| `api/providers/naming/` | 9 |

**Verificação:** `go test ./internal/api/movie/ ./internal/api/providers/... ./internal/api/ -run "TestEnhanced|TestEpisodeProv|TestAllAnimeEnhanced" -v -race`

---

## FASE 14 ✅ — Handlers + Playback + Resto (~120 funções)
**Arquivos:** `handlers/`(28), `playback/`(23), `download/workflow.go`(10), `discord/`(34), `tracking/`(7), `updater/`(12), `tui/`(7), `upscaler/`(47), `appflow/`(2), `pkg/goanime/client.go`(7), `scraper/movie/`(42)

| Sub-pacote | Funções | Nota |
|---|---|---|
| `handlers/` | 28 | Muitas dependem de TUI → testar routing logic |
| `playback/` | 23 | MPV boundary → mock IPC |
| `download/` | 10 | Workflow → mock API |
| `discord/` | 34 | RPC daemon → mock interface |
| `tracking/` | 7 | SQLite → t.TempDir() |
| `updater/` | 12 | HTTP → httptest |
| `upscaler/` | 47 | FFmpeg → skip GPU, testar config/options |
| `scraper/movie/` | 42 | Delegates → mock |

**Verificação:** `go test ./internal/handlers/ ./internal/playback/ ./internal/download/ ./internal/discord/ ./internal/tracking/ ./internal/updater/ ./internal/upscaler/ ./internal/scraper/movie/ ./internal/appflow/ ./pkg/goanime/ -v -race`

---

# Fases Adicionais — Push para 70% via "1 Teste por Função" Estrito (2026-05-18)

Estado pós-FASE 14: **52.8%** (12065 statements / 5692 missed) · **165 funções ainda a 0%**.

**Meta dupla:**
1. **≥ 70.0% cobertura total** (≥ 8447 statements cobertos, +2073 vs. atual)
2. **≤ 30 funções a 0%** (cobrir 135+ das 165 atuais — apenas TUI/IPC/main intratáveis ficam)

**Regra absoluta reafirmada (CLAUDE.md):** *cada* função listada como 0% recebe seu próprio `TestNomeDaFuncao_Cenario`. Sem agrupar, sem pular. Eficácia brutal.

**Refactor agora amplamente permitido** quando necessário para testar — usuário autoriza "vale tudo". Mantemos apenas a regra dura: API pública NÃO quebra (semver). Adicionar interface, var injetável, helper `*ForTesting`, split de função orquestrada — tudo OK.

### Distribuição Real das 165 Funções 0% (extraído 2026-05-18)

| Pacote | Funcs 0% | Cov atual | Fase alvo |
|---|---:|---:|:---:|
| `internal/util/` | 30 | 44.7% | 15 |
| `internal/api/` | 27 | 42.8% | 15 |
| `internal/playback/` | 14 | 13.8% | 16 |
| `internal/scraper/` | 13 | 78.6% | 17 |
| `internal/handlers/` | 12 | 5.7% | 16 |
| `internal/upscaler/` | 11 | 49.2% | 16 |
| `internal/discord/` | 10 | 29.5% | 16 |
| `internal/updater/` | 8 | 53.4% | 16 |
| `internal/api/providers/` | 8 | 48.7% | 17 |
| `pkg/goanime/types/` | 7 | 0.0% | 17 |
| `pkg/goanime/` | 5 | 5.0% | 17 |
| `internal/downloader/` | 5 | 34.0% | 17 |
| `internal/tui/` | 3 | 63.6% | 17 |
| `internal/download/` | 2 | 0.0% | 17 |
| `internal/appflow/` | 2 | 22.9% | 17 |
| `pkg/goanime/examples/*` | 4 | 0.0% | (exceção `main()`) |
| `internal/tracking/` | 1 | 68.5% | 17 |
| `internal/player/` | 1 | 51.5% | 17 |
| `internal/api/providers/metadata/` | 1 | 72.3% | 17 |
| `cmd/goanime/` | 1 | — | (exceção `main()`) |
| **TOTAL** | **165** | | |

### Manifestos

| Fase | Arquivo | Funções |
|---|---|---:|
| 15 | `.test_manifests/p15_api_util.txt` | 57 |
| 16 | `.test_manifests/p16_tui_ipc.txt` | 55 |
| 17 | `.test_manifests/p17_remaining.txt` | 53 |
| **TOTAL** | | **165** |

Para regenerar manifestos (após cada fase concluída):
```bash
go test ./... -short -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | awk '$NF == "0.0%" {print $1, $2}' > /tmp/zero_funcs.txt
# Atenção: usar awk '$NF == "0.0%"', NÃO grep "0.0%" — grep também matches "100.0%", "80.0%", etc.
```

---

## FASE 15 ✅ — API + Util: 57 funções (Branches + Error Paths) (2026-05-21)

**Pacotes:**
- `internal/api/` (1236 stmts, 42.8% → 60%, 27 funcs 0%)
- `internal/util/` (978 stmts, 44.7% → 65%, 30 funcs 0%)

Lista completa em `.test_manifests/p15_api_util.txt` e `TEST_PLAN_FUNCTIONS.md` (FASE 15).

### Funções-chave por arquivo

#### `internal/api/` (27 funcs)
- `allanime_enhanced.go`: `GetEpisodeStreamURLEnhanced`, `GetAllAnimeEpisodeURLDirect`
- `allanime_smart.go`: `DownloadAllAnimeSmartRange`, `smartDownload`, `smartDownloadDirect`, `resolveStreamURLForEpisode`
- `anime.go`: `GetEpisodeData`, `GetMovieData`, `FetchAnimeDetails`, `enrichAnimeData`, `httpPostFast`, `safeClose` (mover via injection)
- `anime_url_title.go`: `FetchAnimeFromAniListWithURL`
- `aniskip.go`: `GetAniSkipData`, `GetAndParseAniSkipData`
- `enhanced.go`: `SearchAnimeEnhanced`, `GetAnimeEpisodesEnhanced`, `GetEpisodeStreamURL`, `DownloadEpisodeEnhanced`, `DownloadEpisodeRangeEnhanced`, `downloadFromURL`, `GetSuperFlixEpisodes`, `GetSuperFlixStreamURL`
- `episode_providers.go`: `FetchEpisodeData` per provider, `getAniListIDFromMAL`, `getKitsuAnimeID`

#### `internal/util/` (30 funcs)
Maioria são helpers de path, URL, fila de progresso, locale, formatters. Ver lista exata em `TEST_PLAN_FUNCTIONS.md`.

### Refactors permitidos
1. `var anilistBaseURL`, `var kitsuBaseURL`, `var jikanBaseURL`, `var aniskipBaseURL` substituíveis
2. `SearchAnimeEnhanced` → extrair `searchAllSources(name) (results, error)` (sem TUI/fuzzyfinder)
3. `DownloadAllAnimeSmartRange` → extrair `validateAndPlanRange` testável

**Verificação:**
```bash
go test ./internal/api/... ./internal/util/ -count=1 -race -coverprofile=p15.out
go tool cover -func=p15.out | awk '$NF == "0.0%"' | grep -E "internal/(api|util)" | wc -l  # Esperado: ≤ 10
go tool cover -func=p15.out | tail -1  # Esperado: ≥ 62%
```

---

## FASE 16 ✅ — Playback + Handlers + Discord + Upscaler + Updater: 55 funções (TUI/IPC com refactor) (2026-05-22)

**Pacotes:**
- `internal/playback/` (14 funcs) — TUI navigation
- `internal/handlers/` (12 funcs) — TUI menus
- `internal/upscaler/` (11 funcs) — shaders install + video probe
- `internal/discord/` (10 funcs) — RPC presence
- `internal/updater/` (8 funcs) — release download/apply

Lista completa em `.test_manifests/p16_tui_ipc.txt` e `TEST_PLAN_FUNCTIONS.md` (FASE 16).

### Refactors permitidos (autorizados — "vale tudo")

#### `playback/`
- `HandleSeries`, `HandleMovie` → extrair `prepareSeriesContext(anime) (*seriesCtx, error)` e `prepareMovieContext(anime) (*movieCtx, error)` puros + testáveis. Loop interativo permanece intratável → testado via helpers extraídos.
- `handleUserNavigation*` → extrair `decideNavigation(input, state) NavCmd` puro.
- `SelectInitialEpisode` → extrair `pickInitialEpisode(episodes, savedEp) int` puro.

#### `handlers/`
- `SearchMedia`, `SelectMediaType`, `InteractiveMediaFlow`, `handleAnimePlayback` → extrair `dispatchByMediaType(mt MediaType) HandlerFunc`.
- `HandleDownloadRequest`, `HandleMovieDownloadRequest`, `HandleUpdateRequest`, `HandleUpscaleRequest`, `HandlePlaybackMode` → injetar dependências via `Options{...}` struct.
- `handleImageUpscale`/`handleVideoUpscale` → extrair `validateUpscaleInput(path, opts) error`.

#### `discord/`
- `type rpcClient interface { Login() error; Logout() error; SetActivity(client.Activity) error }`
- `var discordClientFactory = func(id string) rpcClient { return client.NewClient(id) }` substituível
- `LoginClient`, `LogoutClient`, `Start`, `Stop`, `updateDiscordPresence`, `getPrecisePlaybackState`, `buildPreciseTimestamps`, `FetchDuration` ficam testáveis com mock

#### `upscaler/`
- `var shadersZipURL = "..."` substituível (também `ganShadersZipURL`)
- `InstallShaders`, `InstallGANShaders`, `extractZip` → testar com httptest serve in-memory zip + `t.TempDir()`
- `UpscaleVideo`, `probeVideo`, `extractFrames`, `upscaleFrames`, `encodeVideo` → mockar `exec.Command` via interface

#### `updater/`
- `var releaseAPIURL`, `var releaseDownloadURL` substituíveis
- `CheckForUpdates`, `PromptForUpdate`, `PerformUpdate`, `downloadUpdate`, `extractArchive`, `applyUpdate` → httptest serve tarball + `t.TempDir()` para o destino

**Verificação:**
```bash
go test ./internal/playback/ ./internal/handlers/ ./internal/discord/ ./internal/upscaler/ ./internal/updater/ -count=1 -race -coverprofile=p16.out
go tool cover -func=p16.out | awk '$NF == "0.0%"' | wc -l  # Esperado: ≤ 15
```

**Sessão completa (2026-05-22)** — 55/55 funções com teste dedicado (CLAUDE.md regra #1). Combined p16 coverage 33.5% → **64.5%**. Refatorações aplicadas (vars injetáveis, sem quebra de API pública):

| Pacote | Refactor | Cov antes → depois | Pendentes a 0% |
|---|---|---|---|
| `discord/` | `rpcClient` interface + `var newRPCClient` factory | 29.5% → **94.0%** | 0 |
| `updater/` | `var releaseAPIURL`, `findAssetFn`, `downloadFn`, `osExecutableFn`, `replaceExecutableFn`, `runForm` | 53.4% → **72.2%** | 0 |
| `upscaler/` | `var anime4kShaderURL`, `anime4kGANShaderBaseURL`, `shaderDirOverride` | 49.2% → **73.0%** | `Close` (empty body → cover tool quirk) |
| `handlers/` | `mediaSource` interface + `runFormFn`/`findFn`/`findResultFn` hooks | 5.7% → **55.9%** | `HandlePlaybackMode` (real-network TUI loop) |
| `playback/` | sem refactor; testes exercitam paths internos + symbol-pin TUI | 13.3% → **33.4%** | `GetUserInput`, `HandleMovie`, `HandleSeries`, `SelectInitialEpisode`, `ChangeAnimeLocal` |

**Notas:**
- IPC do Discord: stub via `rpcClient` interface; cobre Login/Logout/Start/Stop/updateDiscordPresence/getPrecisePlaybackState/buildPreciseTimestamps/FetchDuration sem socket real.
- Updater: `runForm = tui.RunClean` substituível permite testar `PromptForUpdate` e `CheckAndPromptUpdate` sem TTY. `replaceExecutableFn` permite testar `PerformUpdate` sem mexer no binário do teste.
- Upscaler: `extractFrames`/`encodeVideo` testados com `/usr/bin/true` como FFmpeg stub. `extractZip` + `InstallShaders` exercitados via httptest serve zip in-memory. `upscaleSingleFrame` testado com PNG tiny real (sem FFmpeg).
- Handlers: `mediaSource` interface elimina rede em `SearchMedia`/`InteractiveMediaFlow`/`handleAnimePlayback`. `findFn`/`runFormFn` hooks substituem TUI por stubs determinísticos.
- Playback: funções TUI puras (`HandleSeries`/`HandleMovie`/`ChangeAnimeLocal`/`GetUserInput`) ficam como symbol-pin — driveriam huh forms + tcell fuzzyfinder que requerem TTY real. `PlayEpisode` parcialmente coberto via path de erro (videoErr early return).

---

## FASE 17 ✅ — Scraper + Providers + Downloader + SDK + Misc: 53 funções

**Pacotes (organizado por gap):**
- `internal/scraper/` (13 funcs)
- `internal/api/providers/` (8 funcs)
- `pkg/goanime/` + `pkg/goanime/types/` (12 funcs)
- `internal/downloader/` (5 funcs)
- `internal/tui/` (3 funcs)
- `internal/download/` (2 funcs)
- `internal/appflow/` (2 funcs)
- `internal/tracking/`, `internal/player/`, `internal/api/providers/metadata/` (1 func cada)

Lista completa em `.test_manifests/p17_remaining.txt` e `TEST_PLAN_FUNCTIONS.md` (FASE 17).

### Exceções aceitas nesta fase (não testar)
- `cmd/goanime/main.go:main` (CLI entry)
- `pkg/goanime/examples/*/main.go:main` (4 funcs — exemplos)

### Refactors permitidos
- `pkg/goanime/Client`: receber `httpClient *http.Client` injetável
- Downloader: `httpClient` injetável + worker pool size opt-in
- Restante: refactor mínimo se necessário (vars de URL, interface wrap)

**Verificação:**
```bash
go test ./... -count=1 -race -short -coverprofile=p17.out
go tool cover -func=p17.out | awk '$NF == "0.0%"' | wc -l  # Esperado: ≤ 30 (apenas exceções `main()`)
go tool cover -func=p17.out | tail -1  # Esperado: ≥ 70%
```

---

## Resumo Push 70% (FASES 15–17)

| Fase | Pacotes | Funcs 0% alvo | Stmts esperados | Refactors |
|---|---|---:|---:|:---:|
| 15 | api + util | 57 | +600 | 3 mín (vars + 2 splits) |
| 16 | playback + handlers + discord + upscaler + updater | 55 | +900 | 5–6 (interface + vars + splits) |
| 17 | scraper + providers + downloader + SDK + misc | 53 | +600 | 2–3 (injetar httpClient) |
| **TOTAL** | | **165** | **+2100** | **10–12 refactors** |

**Cobertura projetada após FASE 17:** 52.8% + (2100/12065 × 100) ≈ **70.2%**

**Funções 0% projetadas:** 165 → ≤ 30 (apenas: `main()` do CLI + 4 exemplos do SDK + Bubble Tea integrated loops + funções com hardware GPU)

---

## FASE 18 ✅ — Push Final 70%: types + SDK + API success paths + exec mock + appflow + download (28 funções)

**Meta:** 62.9% → ≥ 70.0% | Gap: **858 statements** (12102 total, 7613 cobertos → precisam 8471)
**Data planejada:** 2026-05-24

**Medições base (2026-05-24, `go test ./... -race -coverprofile`):**
| Pacote | Cobertura | Stmts perdidos |
|---|---|---|
| `pkg/goanime/types` | 0.0% | 41 |
| `pkg/goanime` | 4.9% | 39 |
| `internal/api` | 63.3% | 453 |
| `internal/downloader` | 37.7% | 370 |
| `internal/appflow` | 15.6% | 92 |
| `internal/download` | 2.5% | 77 |

---

### Ação 18A — `pkg/goanime/types` (7 funções a 0%, ~41 stmts)

**Tipo:** Unitário puro — sem I/O, sem dependências externas
**Arquivo de teste:** `pkg/goanime/types/types_test.go` (novo)
**Refactor:** nenhum

| Arquivo | Linha | Função |
|---|---|---|
| `types/anime.go` | 97 | `FromInternalAnime` |
| `types/anime.go` | 148 | `FromInternalAnimeList` |
| `types/anime.go` | 157 | `FromInternalEpisode` |
| `types/anime.go` | 199 | `FromInternalEpisodeList` |
| `types/source.go` | 20 | `String` |
| `types/source.go` | 32 | `ToScraperType` |
| `types/source.go` | 44 | `ParseSource` |

Padrão: table-driven com structs `internal.Anime` / `internal.Episode` construídas manualmente. Verificar que os campos mapeados são corretos (title, episodes, URL, source string).

---

### Ação 18B — `pkg/goanime/client` (6 funções a 0%, ~39 stmts)

**Tipo:** Unit + httptest — `NewClientForTest` já aceita injeção de `*http.Client`
**Arquivo de teste:** `pkg/goanime/client_test.go` (existente — expandir)
**Refactor:** confirmar/ajustar `NewClientForTest` para aceitar `baseURL` ou `*http.Client` injetável

| Arquivo | Linha | Função |
|---|---|---|
| `client.go` | 25 | `SearchAnime` |
| `client.go` | 43 | `GetAnimeEpisodes` |
| `client.go` | 68 | `GetStreamURL` |
| `client.go` | 86 | `DefaultStreamOptions` |
| `client.go` | 105 | `GetEpisodeStreamURL` |
| `client.go` | 137 | `NewClientForTest` |

Cada função recebe um `httptest.Server` que retorna JSON válido de AllAnime/AniList. Usar `NewClientForTest(srv.Client(), srv.URL)`.

---

### Ação 18C — `internal/api` success paths (7 funções parciais, ~280 stmts)

**Tipo:** Unit + httptest — adicionar casos de **sucesso** aos testes existentes das fases 2 e 15
**Arquivos de teste:** expandir `internal/api/*_test.go` existentes
**Refactor:** verificar que `var allAnimeBaseURL`, `var aniListBaseURL`, `var jikanBaseURL` estão injetáveis (devem estar da Fase 15); adicionar se faltarem

| Arquivo | Linha | Função | Cobertura atual | Problema |
|---|---|---|---|---|
| `anime.go` | 28 | `GetEpisodeData` | < 50% | Só caminhos de erro testados; falta mock de resposta GraphQL válida |
| `anime.go` | 33 | `GetMovieData` | < 50% | Idem — falta mock AllAnime movie response |
| `enhanced.go` | 481 | `DownloadEpisodeEnhanced` | 38.5% | Caminho de download nunca exercitado; usar `t.TempDir()` como destino |
| `enhanced.go` | 509 | `DownloadEpisodeRangeEnhanced` | 29.4% | Range [1-2] com mock CDN serving tiny file |
| `allanime_smart.go` | 81 | `smartDownload` | < 50% | mock AllAnime + destino em `t.TempDir()` |
| `allanime_smart.go` | 21 | `DownloadAllAnimeSmartRange` | < 50% | mock + range "1-2" |
| `allanime_enhanced.go` | 14 | `GetEpisodeStreamURLEnhanced` | < 50% | mock retornando URL de stream válido |

Fixtures JSON necessárias (criar em `internal/api/testdata/`):
- `allanime_episode_response.json` — resposta GraphQL `episode` com `sourceUrls`
- `allanime_movie_response.json` — resposta GraphQL `show` com tipo movie
- `allanime_stream_response.json` — resposta de URL de stream com `links`

---

### Ação 18D — `internal/downloader` exec mock (3 funções a 0%, ~150 stmts)

**Tipo:** Refactor mínimo + Unit
**Arquivo:** `internal/downloader/downloader.go` (1 var nova) + `downloader_test.go` (expandir)
**Refactor:**
```go
// downloader.go — adicionar no topo do arquivo (junto às outras vars)
var execCommand = exec.Command
// Substituir todas as chamadas `exec.Command(...)` por `execCommand(...)`
```

| Arquivo | Linha | Função |
|---|---|---|
| `downloader.go` | 710 | `downloadWithProgress` |
| `downloader.go` | 941 | `downloadM3U8WithYtDlp` |
| `downloader.go` | 1171 | `playEpisode` |

Padrão de teste:
```go
func TestDownloadWithProgress_ExecMock(t *testing.T) {
    t.Parallel()
    orig := execCommand
    t.Cleanup(func() { execCommand = orig })
    execCommand = func(name string, args ...string) *exec.Cmd {
        return exec.Command("/usr/bin/true")  // noop stub
    }
    // exercitar downloadWithProgress com httptest server como CDN
}
```

---

### Ação 18E — `internal/appflow` injection (4 funções a 0%, ~92 stmts)

**Tipo:** Refactor + Unit + MockScraper (reutilizar `createTestManager` de `internal/scraper/unified_test.go`)
**Arquivo de teste:** `internal/appflow/anime_data_test.go` (novo)
**Refactor:** cada função aceita um parâmetro opcional `manager ...ScraperManagerInterface` ou via `var defaultScraperFactory`

```go
// appflow/anime_data.go — refactor
var defaultScraperFactory = func() ScraperManagerInterface {
    return scraper.GetScraperManager()
}

func SearchAnime(name string, manager ...ScraperManagerInterface) ([]scraper.Anime, error) {
    m := defaultScraperFactory()
    if len(manager) > 0 { m = manager[0] }
    // ...resto igual
}
```

| Arquivo | Linha | Função |
|---|---|---|
| `anime_data.go` | 20 | `SearchAnime` |
| `anime_data.go` | 34 | `SearchAnimeEnhanced` |
| `anime_data.go` | 48 | `SearchAnimeWithRetry` |
| `anime_data.go` | 103 | `FetchAnimeDetails` |

---

### Ação 18F — `internal/download/workflow.go` injection (1 função a 0%, ~77 stmts)

**Tipo:** Refactor + Unit com mock downloader
**Arquivo de teste:** `internal/download/workflow_test.go` (novo)
**Refactor:** `HandleDownloadRequest` aceita interface injetável

```go
// download/workflow.go
type EpisodeDownloader interface {
    DownloadSingleEpisode(ep scraper.Episode, opts ...DownloadOption) error
    DownloadEpisodeRange(start, end int, eps []scraper.Episode, opts ...DownloadOption) error
}

var defaultDownloaderFactory = func(anime scraper.Anime) EpisodeDownloader {
    return downloader.NewEpisodeDownloader(anime)
}
```

| Arquivo | Linha | Função |
|---|---|---|
| `workflow.go` | 18 | `HandleDownloadRequest` |

---

### Verificação FASE 18

```bash
# Por ação — rodar após cada ação concluída
go test ./pkg/goanime/types/ -v -race -count=1                          # 18A
go test ./pkg/goanime/ -v -race -count=1                                # 18B
go test ./internal/api/ -v -race -count=1 -run "TestGetEpisode|TestDownload|TestSmart|TestGetEpisodeStreamURL"  # 18C
go test ./internal/downloader/ -v -race -count=1 -run "TestDownloadWith|TestM3U8|TestPlayEpisode"  # 18D
go test ./internal/appflow/ -v -race -count=1                           # 18E
go test ./internal/download/ -v -race -count=1                          # 18F

# Final — meta ≥ 70%
go test ./... -coverprofile=coverage.out -covermode=atomic -race
go tool cover -func=coverage.out | tail -1
go tool cover -func=coverage.out | awk '$NF == "0.0%"' | grep -v "examples\|cmd/goanime" | wc -l
```

**Critérios de aceite:**
- `total: ≥ 70.0%`
- Funções não-main a 0%: ≤ 15

---

## Checklist

| Fase | Escopo | Funções | Status |
|---|---|---|---|
| 1 | Models + Types + Source + AniSkip | ~50 | ✅ |
| 2 | API Pure (anime, episodes, enhanced, smart) | ~45 | ✅ |
| 3 | SSRF + Player Pure | ~40 | ✅ |
| 4 | Scraper Infrastructure | ~45 | ✅ |
| 5 | Unified Adapters | ~45 | ✅ |
| 6 | Util Completo | ~83 | ✅ |
| 7 | FlixHQ | — | ❌ (removido 2026-05-17) |
| 8 | SFlix | — | ❌ (removido 2026-05-17) |
| 9 | AnimeFire + Goyabu + AllAnime | ~29 | ✅ (NineAnime removido 2026-05-17) |
| 10 | SuperFlix + MediaManager | ~69 | ✅ (AnimeDrive removido 2026-05-17) |
| 11 | Player Completo | ~128 | ✅ |
| 12 | Downloader Completo | ~84 | ✅ |
| 13 | API Movie + Enhanced + Providers | ~100 | ✅ (2026-05-18) |
| 14 | Handlers + Playback + Discord + Upscaler + Resto | ~120 | ✅ (2026-05-18) |
| 15 | API + Util (57 funcs 0%) | +57 funcs / +600 stmts | ✅ (2026-05-21 — api 42.3%→63.6%, util 44.7%→75.8%, total 52.8%→56.8%) |
| 16 | Playback + Handlers + Discord + Upscaler + Updater (55 funcs) | +55 funcs / +900 stmts | ✅ (2026-05-22 — discord 29.5%→94.0%, handlers 5.7%→55.9%, playback 13.3%→33.4%, updater 53.4%→72.2%, upscaler 49.2%→73.0%, total 56.8%→61.5%, 0% funcs 112→64) |
| 17 | Scraper + Providers + Downloader + SDK + Misc (53 funcs) | +53 funcs / +600 stmts | ✅ (2026-05-23 — scraper 83.2%, providers 69.9%, types 80.5%→100%, pkg/goanime 95.1%, tui 91.1%, total 57.0%→59.0% [-short], 0% funcs 112→68; 68 restantes = 5 main()+4 ex.main()+MPV/TTY/HW não-testáveis) |
| 18 | types + SDK client + API success paths + exec mock + appflow + download (28 funcs) | +28 funcs / +858 stmts | ✅ (2026-05-24 — 62.9% → **75.7%**, non-main 0%: 26 → **2** [tui.Find + upscaler.Close]) |
| **TOTAL** | | **~1176 funcs / ~+2958 stmts** | |

**Medições pós-FASE 17 (reais, 2026-05-24):** 62.9% cobertura (full suite sem -short) · 36 funções a 0% (26 non-main)
**Medições pós-FASE 18 (reais, 2026-05-24):** **75.7%** cobertura (full suite com -race) · **2 funções non-main a 0%** (`tui.Find` + `upscaler.Close`) — ambas TUI/hardware não-testáveis. Meta ≥70% ATINGIDA.
