# GoAnime — Estratégia de Testes por Tipo

> **1217 funções** · **1033 sem cobertura** · Guia de qual tipo de teste usar em cada caso

---

## Taxonomia de Testes

| Tipo | Quando Usar | Padrão Go |
|---|---|---|
| **Unitário Puro** | Função sem I/O, sem estado externo, lógica determinística | `assert.Equal(t, expected, fn(input))` |
| **Unitário com Mock** | Função que chama HTTP/DB/FS mas lógica interna é o alvo | `httptest.Server` / interface mock |
| **Integração (Cascata)** | Fluxo que atravessa 2+ camadas (API→Scraper→Parser) | `httptest` end-to-end com fixtures HTML |
| **Lógica de Estado** | Funções com mutex, circuit breaker, cache TTL | `t.Parallel()` + `sync/atomic` + `-race` |
| **Segurança** | Validação de input, SSRF, path traversal, injection | Table-driven com vetores maliciosos |
| **Stress/Concorrência** | Worker pools, download paralelo, singleton init | `-race -count=100` + goroutines concorrentes |

---

## 1. SEGURANÇA — Unitário Puro + Vetores Maliciosos

### `internal/api/api.go` (7 funções a 0%)

| Função | Tipo de Teste | Justificativa |
|---|---|---|
| `IsDisallowedIP` | **Unitário puro** — table-driven | Lógica pura: IP string → bool. Sem I/O. Testar com vetores SSRF |
| `checkDisallowedIP` | **Unitário com mock** | Recebe `net.Conn` — mockar com `net.Pipe()` |
| `dialFunc` | **Unitário com mock** | Wrapper de `net.Dialer` — mockar DNS resolver |
| `SafeTransport` | **Unitário puro** | Retorna `*http.Transport` — validar campos TLS/Dial |
| `SafeGet` | **Integração** | Chama HTTP real — usar `httptest.Server` |
| `ValidateExternalURL` | **Unitário puro** — table-driven | URL string → error. Vetores: `127.0.0.1`, `[::1]`, `0x7f000001` |
| `SafeDialContext` | **Unitário com mock** | Dial + IP check — mockar DNS |

```go
// Exemplo: IsDisallowedIP — UNITÁRIO PURO
func TestIsDisallowedIP(t *testing.T) {
    t.Parallel()
    tests := []struct{ ip string; want bool }{
        {"127.0.0.1", true}, {"::1", true}, {"10.0.0.1", true},
        {"192.168.1.1", true}, {"0.0.0.0", true}, {"", true},
        {"8.8.8.8", false}, {"1.1.1.1", false},
    }
    for _, tt := range tests {
        assert.Equal(t, tt.want, IsDisallowedIP(tt.ip), "IP: %s", tt.ip)
    }
}
```

### `internal/scraper/ssrf.go` (3 funções)

| Função | Tipo | Justificativa |
|---|---|---|
| `isPrivateIP` | **Unitário puro** | Mesma lógica que `IsDisallowedIP` |
| `validateURL` | **Unitário puro** | URL → error |
| `safeHTTPClient` | **Unitário puro** | Valida config do client retornado |

### `internal/api/movie/ssrf.go` (3 funções)

| Função | Tipo | Justificativa |
|---|---|---|
| Todas | **Unitário puro** | Duplicação da validação SSRF para movie API |

---

## 2. LÓGICA PURA — Unitário Puro (Table-Driven)

Funções sem efeitos colaterais. Input → Output determinístico.

### `internal/api/anime.go`

| Função | Tipo | Input → Output |
|---|---|---|
| `CleanTitle` | ✅ Já testado | string → string |
| `normalizeAccents` | ✅ Já testado | string → string |
| `generateSearchVariations` | ✅ Já testado | string → []string |
| `getStringValue` | **Unitário puro** | `map[string]any` → string |
| `getIntValue` | **Unitário puro** | `map[string]any` → int |
| `getBoolValue` | **Unitário puro** | `map[string]any` → bool |
| `resolveURL` | **Unitário puro** | (base, relative) → absolute URL |
| `parseEpisodeNumber` | **Unitário puro** | string → int |
| `safeClose` | **Unitário puro** | `io.Closer` → chamada sem panic |

### `internal/api/enhanced.go`

| Função | Tipo | Input → Output |
|---|---|---|
| `sanitizeFilename` | **Unitário puro** | `"[PT-BR] One/Piece"` → `"One_Piece"` |
| `extractMediaIDFromURL` | **Unitário puro** | URL → ID string |
| `languagePriority` | **Unitário puro** | nome com tag → int (ordem) |
| `isStdoutTerminal` | **Unitário puro** | Retorna bool cached |

### `internal/api/allanime_smart.go`

| Função | Tipo | Input → Output |
|---|---|---|
| `sanitizeSmart` | **Unitário puro** | string → sanitized string |
| `sanitizeSmartDest` | **Unitário puro** | path → safe path |
| `validateSmartRangeInputs` | **Unitário puro** | (start, end) → error |
| `shouldUseYtDlp` | **Unitário puro** | error → bool |
| `isUnsafeExtensionError` | **Unitário puro** | error → bool |
| `alreadyDownloaded` | **Unitário puro** | path → bool (via `os.Stat`) |
| `smartOutputDir` | **Unitário puro** | anime → dir path |

### `internal/api/aniskip.go`

| Função | Tipo | Input → Output |
|---|---|---|
| `RoundTime` | **Unitário puro** | float64 → int |
| `ParseAniSkipResponse` | **Unitário puro** | JSON bytes → struct |

### `internal/models/media.go` (9 funções)

| Função | Tipo | Input → Output |
|---|---|---|
| `IsAnime`, `IsMovie`, `IsTV`, `IsMovieOrTV` | **Unitário puro** | Media → bool |
| `GetDisplayName` | **Unitário puro** | Media → string formatado |
| `OfficialTitle` | **Unitário puro** | Media com TMDB/AniList → melhor título |
| `GetRatingDisplay` | **Unitário puro** | float64 → "★ 8.5" |
| `GetGenresDisplay` | **Unitário puro** | []string → "Action, Drama" |
| `GetRuntimeDisplay` | **Unitário puro** | int → "2h 15m" |

### `internal/util/util.go` (funções puras)

| Função | Tipo | Input → Output |
|---|---|---|
| `SetGlobalSubtitles` / `ClearGlobalSubtitles` | **Unitário puro** | set/get global state |
| `SetGlobalReferer` / `GetGlobalReferer` / `ClearGlobalReferer` | **Unitário puro** | set/get/clear |
| `SetGlobalAnimeSource` / `GetGlobalAnimeSource` / `Is9AnimeSource` | **Unitário puro** | source state |
| `TreatingAnimeName` | **Unitário puro** | string → sanitized |
| `stripTrailingAnimeMetadata` | **Unitário puro** | string → string |
| `BuildMediaFolderName` / `BuildMediaFileName` | **Unitário puro** | params → path string |
| `DefaultDownloadDir` / `DefaultMovieDownloadDir` | **Unitário puro** | → default path |
| `FormatPlexMovieDir` / `FormatPlexEpisodePath` / `FormatPlexEpisodeDir` | **Unitário puro** | params → Plex path |
| `RegisterCleanup` / `RunCleanup` | **Lógica de estado** | callback registration |
| `ErrorHandler` | **Unitário puro** | error → formatted error |

### `internal/version/version.go` (2 funções)

| Função | Tipo |
|---|---|
| `HasVersionArg` | **Unitário puro** — `os.Args` → bool |
| `ShowVersion` | **Unitário puro** — prints to stdout |

### `pkg/goanime/types/` (7 funções)

| Função | Tipo |
|---|---|
| `FromInternalAnime`, `FromInternalAnimeList`, `FromInternalEpisode`, `FromInternalEpisodeList` | **Unitário puro** — struct conversion |
| `String`, `ToScraperType`, `ParseSource` | **Unitário puro** — enum mapping |

---

## 3. UNITÁRIO COM MOCK — Funções com I/O Externo

Funções que fazem HTTP, leem FS, ou chamam processos externos. Mock a dependência, teste a lógica.

### `internal/scraper/*.go` — Scrapers (220 funções)

**Padrão universal para scrapers:**

```go
// TODOS os scrapers seguem este padrão: httptest.Server + HTML fixture
func TestFlixHQ_SearchAnime(t *testing.T) {
    fixture := loadFixture(t, "testdata/flixhq/search_results.html")
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, fixture)
    }))
    defer srv.Close()

    client := NewFlixHQClient()
    client.baseURL = srv.URL // injetar URL do mock
    results, err := client.SearchAnime("test")
    require.NoError(t, err)
    assert.Len(t, results, 5)
}
```

| Arquivo | Funções | Tipo | Mock |
|---|---|---|---|
| `flixhq.go` (54) | `SearchAnime`, `GetEpisodes`, `GetStreamURL`, etc. | **Unit + httptest** | HTML fixtures por endpoint |
| `sflix.go` (46) | Mesmas operações | **Unit + httptest** | HTML fixtures |
| `nineanime.go` (21) | AJAX + decrypt | **Unit + httptest** | JSON responses |
| `allanime.go` (14) | GraphQL queries | **Unit + httptest** | GraphQL JSON responses |
| `animefire.go` (8) | HTML parsing | **Unit + httptest** | HTML fixtures |
| `animedrive.go` (21) | HTML parsing | **Unit + httptest** | HTML fixtures |
| `goyabu.go` (7) | HTML parsing | **Unit + httptest** | HTML fixtures |
| `superflix.go` (9) | HTML + API | **Unit + httptest** | HTML + JSON |
| `media_manager.go` (60) | Orquestração | **Unit + MockScraper** | Reutilizar `MockScraper` existente |

### `internal/player/player.go` (40 funções)

| Função | Tipo | Mock |
|---|---|---|
| `filterMPVArgs` | **Unitário puro** | Nenhum — []string → []string |
| `sanitizeMediaTarget` | **Unitário puro** | URL → (safe URL, error) |
| `sanitizeOutputPath` | **Unitário puro** | path → (safe path, error) |
| `buildMPVCommand` | **Unitário puro** | params → exec.Cmd args |
| `mpvSendCommand` | **Unit + mock** | `net.Pipe()` para IPC socket |
| `getPlaybackTime` | **Unit + mock** | Mockar socket response |
| `StartVideo` | **Integração** | Não testável sem mpv — skip em CI |

### `internal/downloader/` (77 funções)

| Arquivo | Tipo | Mock |
|---|---|---|
| `downloader.go` — `DownloadEpisode` | **Unit + httptest** | Mock CDN servindo segmentos |
| `downloader.go` — `BuildOutputPath` | **Unitário puro** | params → path |
| `movie_downloader.go` — `DownloadMovie` | **Unit + httptest** | Mock CDN |
| `nineanime_downloader.go` | **Unit + httptest** | Mock 9anime API responses |

### `internal/api/anime.go` — Funções com HTTP

| Função | Tipo | Mock |
|---|---|---|
| `SearchAnime` | **Unit + httptest** | Mock AllAnime HTML page |
| `ParseAnimes` | **Unit + mock** | `goquery.NewDocumentFromReader(strings.NewReader(html))` |
| `FetchAnimeFromAniList` | **Unit + httptest** | Mock AniList GraphQL response |
| `httpGetWithUA` | **Unit + httptest** | Validar User-Agent header |
| `httpPostFast` | **Unit + httptest** | Validar POST body/headers |
| `enrichAnimeData` | **Unit + httptest** | Mock Jikan/AniList APIs |

---

## 4. LÓGICA DE ESTADO — Mutex, Cache, Circuit Breaker

### `internal/scraper/source_circuit.go` (7 funções)

| Função | Tipo | Padrão |
|---|---|---|
| `RecordFailure` | **Lógica de estado** | Incrementar failures, verificar threshold |
| `RecordSuccess` | **Lógica de estado** | Resetar failures counter |
| `ShouldSkip` | **Lógica de estado** | Verificar cooldown com `time.Now()` |
| `IsOpen` | **Lógica de estado** | Estado do circuit |

```go
func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
    cb := newSourceCircuitBreaker()
    cb.threshold = 3
    for i := 0; i < 3; i++ {
        cb.RecordFailure()
    }
    assert.True(t, cb.IsOpen())
    assert.True(t, cb.ShouldSkip())
}

func TestCircuitBreaker_RecoverAfterCooldown(t *testing.T) {
    cb := newSourceCircuitBreaker()
    cb.threshold = 2
    cb.cooldown = 10 * time.Millisecond
    cb.RecordFailure()
    cb.RecordFailure()
    assert.True(t, cb.ShouldSkip())
    time.Sleep(20 * time.Millisecond)
    assert.False(t, cb.ShouldSkip()) // half-open
}
```

### `internal/scraper/source_diagnostic.go` (14 funções)

| Função | Tipo |
|---|---|
| `DiagnoseError` | **Unitário puro** — error → DiagnosticKind |
| `Error()`, `UserMessage()` | **Unitário puro** — formatting |
| `Is()` | **Unitário puro** — sentinel matching |
| `containsAny` | **Unitário puro** — string helper |

### `internal/util/httpclient.go` — Cache e Worker Pool

| Componente | Tipo | Padrão |
|---|---|---|
| `ResponseCache.Get/Set` | **Lógica de estado** | Set → Get → assert data. Set expired → Get → assert miss |
| `ResponseCache.cleanup` | **Lógica de estado** | Set + sleep(>maxAge) + cleanup → verify deleted |
| `WorkerPool.Submit/Wait` | **Stress** | 100 tasks, verify all complete, `-race` |
| `ParallelExecute` | **Stress** | Atomic counter incremented by N tasks |

```go
func TestResponseCache_Expiry(t *testing.T) {
    cache := NewResponseCache(50*time.Millisecond, 10)
    cache.Set("key", []byte("value"))
    data, ok := cache.Get("key")
    assert.True(t, ok)
    assert.Equal(t, []byte("value"), data)
    time.Sleep(60 * time.Millisecond)
    _, ok = cache.Get("key")
    assert.False(t, ok, "entry should have expired")
}

func TestWorkerPool_ConcurrentExecution(t *testing.T) {
    var counter atomic.Int32
    pool := NewWorkerPool(5)
    for i := 0; i < 50; i++ {
        pool.Submit(func() { counter.Add(1) })
    }
    pool.Wait()
    assert.Equal(t, int32(50), counter.Load())
}
```

### `internal/tracking/local.go`

| Função | Tipo | Padrão |
|---|---|---|
| `GetGlobalTracker` | **Lógica de estado** | Singleton — verify same instance |
| `CloseGlobalTracker` | **Lógica de estado** | Close + verify nil |
| `migrateOldData` | **Integração** | Create old schema DB → migrate → verify |

---

## 5. INTEGRAÇÃO (CASCATA) — Fluxos Multi-Camada

Testes que validam o pipeline completo. Mockar apenas a camada de rede.

### Pipeline: Search → Select → Episodes → Stream

```
SearchAnimeEnhanced → ScraperManager.SearchAnime → [AllAnime, AnimeFire, FlixHQ]
       ↓
GetAnimeEpisodesEnhanced → scraper.GetAnimeEpisodes → HTML parser
       ↓
GetEpisodeStreamURL → scraper.GetStreamURL → URL resolution
```

| Teste Cascata | Camadas | Mock |
|---|---|---|
| `TestSearchToStream_AllAnime` | enhanced → scraper → allanime | `httptest` com GraphQL fixture |
| `TestSearchToStream_AnimeFire` | enhanced → scraper → animefire | `httptest` com HTML fixture |
| `TestSearchToStream_FlixHQ` | enhanced → scraper → flixhq | `httptest` com HTML fixture |
| `TestSearchToStream_9Anime` | enhanced → scraper → nineanime | `httptest` com JSON fixture |

```go
// Exemplo: Teste de integração cascata
func TestSearchToEpisodes_FlixHQ_Integration(t *testing.T) {
    // Layer 1: Mock search results
    searchSrv := httptest.NewServer(/* search HTML fixture */)
    // Layer 2: Mock episode list
    episodeSrv := httptest.NewServer(/* episode HTML fixture */)
    // Layer 3: Mock stream URL
    streamSrv := httptest.NewServer(/* stream JSON fixture */)

    client := NewFlixHQClient()
    client.baseURL = searchSrv.URL

    // Cascata completa: search → episodes → stream
    results, _ := client.SearchAnime("Breaking Bad")
    episodes, _ := client.GetAnimeEpisodes(results[0].URL)
    streamURL, _, _ := client.GetStreamURL(episodes[0].URL, "1080")

    assert.NotEmpty(t, streamURL)
    assert.Contains(t, streamURL, ".m3u8")
}
```

### Pipeline: Download Workflow

```
HandleDownloadRequest → SearchAnimeWithRetry → GetAnimeEpisodesEnhanced
       ↓
HandleBatchDownload → getBestQualityURL → downloadDirectHTTP/downloadHLS
       ↓
HLS: fetchPlaylist → selectBestStream → downloadSegments → combineParts
```

| Teste Cascata | Camadas | Mock |
|---|---|---|
| `TestDownloadWorkflow_SingleEpisode` | workflow → api → scraper → downloader | Todos mockados |
| `TestDownloadWorkflow_BatchRange` | workflow → batch → HLS | `httptest` CDN com playlist |
| `TestDownloadWorkflow_MovieRoute` | workflow → movie detection → movie downloader | Mock FlixHQ |
| `TestDownloadWorkflow_9AnimeRoute` | workflow → 9anime detection → nineanime downloader | Mock 9anime |

### Pipeline: Playback Series

```
HandleSeries → SelectInitialEpisode → PlayEpisode
       ↓
PlayEpisode → GetEpisodeStreamURL → StartVideo (MPV)
       ↓
Navigation: handleUserNavigationEnhanced → next/prev/select
```

| Teste Cascata | Camadas |
|---|---|
| `TestHandleSeries_NavigationLoop` | series → episode selection → navigation logic |
| `TestHandleSeries_ChangeAnime` | series → change anime → new episodes |
| `TestHandleSeries_BackToAnimeSelection` | series → back error propagation |

> ⚠️ **MPV é boundary do sistema.** Testes cascata param ANTES de `StartVideo`. Mock o IPC socket para testar tracking/progress.

---

## 6. STRESS / CONCORRÊNCIA

| Componente | Teste | Padrão |
|---|---|---|
| `ScraperManager.SearchAnime` | 10 goroutines buscando simultâneo | `-race` + verify no panic |
| `WorkerPool` | 100 tasks com pool de 5 | Atomic counter + `-race` |
| `ParallelExecute` | Edge: 0 tasks, 1 task, tasks > workers | Verify completion |
| `ResponseCache` | Concurrent Get/Set | `-race -count=100` |
| `LocalTracker` | 10 goroutines CRUD | `-race` + verify data integrity |
| `HLS downloadSegments` | 50+ segments simultâneos | Mock CDN + verify order |
| Global singletons | `GetSharedClient`, `GetFastClient` etc. | `sync.Once` thread-safety |

```go
func TestScraperManager_ConcurrentSearch(t *testing.T) {
    manager := createTestManager(t)
    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            _, _ = manager.SearchAnime("test", nil)
        }()
    }
    wg.Wait() // No race, no panic = pass
}
```

---

## 7. NÃO TESTÁVEL EM CI — Exclusões

| Componente | Razão | Alternativa |
|---|---|---|
| `player.StartVideo` | Requer binário `mpv` | Tag `//go:build integration` |
| `discord/discord.go` | Requer Discord RPC daemon | Mock interface `DiscordClient` |
| `tui.Find` (fuzzyfinder) | Requer terminal interativo | Testar lógica antes/depois do Find |
| `util.getUserInput` | Lê stdin | Injetar `io.Reader` |
| `cmd/goanime/main.go` | Entry point | Testar via `os/exec` em integration |
| `upscaler/video.go` — `UpscaleVideo` | Requer FFmpeg + GPU | Tag `//go:build ffmpeg` |
| `PreWarmConnections` | Faz requests reais | Skip ou mock DNS |

---

## 8. RESUMO — Decisão por Pacote

| Pacote | Unitário Puro | Unit + Mock | Cascata | Estado | Stress | Não Testável |
|---|---|---|---|---|---|---|
| `api/api.go` | 3 | 3 | 1 | — | — | — |
| `api/anime.go` | 5 | 4 | — | — | — | 1 (`selectAnime`) |
| `api/enhanced.go` | 4 | — | 3 | — | — | 2 (spinner, fuzzy) |
| `api/allanime_smart.go` | 7 | 2 | 1 | — | — | — |
| `api/aniskip.go` | 2 | 1 | — | — | — | — |
| `scraper/unified.go` | 12 | 5 | 2 | 3 | 2 | — |
| `scraper/flixhq.go` | — | 14 | 1 | — | — | — |
| `scraper/sflix.go` | — | 12 | 1 | — | — | — |
| `scraper/media_manager.go` | — | 10 | 2 | 1 | 1 | — |
| `scraper/source_circuit.go` | — | — | — | 5 | 1 | — |
| `scraper/source_diagnostic.go` | 8 | — | — | — | — | — |
| `player/player.go` | 5 | 3 | — | — | — | 2 (mpv) |
| `player/playvideo.go` | 6 | — | 1 | — | — | 1 (StartVideo) |
| `player/download.go` | 4 | 5 | 2 | — | 1 | — |
| `downloader/hls` | 1 | 2 | — | — | 1 | — |
| `downloader/*.go` | 2 | 8 | 2 | — | 1 | — |
| `handlers/media.go` | 2 | — | 1 | — | — | 4 (TUI) |
| `playback/series.go` | — | — | 2 | — | — | 3 (TUI) |
| `download/workflow.go` | 1 | — | 3 | — | — | 2 (TUI) |
| `tracking/local.go` | 1 | 3 | 1 | 1 | 1 | — |
| `util/httpclient.go` | 2 | — | — | 4 | 2 | 1 (PreWarm) |
| `util/perf.go` | 4 | — | — | 3 | — | — |
| `util/logger.go` | 5 | — | — | — | — | — |
| `util/util.go` | 15 | — | — | 1 | — | 3 (TUI) |
| `models/media.go` | 9 | — | — | — | — | — |
| `upscaler/*.go` | 5 | 2 | — | — | — | 6 (FFmpeg) |
| `pkg/goanime/` | 4 | 3 | 1 | — | — | — |

### Totais por Tipo

| Tipo | Quantidade | % do Total |
|---|---|---|
| **Unitário Puro** | ~100 | 55% |
| **Unitário com Mock** | ~45 | 25% |
| **Integração (Cascata)** | ~15 | 8% |
| **Lógica de Estado** | ~12 | 7% |
| **Stress/Concorrência** | ~8 | 4% |
| **Não testável em CI** | ~20 | (excluído) |
| **TOTAL** | **~183** | **100%** |

---

## 8.5 Push 70% — Estratégia Estrita "1 Teste por Função" (FASES 15–20, 2026-05-18)

Pós-FASE 14: **52.8%** total · **591 funções a 0%**. Para alcançar **≥ 70%** com **eficácia brutal** e máxima precisão, mantemos a regra dura da CLAUDE.md.

### Princípios (REGRA #0 reafirmada)

1. **CADA função listada em `go tool cover -func | grep "0.0%"` recebe seu próprio `TestNomeDaFuncao_Cenario`.** Sem agrupamento. Sem "uma tabela cobre cinco funções". Cada `func X()` no código produção → `func TestX*()` em arquivo de teste.
2. **Refactor amplamente permitido para tornar testável.** Usuário autoriza "vale tudo para software profissional". Restrição única: **API pública não quebra** (semver-friendly). Permitido:
   - `var fooBaseURL = "..."` substituível em teste
   - Interface wrap para globals (`type rpcClient interface { ... }`)
   - Split de função orquestrada em wrapper público + helper privado testável
   - `*ForTesting` setters (mas ≤ 2 por pacote para não poluir)
   - Receber `httpClient`, `logger`, etc. como dependência opcional
3. **Table-driven é técnica, não atalho.** Tabelas continuam OK *dentro de uma única função `TestX`* para cobrir múltiplos cenários da mesma função. Mas NÃO use uma `TestFoo` para cobrir as funções `A`, `B`, `C` — cada uma precisa de seu próprio `TestA`, `TestB`, `TestC`.
4. **Fixtures em testdata/.** M3U8/VTT/SRT/zip/HTML ficam em `internal/<pkg>/testdata/`. Carregar via `loadFixture(t, path)`.
5. **Cobertura é métrica secundária; nº de funções 0% é métrica primária.** O alvo final é "≤ 50 funções a 0%" — a cobertura ≥ 70% surge como consequência.

### Padrões novos

#### Pipe IPC (substitui rede real para MPV/sockets)
```go
func newTestMPVPipe(t *testing.T) (clientEnd, serverEnd net.Conn) {
    t.Helper()
    c, s := net.Pipe()
    t.Cleanup(func() { _ = c.Close(); _ = s.Close() })
    return c, s
}
```

#### Injeção mínima por package var
```go
// In production code:
var anilistBaseURL = "https://graphql.anilist.co"
// In test:
func TestX(t *testing.T) {
    orig := anilistBaseURL
    t.Cleanup(func() { anilistBaseURL = orig })
    srv := httptest.NewServer(handler)
    t.Cleanup(srv.Close)
    anilistBaseURL = srv.URL
    // ...
}
```

#### Zip server in-memory para downloader/upscaler
```go
func zipServer(t *testing.T, files map[string]string) *httptest.Server {
    t.Helper()
    var buf bytes.Buffer
    w := zip.NewWriter(&buf)
    for name, content := range files {
        f, _ := w.Create(name)
        _, _ = f.Write([]byte(content))
    }
    _ = w.Close()
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "application/zip")
        _, _ = w.Write(buf.Bytes())
    }))
    t.Cleanup(srv.Close)
    return srv
}
```

#### Interface wrap para globais Discord
```go
type rpcClient interface {
    Login() error
    Logout() error
    SetActivity(activity client.Activity) error
}
// Tests inject via:
func SetClientForTesting(c rpcClient) { discordClient = c }
```

### Alvos por FASE (estrito 1 teste/função)

| Fase | Pacotes | Funcs 0% | Stmts alvo |
|---|---|---:|---:|
| 15 | `player/` | 77 | +500 |
| 16 | `api/`, `util/` | 147 | +600 |
| 17 | `scraper/`, `api/providers/...`, `api/source/` | 132 | +400 |
| 18 | `api/movie/`, `api/providers/`, `downloader/`, `downloader/hls/` | 67 | +250 |
| 19 | `upscaler/`, `discord/`, `updater/` | 80 | +350 |
| 20 | `handlers/`, `playback/`, `models/`, `pkg/goanime/`, `tui/`, `tracking/`, `version/`, `download/`, `appflow/` | 88 | +200 |
| **TOTAL** | | **591** | **+2300** |

### Anti-padrões a evitar

- ❌ Mockar TUI (huh, fuzzyfinder, bubbletea). Em vez disso: refactor para extrair lógica testável antes do widget interativo.
- ❌ Testar `Run*Player()`, `HandleSeries()`, `Start()` (loops com IPC real) **sem refactor**. Refator extrai helper testável → testar helper.
- ❌ Aumentar cobertura via testes triviais (`assert.NotNil(NewX())`) — cada teste deve verificar **comportamento observável**.
- ❌ Pular função sem teste — **PROIBIDO** pela CLAUDE.md. Se não consegue testar, faça refactor.
- ❌ `*ForTesting` setters em excesso. Limite ≤ 2 por pacote; preferir variável injetável ou interface.

### Tabela de Decisão: "Como testar essa função?"

| Tipo da função | Estratégia | Refactor necessário? |
|---|---|:---:|
| Pure (sem I/O, sem estado) | Table-driven `TestX_*` com 5–12 casos | ❌ |
| Globals/accessor | `t.Cleanup` para resetar + verificar getter | ❌ |
| HTTP client method | `httptest.Server` + injeção de `baseURL` ou `*http.Client` | Talvez |
| Função com global `*Client` | Interface wrap + `var fooFactory = func(...) iface { ... }` | ✅ pequeno |
| Função orquestradora (rede + TUI + outros) | Split em wrapper + helper testável | ✅ pequeno |
| MPV IPC | `net.Pipe()` + injeção de `mpvSendCommand func` | ❌ (já injetado) |
| Download de zip/tarball remoto | `httptest.Server` + zip in-memory + `t.TempDir()` | ❌ ou pequeno |
| `main()` / exemplos | Não testar (exceção CLAUDE.md) | n/a |
| Loop interativo (`Run*Player`, `HandleSeries`) | Não testar a função; extrair helpers internos | ✅ |

---

## 9. Regras de Ouro

1. **Se não tem I/O → Unitário Puro.** Sem mock, sem server, sem fixture. Rápido.
2. **Se faz HTTP → `httptest.Server`.** NUNCA rede real em CI.
3. **Se tem mutex/atomic → Rodar com `-race`.** Sempre.
4. **Se atravessa 2+ camadas → Cascata.** Mas mockar a rede.
5. **Se depende de binário externo (mpv, ffmpeg) → Build tag.** `//go:build integration`
6. **Se depende de terminal interativo → Não testar a UI.** Testar a lógica antes e depois.
7. **Table-driven SEMPRE.** Um `[]struct` com 10 casos > 10 funções separadas.
8. **`t.Parallel()` em TUDO.** Exceto testes que manipulam globals (marcar explícito).
