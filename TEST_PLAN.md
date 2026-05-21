# GoAnime — Plano de Testes: 25.9% → 70%

> **Data inicial:** 2026-04-26 · **Estado atual (2026-05-18):** `52.8%` · **Alvo:** `70%` · **Gap restante:** `17.2 pontos percentuais`
> **Total statements:** 12065 · **Missed:** 5692 · **A cobrir para 70%:** +2073 statements
> **Funções a 0%:** 165 (contagem exata via `awk '$NF == "0.0%"'`)

> **Histórico:**
> - Sprint inicial (2026-04-26 → 2026-05-17): FASES 1–14, alcançou **52.8%**, ~983 funções testadas
> - Push 70% (2026-05-18 → em planejamento): FASES 15–17, alvo +2100 statements, +165 funções → ≤ 30 funcs 0%

---

## 1. Diagnóstico Real — Cobertura por Pacote

| Pacote | Cobertura | Funções 0% | Status |
|---|---|---|---|
| `api/source` | **98.5%** | 5 | ✅ Excelente |
| `api/providers/naming` | **96.5%** | 9 | ✅ Excelente |
| `downloader/hls` | **85.8%** | 7 | ✅ Ótimo |
| `api/providers/metadata` | **72.6%** | 7 | ✅ OK |
| `tui` | **63.6%** | 7 | 🟡 Médio |
| `tracking` | **60.4%** | 7 | 🟡 Médio |
| `updater` | **53.4%** | 12 | 🟡 Médio |
| `scraper` | **42.1%** | 181 | 🔴 Baixo — MAIOR GAP |
| `upscaler` | **29.2%** | 47 | 🔴 Baixo |
| `appflow` | **22.7%** | - | 🔴 Baixo |
| `api` (core) | **17.0%** | 78 | 🔴 Crítico |
| `player` | **17.2%** | 123 | 🔴 Crítico |
| `api/movie` | **15.2%** | 16 | 🔴 Crítico |
| `util` | **14.6%** | 66 | 🔴 Crítico |
| `playback` | **2.0%** | 25 | 🔴 Crítico |
| `pkg/goanime` | **5.0%** | 11 | 🔴 Crítico |
| `handlers` | **0%** | 27 | 🔴 Zero |
| `downloader` | **0%** | 77 | 🔴 Zero |
| `discord` | **0%** | 34 | 🔴 Zero |
| `models` | **0%** | 13 | 🔴 Zero |
| `download` | **0%** | 10 | 🔴 Zero |

### Contagem Total de Funções no Projeto

| Métrica | Valor |
|---|---|
| **Funções totais no projeto** | **1217** |
| Funções com cobertura > 0% | 184 (15.1%) |
| Funções a 0.0% (sem nenhum teste) | **1033** (84.9%) |

**Distribuição das 1033 funções sem cobertura por pacote:**

| Pacote | Funções 0% | % do gap total |
|---|---|---|
| `internal/scraper/` | 220 | 21.3% |
| `internal/player/` | 123 | 11.9% |
| `internal/util/` | 83 | 8.0% |
| `internal/api/` (core + enhanced + smart) | 78 | 7.6% |
| `internal/downloader/` | 77 | 7.5% |
| `internal/upscaler/` | 47 | 4.6% |
| `internal/api/providers/` | 46 | 4.5% |
| `internal/discord/` | 34 | 3.3% |
| `internal/handlers/` | 27 | 2.6% |
| `internal/playback/` | 25 | 2.4% |
| `internal/scraper/movie/` | 22 | 2.1% |
| `pkg/goanime/` | 18 | 1.7% |
| `internal/api/movie/` | 16 | 1.5% |
| `internal/models/` | 13 | 1.3% |
| `internal/updater/` | 12 | 1.2% |
| `internal/download/` | 10 | 1.0% |
| `internal/appflow/` | 8 | 0.8% |
| `internal/tui/` | 7 | 0.7% |
| `internal/tracking/` | 7 | 0.7% |
| `internal/api/source/` | 5 | 0.5% |
| `internal/version/` | 2 | 0.2% |
| `cmd/goanime/` | 1 | 0.1% |

> 📋 **Lista completa de todas as 1033 funções:** [TEST_PLAN_FUNCTIONS.md](TEST_PLAN_FUNCTIONS.md)

---

## 2. Cálculo: Quantos Testes para 70%?

### Método

- **1217 funções** totais no projeto
- Cobertura atual: **25.9%** ≈ 315 funções com alguma cobertura
- Para 70%: **852 funções** precisam ter cobertura
- **Gap: ~537 funções** precisam de pelo menos 1 teste que passe por elas
- Estimativa: **1 test function cobre ~3-5 funções** (table-driven com subcases)
- **Estimativa total: ~130-160 novos testes**

### Distribuição por Arquivo (TOP 25 — Maior Impacto)

Os 25 arquivos abaixo concentram **~750 das 1033 funções a 0%**. Cobrir eles é o caminho mais rápido para 70%.

| # | Arquivo | Funções 0% | Testes Necessários | Impacto |
|---|---|---|---|---|
| 1 | `scraper/media_manager.go` | 60 | ~15 | 🔴 Alto |
| 2 | `scraper/flixhq.go` | 54 | ~14 | 🔴 Alto |
| 3 | `scraper/sflix.go` | 46 | ~12 | 🔴 Alto |
| 4 | `scraper/unified.go` | 45 | ~12 | 🔴 Alto |
| 5 | `api/providers/source_providers.go` | 41 | ~10 | 🔴 Alto |
| 6 | `player/player.go` | 40 | ~10 | 🔴 Alto |
| 7 | `downloader/downloader.go` | 33 | ~8 | 🔴 Alto |
| 8 | `player/playvideo.go` | 31 | ~8 | 🔴 Alto |
| 9 | `downloader/movie_downloader.go` | 28 | ~7 | 🟡 Alto |
| 10 | `player/download.go` | 27 | ~7 | 🟡 Alto |
| 11 | `util/util.go` | 26 | ~7 | 🟡 Médio |
| 12 | `discord/discord.go` | 26 | ~7 | 🟡 Baixo |
| 13 | `player/scraper.go` | 25 | ~6 | 🟡 Alto |
| 14 | `scraper/nineanime.go` | 21 | ~5 | 🟡 Alto |
| 15 | `scraper/movie/sflix.go` | 21 | ~5 | 🟡 Médio |
| 16 | `scraper/movie/flixhq.go` | 21 | ~5 | 🟡 Médio |
| 17 | `scraper/animedrive.go` | 21 | ~5 | 🟡 Médio |
| 18 | `util/httpclient.go` | 20 | ~5 | 🟡 Médio |
| 19 | `handlers/media.go` | 20 | ~5 | 🟡 Médio |
| 20 | `api/enhanced.go` | 20 | ~5 | 🟡 Médio |
| 21 | `api/anime.go` | 19 | ~5 | 🔴 Alto |
| 22 | `upscaler/video.go` | 17 | ~4 | 🟡 Médio |
| 23 | `upscaler/anime4k.go` | 17 | ~4 | 🟡 Médio |
| 24 | `util/perf.go` | 16 | ~4 | 🟡 Baixo |
| 25 | `util/logger.go` | 16 | ~4 | 🟡 Baixo |

**Subtotal TOP 25:** 711 funções · ~177 testes

**Restante (56 arquivos menores):** 322 funções · ~60 testes (muitas são 1-3 funções por arquivo)

### Resumo Final

| Categoria | Testes Novos | Funções Cobertas |
|---|---|---|
| **Scrapers** (flixhq, sflix, nineanime, animedrive, superflix, goyabu, allanime, media_manager) | ~65 | ~270 |
| **Player** (player, playvideo, download, scraper) | ~31 | ~123 |
| **API** (anime, api, enhanced, episodes, allanime_smart, aniskip) | ~25 | ~78 |
| **Downloader** (downloader, movie_downloader, nineanime_downloader) | ~15 | ~77 |
| **Util** (util, httpclient, perf, logger, ytdlp) | ~15 | ~66 |
| **Handlers/Playback** (media, download, playback series/movie) | ~12 | ~52 |
| **Upscaler** (anime4k, video, shaders) | ~8 | ~47 |
| **Outros** (models, discord, pkg, appflow, providers) | ~12 | ~60 |
| **TOTAL** | **~183 testes** | **~773 funções** |

> **~183 novos testes** levam de **25.9% → ~70%** de cobertura.

---

## 3. Plano de Testes Detalhado por Arquivo

### 3.1 🔴 `internal/scraper/` — 65 testes (~270 funções)

#### `scraper/media_manager.go` — 15 testes

```
TestMediaManager_GetInstance_Singleton
TestMediaManager_SearchMedia_RoutesToCorrectScraper
TestMediaManager_SearchMedia_MovieVsAnimeRouting
TestMediaManager_GetEpisodes_FlixHQ
TestMediaManager_GetEpisodes_SFlix
TestMediaManager_GetStreamURL_WithHeaders
TestMediaManager_GetStreamURL_FallbackOnError
TestMediaManager_ConcurrentAccess_ThreadSafe
TestMediaManager_SearchByType_FiltersCorrectly
TestMediaManager_GetInstance_LazyInit
TestMediaManager_MovieSearch_RoutesToFlixHQ
TestMediaManager_TVSearch_RoutesToFlixHQ
TestMediaManager_RegisterScraper_CustomProvider
TestMediaManager_SearchWithTimeout_ReturnsPartial
TestMediaManager_SearchWithContext_Cancellation
```

**Mock strategy:** Usar `MockScraper` já existente em `unified_test.go`. Injetar no manager via campo `scrapers`.

#### `scraper/flixhq.go` — 14 testes

```
TestFlixHQ_SearchAnime_ParsesHTMLCorrectly
TestFlixHQ_SearchAnime_EmptyResult
TestFlixHQ_SearchAnime_NetworkError
TestFlixHQ_GetEpisodes_ParsesSeasonStructure
TestFlixHQ_GetEpisodes_MovieSingleEpisode
TestFlixHQ_GetStreamURL_ExtractsVideoSrc
TestFlixHQ_GetStreamURL_FallbackServer
TestFlixHQ_GetStreamURL_CaptchaDetection
TestFlixHQ_ParseSearchPage_MalformedHTML
TestFlixHQ_ParseEpisodePage_NoEpisodes
TestFlixHQ_ResolveServerURL_MultipleServers
TestFlixHQ_ExtractVidCloudURL_EncodedPayload
TestFlixHQ_GetType_ReturnsFlixHQType
TestFlixHQ_SearchAnime_SpecialCharacters
```

**Mock strategy:** `httptest.Server` servindo HTML estático capturado de respostas reais. Fixtures em `testdata/flixhq/`.

#### `scraper/sflix.go` — 12 testes

```
TestSFlix_SearchAnime_ParsesHTMLCorrectly
TestSFlix_SearchAnime_EmptyResult
TestSFlix_GetEpisodes_ParsesCorrectly
TestSFlix_GetStreamURL_ExtractsSource
TestSFlix_GetStreamURL_Blocked403
TestSFlix_ParseSearchResults_MalformedHTML
TestSFlix_DecryptPayload_ValidKey
TestSFlix_DecryptPayload_InvalidKey
TestSFlix_ExtractServerList_MultipleServers
TestSFlix_SelectBestServer_ByPriority
TestSFlix_GetType_ReturnsSFlixType
TestSFlix_HandleRateLimit_429Response
```

**Mock strategy:** Mesmo padrão FlixHQ. Fixtures em `testdata/sflix/`.

#### `scraper/unified.go` (funções restantes) — 12 testes

```
TestScraperManager_GetInstance_InitializesAll
TestScraperManager_GetScraperDisplayName_AllTypes
TestScraperManager_AllScrapersRegistered
TestScraperManager_GetStreamURL_DelegatesToCorrectScraper
TestScraperManager_GetEpisodes_DelegatesToCorrectScraper
TestScraperManager_SearchAnime_PerScraperTimeout
TestScraperManager_SearchAnime_PanicRecovery
TestScraperManager_TagResults_PTBRFirst
TestScraperManager_TagResults_MediaTypeDisambig
TestScraperManager_CleanPTBRTitle_AllPatterns
TestScraperManager_NeedsMediaTypeDisambig
TestScraperManager_SortPTBRFirst_StableOrder
```

#### `scraper/nineanime.go` — 5 testes

```
TestNineAnime_SearchAnime_ParsesAJAXResponse
TestNineAnime_GetEpisodes_ParsesEpisodeList
TestNineAnime_GetStreamURL_DecryptsPayload
TestNineAnime_GetType_ReturnsNineAnimeType
TestNineAnime_SearchAnime_CloudflareChallenge
```

#### `scraper/animedrive.go` — 5 testes

```
TestAnimeDrive_SearchAnime_ParsesHTML
TestAnimeDrive_GetEpisodes_ParsesLinks
TestAnimeDrive_GetStreamURL_ExtractsVideoURL
TestAnimeDrive_GetType_ReturnsAnimeDriveType
TestAnimeDrive_SearchAnime_EmptyResults
```

#### `scraper/allanime.go` + `animefire.go` + `goyabu.go` + `superflix.go` — 7 testes

```
TestAllAnime_GetType_ReturnsAllAnimeType
TestAllAnime_SearchAnime_GraphQLQuery
TestAnimeFire_SearchAnime_ParsesHTML
TestAnimeFire_GetStreamURL_ParsesVideoJSON
TestGoyabu_SearchAnime_ParsesHTML
TestGoyabu_GetStreamURL_ExtractsPlayer
TestSuperFlix_SearchAnime_ParsesHTML
```

#### `scraper/movie/flixhq.go` + `movie/sflix.go` — 5 testes

```
TestMovieFlixHQ_SearchMovies_ParsesResults
TestMovieFlixHQ_GetStreamURL_MovieType
TestMovieSFlix_SearchMovies_ParsesResults
TestMovieSFlix_GetStreamURL_MovieType
TestMovieFlixHQ_HandleNoResults
```

#### `scraper/source_diagnostic.go` (restantes) — 5 testes

```
TestDiagnoseError_NilReturnsNil
TestSourceDiagnostic_ErrorFormatting_AllBranches
TestSourceDiagnostic_Is_AllKinds
TestSourceDiagnostic_UserMessage_AllKinds
TestContainsAny_EdgeCases
```

---

### 3.2 🔴 `internal/player/` — 31 testes (~123 funções)

#### `player/player.go` — 10 testes

```
TestFilterMPVArgs_Whitelist
TestFilterMPVArgs_RejectsInjection
TestSanitizeMediaTarget_ValidURLs
TestSanitizeMediaTarget_RejectsUnsafeSchemes
TestSanitizeOutputPath_TraversalPrevention
TestMpvSendCommand_FormatsJSON
TestGetPlaybackTime_ParsesResponse
TestGetMediaDuration_ParsesResponse
TestBuildMPVCommand_DefaultArgs
TestBuildMPVCommand_WithShaders
```

#### `player/playvideo.go` — 8 testes

```
TestResolveSeasonForEpisode_WithSeasonMap
TestResolveSeasonForEpisode_SelectedSeason
TestResolveSeasonForEpisode_NoSeasonData
TestGetCurrentEpisode_ByNumber
TestGetCurrentEpisode_ByNum
TestGetCurrentEpisode_ByIndex
TestTrackingKey_Format
TestHandleDownloadAndPlay_RoutingDecision
```

#### `player/download.go` (restantes) — 7 testes

```
TestDownloadPart_ResumeWithRangeHeader
TestCombineParts_AllPresent
TestCombineParts_MissingPart
TestSafePartPath_TraversalAttempt
TestDownloadVideo_ZeroContentLength
TestGetContentLength_HeadRequest
TestSanitizeOutputPath_ValidPaths
```

#### `player/scraper.go` — 6 testes

```
TestExtractStreamURL_AllAnimeSource
TestExtractStreamURL_AnimeFireSource
TestExtractStreamURL_FlixHQSource
TestExtractStreamURL_FallbackChain
TestExtractStreamURL_NoSourceFound
TestResolveStreamWithRetry_TransientError
```

---

### 3.3 🔴 `internal/api/` — 25 testes (~78 funções)

#### `api/api.go` — SSRF (7 testes)

```
TestIsDisallowedIP_PrivateRanges
TestIsDisallowedIP_Loopback
TestIsDisallowedIP_PublicIPs
TestValidateExternalURL_SSRFVectors
TestValidateExternalURL_ValidURLs
TestSafeTransport_CreatesTransport
TestSafeDialContext_RejectsPrivateIP
```

#### `api/anime.go` (restantes) — 7 testes

```
TestParseAnimes_MockHTML
TestResolveURL_RelativeAndAbsolute
TestParseEpisodeNumber_AllFormats
TestSortEpisodesByNum_StableOrder
TestSearchAnime_BuildsCorrectURL
TestHTTPGetWithUA_SetsUserAgent
TestHTTPPostFast_SetsHeaders
```

#### `api/enhanced.go` — 5 testes

```
TestSanitizeFilename_SpecialChars
TestExtractMediaIDFromURL_AllPatterns
TestLanguagePriority_Ordering
TestSelectFlixHQQualityOptions_AllResolutions
TestGetNineAnimeStreamURL_ParsesResponse
```

#### `api/episodes.go` — 3 testes

```
TestGetAnimeEpisodes_ParsesPage
TestParseEpisodes_MalformedHTML
TestParseEpisodeNumber_EdgeCases
```

#### `api/aniskip.go` — 3 testes

```
TestRoundTime_AllCases
TestParseAniSkipResponse_ValidJSON
TestParseAniSkipResponse_EmptyResults
```

---

### 3.4 🔴 `internal/downloader/` — 15 testes (~77 funções)

#### `downloader/downloader.go` — 8 testes

```
TestDownloader_DownloadEpisode_RoutingBySource
TestDownloader_BuildOutputPath_SanitizesFilename
TestDownloader_DownloadWithProgress_UpdatesCallback
TestDownloader_HandleExistingFile_SkipsIfPresent
TestDownloader_DownloadWithRetry_TransientError
TestDownloader_SelectQuality_BestAvailable
TestDownloader_ValidateURL_RejectsEmpty
TestDownloader_ConcurrentDownloads_ThreadSafe
```

#### `downloader/movie_downloader.go` — 4 testes

```
TestMovieDownloader_DownloadMovie_FlixHQSource
TestMovieDownloader_BuildMoviePath_Sanitizes
TestMovieDownloader_HandleExistingMovie_Skips
TestMovieDownloader_DownloadMovie_NetworkError
```

#### `downloader/nineanime_downloader.go` — 3 testes

```
TestNineAnimeDownloader_Download_ParsesStream
TestNineAnimeDownloader_SelectQuality
TestNineAnimeDownloader_HandleError
```

---

### 3.5 🟡 `internal/util/` — 15 testes (~66 funções)

#### `util/util.go` — 7 testes

```
TestFileExists_ExistingFile
TestFileExists_MissingFile
TestEnsureDir_CreatesDirectory
TestSanitizePath_TraversalAttempt
TestGetHomeDir_ReturnsValid
TestExpandPath_TildeExpansion
TestFormatBytes_AllUnits
```

#### `util/httpclient.go` — 5 testes

```
TestNewHTTPClient_DefaultConfig
TestHTTPClient_Get_Success
TestHTTPClient_Get_Timeout
TestHTTPClient_SetHeaders
TestHTTPClient_FollowsRedirects
```

#### `util/perf.go` + `util/logger.go` — 3 testes

```
TestDebug_WritesToLog
TestDebug_NoOpWhenDisabled
TestPerfTimer_MeasuresElapsed
```

---

### 3.6 🟡 `internal/handlers/` + `internal/playback/` — 12 testes (~52 funções)

#### `handlers/media.go` — 5 testes

```
TestHandleMedia_RoutesAnimeToPlayer
TestHandleMedia_RoutesMovieToDownloader
TestHandleMedia_InvalidInput_ReturnsError
TestHandleMedia_MissingEpisodes_ReturnsError
TestHandleMedia_WithBatchFlag_RoutesBatch
```

#### `playback/series.go` + `playback/movie.go` — 4 testes

```
TestPlaySeries_BuildsCorrectArgs
TestPlaySeries_HandlesEpisodeNavigation
TestPlayMovie_SingleEpisode
TestPlayMovie_WithSubtitles
```

#### `handlers/download.go` + `handlers/upscale.go` — 3 testes

```
TestHandleDownload_ValidatesInput
TestHandleDownload_RoutesToCorrectDownloader
TestHandleUpscale_ValidatesShaderPath
```

---

### 3.7 🟡 `internal/upscaler/` — 8 testes (~47 funções)

```
TestAnime4K_FindShaders_DefaultPaths
TestAnime4K_FindShaders_CustomPath
TestAnime4K_BuildGLSLArgs_AllModes
TestAnime4K_ValidateShader_ExistingFile
TestAnime4K_ValidateShader_MissingFile
TestUpscalerVideo_BuildFFmpegArgs
TestUpscalerVideo_ValidateInputFile
TestUpscaler_GetAvailableShaders
```

---

### 3.8 🟡 Outros — 12 testes (~60 funções)

#### `models/media.go` — 3 testes

```
TestMedia_IsMovieOrTV_AllTypes
TestMedia_MediaType_Constants
TestMedia_DefaultValues
```

#### `api/providers/source_providers.go` — 4 testes

```
TestSourceProvider_ResolveForAnime
TestSourceProvider_ResolveForMovie
TestSourceProvider_FallbackChain
TestSourceProvider_AllProvidersRegistered
```

#### `pkg/goanime/client.go` — 3 testes

```
TestClient_NewClient_DefaultConfig
TestClient_Search_DelegatesToScraper
TestClient_GetEpisodes_DelegatesToScraper
```

#### `download/workflow.go` — 2 testes

```
TestWorkflow_ValidatesInputs
TestWorkflow_BuildsCorrectPipeline
```

---

## 4. Melhorias nos Testes Existentes

### 4.1 Testes lentos (> 5s)

| Pacote | Tempo | Causa | Fix |
|---|---|---|---|
| `scraper` | 26.6s | Timeouts reais de rede | Mock HTTP, usar `httptest` |
| `downloader/hls` | 16.9s | `time.Sleep` em retry | Injetar clock mockável |
| `player` + `player/test` | 9.3s | MPV process startup | Mockar IPC socket |

### 4.2 Singleton pollution

| Arquivo | Problema | Fix |
|---|---|---|
| `tracking/local_test.go` | `globalTracker` persiste entre testes | `t.Cleanup(CloseGlobalTracker)` |
| `scraper/unified_test.go` | `createTestManager` cria manager sem cleanup | Usar `t.Cleanup()` |

### 4.3 Cobertura parcial existente

| Arquivo | Cobertura | Funções a melhorar |
|---|---|---|
| `scraper/unified.go` | 42% → 70% alvo | `tagResults`, `cleanPTBRTitle`, `sortPTBRFirst` |
| `tracking/local.go` | 60% → 80% alvo | `UpdateProgress` edge cases, `migrateOldData` |
| `updater/updater.go` | 53% → 70% alvo | `CheckForUpdates`, `DownloadUpdate` |

---

## 5. Estratégia de Mock — Padrões Reutilizáveis

### Já existentes (reutilizar)

| Mock | Localização | Interface |
|---|---|---|
| `MockScraper` | `unified_test.go` | `UnifiedScraper` |
| `mockHLSCDN` | `download_test.go` | `httptest.Server` |
| `mockCDN` | `hls_test.go` | `httptest.Server` |

### Novos mocks necessários

| Mock | Para | Padrão |
|---|---|---|
| **HTML fixture server** | Scrapers (FlixHQ, SFlix, etc.) | `httptest.Server` + `testdata/*.html` |
| **GraphQL mock** | AllAnime scraper | `httptest.Server` com response JSON |
| **AniList mock** | `FetchAnimeFromAniList` | `httptest.Server` com GraphQL response |
| **MPV IPC mock** | `player.go` | `net.Pipe()` para socket Unix |
| **File system mock** | `downloader.go` | `t.TempDir()` + `os.Create` |

### Template para mock HTTP scraper

```go
func mockScraperHTTP(t *testing.T, fixtures map[string]string) *httptest.Server {
    t.Helper()
    return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        body, ok := fixtures[r.URL.Path]
        if !ok {
            w.WriteHeader(404)
            return
        }
        fmt.Fprint(w, body)
    }))
}
```

---

## 6. Prioridade de Implementação — Sprints

### Sprint 1: Segurança + Core (dias 1-3) — **+35 testes → ~35%**

| Arquivo | Testes | Funções |
|---|---|---|
| `api/api.go` (SSRF) | 7 | 7 |
| `player/player.go` (sanitize) | 10 | 40 |
| `player/download.go` (paths) | 7 | 27 |
| `api/anime.go` (parsing) | 7 | 19 |
| `api/episodes.go` | 3 | 4 |
| **Subtotal** | **34** | **97** |

### Sprint 2: Scrapers (dias 4-7) — **+65 testes → ~52%**

| Arquivo | Testes | Funções |
|---|---|---|
| `scraper/media_manager.go` | 15 | 60 |
| `scraper/unified.go` (restantes) | 12 | 45 |
| `scraper/flixhq.go` | 14 | 54 |
| `scraper/sflix.go` | 12 | 46 |
| `scraper/nineanime.go` | 5 | 21 |
| `scraper/animedrive.go` + outros | 7 | 44 |
| **Subtotal** | **65** | **270** |

### Sprint 3: Downloader + Player (dias 8-10) — **+28 testes → ~62%**

| Arquivo | Testes | Funções |
|---|---|---|
| `downloader/downloader.go` | 8 | 33 |
| `downloader/movie_downloader.go` | 4 | 28 |
| `downloader/nineanime_downloader.go` | 3 | 16 |
| `player/playvideo.go` | 8 | 31 |
| `player/scraper.go` | 6 | 25 |
| **Subtotal** | **29** | **133** |

### Sprint 4: Util + Handlers + Polish (dias 11-14) — **+55 testes → ~70%**

| Arquivo | Testes | Funções |
|---|---|---|
| `util/` (all) | 15 | 66 |
| `handlers/` (all) | 8 | 27 |
| `playback/` (all) | 4 | 25 |
| `upscaler/` (all) | 8 | 47 |
| `models/` | 3 | 13 |
| `api/providers/source_providers.go` | 4 | 41 |
| `api/enhanced.go` | 5 | 20 |
| `api/aniskip.go` | 3 | 4 |
| `scraper/movie/` | 5 | 42 |
| **Subtotal** | **55** | **285** |

---

## 7. Comandos de Verificação

```bash
# Cobertura total (validar progresso)
go test ./... -short -coverprofile=coverage.out -covermode=atomic
go tool cover -func=coverage.out | tail -1

# Cobertura por pacote
go test ./... -short -cover | sort -t: -k2 -n

# Funções ainda a 0%
go tool cover -func=coverage.out | grep "0\.0%" | wc -l

# HTML visual
go tool cover -html=coverage.out -o coverage.html

# Com race detector
go test ./... -short -race -count=1

# Benchmark
go test ./internal/scraper/ -bench=. -benchmem
```

---

## 8. Métricas Alvo

### Histórico (Sprint Inicial: FASES 1–14)

| Métrica | 2026-04-26 | Sprint 1 | Sprint 2 | Sprint 3 | Sprint 4 |
|---|---|---|---|---|---|
| Cobertura | 25.9% | ~35% | ~52% | ~62% | **~70%** (estimado) |
| Funções 0% | 1033 | ~936 | ~666 | ~533 | **~248** |
| Testes novos | 0 | +34 | +65 | +29 | +55 |
| Total acumulado | 0 | 34 | 99 | 128 | **183** |
| Pacotes 0% | 5 | 3 | 2 | 1 | **0** |

### Real Alcançado (2026-05-18)

| Métrica | Pós-FASE 14 |
|---|---|
| Cobertura total | **52.8%** |
| Statements totais | 12065 |
| Statements missed | 5692 |
| Pacotes a 0% | 0 (todos os pacotes do production têm pelo menos 5% cobertura) |
| Funções/testes criadas | ~983 ao longo de FASES 1–14 |

**Conclusão:** sprint inicial subiu de 25.9% → 52.8% (+26.9pp) em 14 fases. As fases finais (13–14) tiveram diminishing returns porque cada nova função adicionada cobria poucos statements (helpers triviais, accessors).

### Push 70% — Fases Adicionais Estritas (planejado 2026-05-18, corrigido para contagem real)

**Regra reafirmada (CLAUDE.md REGRA #0):** 1 teste por função. **165 funções a 0%** (contagem exata) → cobrir todas exceto ~30 intratáveis.

| Métrica | Pós-14 | Pós-15 | Pós-16 | Pós-17 |
|---|---:|---:|---:|---:|
| Cobertura % | 52.8 | ~58 | ~64 | **~70** |
| Stmts cobertos | 6373 | ~6973 | ~7873 | **~8473** |
| Funcs 0% | 165 | ~108 | ~53 | **≤ 30** |
| Funcs novas testadas | — | +57 | +55 | +53 |
| Refactors aceitos | — | 3 | 5–6 | 2–3 |

**Total funções novas testadas:** +165 → +1148 totais somando FASES 1–17.

**Importante:** A contagem inicial de "591 funcs 0%" estava errada — bug no `grep "0.0%"` que capturava também `100.0%`, `80.0%`, etc. Contagem correta via `awk '$NF == "0.0%"'` revela apenas 165 funcs realmente sem cobertura.

### Refactors permitidos (autorizado 2026-05-18, "vale tudo")

| Tipo | Exemplo | Quando |
|---|---|---|
| Var injetável | `var anilistBaseURL = "..."` | URL hardcoded |
| Interface wrap | `type rpcClient interface { ... }` | Cliente global de SDK externo |
| Split função | `Public()` chama `privateHelper(deps)` | Função orquestradora com TUI + rede |
| `*ForTesting` | `func SetClientForTesting(c iface)` | Singleton inicializado em init() |
| Dependency injection | `New*(opts ...Option)` | Construtores que liam env diretamente |

**Restrição única:** API pública não quebra. Funções exportadas pré-FASE 14 mantêm assinatura.

---

## 9. Sequenciamento das Fases 15–17

### Por que esta ordem?

1. **FASE 15 (api + util):** Maior número de funcs 0% num só conjunto (57). Refactors leves (vars de URL injetáveis). ROI alto, baixo risco.
2. **FASE 16 (playback + handlers + discord + upscaler + updater):** Refactors mais agressivos (interface wrap discord, splits de função orquestrada playback/handlers). Maior payoff de statements (+900 estimados) porque cobre paths longos antes destestados.
3. **FASE 17 (scraper + providers + downloader + SDK + misc):** Limpeza final. Refactors mínimos. Cobre funcs restantes em pacotes já bem testados.

### Critério de saída de cada fase

- `go test ./<pacote>/ -count=1 -race` verde
- Funcs 0% no pacote alvo caiu para meta (ver tabela FASE em TEST_STAGES.md)
- Refactors documentados em comentário `// 2026-05-XX: extracted for testability (CLAUDE.md REGRA #0)`
- Sem regressão: pacotes não-alvo mantêm cobertura ≥ valor anterior

### Critério final FASE 17

```bash
go test ./... -short -coverprofile=coverage.out -covermode=atomic -race
go tool cover -func=coverage.out | tail -1
# Esperado: ≥ 70.0%
go tool cover -func=coverage.out | awk '$NF == "0.0%"' | wc -l
# Esperado: ≤ 30 (apenas main()/exemplos/loops TUI puros)
```

### Verificação contínua (em CADA fase)

```bash
# Antes da fase X — snapshot de funcs 0% no pacote
go tool cover -func=cov_pre.out | awk '$NF == "0.0%" {print $1, $2}' | grep "internal/<pkg>" > pre.txt
wc -l pre.txt

# Após a fase X — comparar
go tool cover -func=cov_post.out | awk '$NF == "0.0%" {print $1, $2}' | grep "internal/<pkg>" > post.txt
wc -l post.txt

# Funções ainda 0% (devem ser justificadas em comentário do PR ou listadas como exceção)
diff pre.txt post.txt
```

**ATENÇÃO:** Use `awk '$NF == "0.0%"'`, NÃO `grep "0.0%"`. O grep também matches "100.0%", "80.0%", "70.0%", etc. → contagem inflada.
