# GoAnime — Migração de Arquitetura por Fases (Model B → hook para C)

> **LEIA ISTO PRIMEIRO** ao continuar a migração de arquitetura.
> **Decisão de referência:** `docs/ARCHITECTURE.md` §4 (Recommendation), §5 (Migration plan), §6 (Repo layout), §7 (prior art).
> **Meta:** colapsar os três dispatchers vivos (`player/scraper.go`, `api/enhanced.go`,
> `scraper/unified.go`) em um único caminho declarativo (`source.Resolve` →
> `Source.FetchStreamURL(ctx, …)`), sem quebrar comportamento visível ao usuário.

---

## Como usar este documento entre sessões

1. Abra este arquivo, ache a primeira ETAPA ⬜.
2. Anuncie: "Executando ETAPA X.Y — [nome]".
3. Releia o(s) arquivo(s)-alvo da etapa **no estado atual** (pode ter mudado desde
   que este plano foi escrito) antes de editar.
4. Execute a etapa inteira nesta sessão. Se for grande demais (ex.: 3.3 SuperFlix),
   faça o máximo e marque 🔄; a próxima sessão continua dali.
5. Rode a verificação da etapa (build + testes). Só marque ✅ se passar.
6. Atualize o checkbox e a seção "Status atual" no fim deste arquivo.
7. **Não pule etapas** — cada uma assume que as anteriores já estão no código.
   Se pular, a árvore de decisão (§5 do ARCHITECTURE.md) para de valer.
8. Nenhuma etapa deve deixar o build vermelho ou a suíte quebrada ao final.
   `go build ./...` e `go test ./... -short -race` são o mínimo antes de marcar ✅.

Símbolos: ⬜ não iniciada · 🔄 parcial (continuar na próxima sessão) · ✅ completa.

## Conformidade com docs/ARCHITECTURE.md — REGRA

**A recomendação (§4) é lei:** Model B agora, capacidades do Model C só quando um
source real precisar (SuperFlix primeiro). O estado FINAL do código deve bater
com o §2 Model B **exatamente** — interface, assinaturas e nomes. Estados
intermediários podem diferir apenas no que o próprio §5 autoriza ("the new path
can shadow the old until the final deletion").

Mapeamento etapa ↔ step do §5 (Migration plan):

| Etapa deste plano | Step do §5 | Observação |
|---|---|---|
| 0.1 | Phase 0, step 1 | `Source`+`Descriptor`+`Priority`+`Register`/`Resolve` scanning `Describe()` |
| 0.2 | Phase 0, step 2 | providers de `source_providers.go` → self-describing, um por vez |
| 1.1 | Phase 1, step 4 | ctx ponta a ponta (executado antes do step 3 para isolar risco; o estado ao FIM da Fase 1 é idêntico ao do doc) |
| 1.2 | Phase 1, step 3 | wrapper fino `Resolve(anime)` → `FetchStreamURL(ctx, …)` |
| 2.1 | Phase 2, step 5 | `ResolveURL` + source; deletar síntese fake-AllAnime |
| 2.2 | Phase 2, step 6 | `Unknown` explícito; best-effort só se configurado |
| 3.1 | Phase 3, step 7 | deletar 3 dispatchers + **rename para nomes exatos do §2** |
| 3.2–3.5 | Phase 3, step 8 (§6.6 ordem 1→5) | netx → superflix → demais → rename manager → doc.go |
| 4.1 | Phase 4, step 9 | `Seasoned`/`BrowserGated` só no SuperFlix, via type-assert |
| 5.1 (backlog) | §9 punch-list item 8 (S1–S3) | fora do §5; não bloqueia a recomendação |

Nota sobre o exemplo `Register(&Source{client: NewClient()})` do §2: a construção
no `init()` deve ser **barata** (struct). Máquina pesada (browser, ScraperManager)
inicializa lazy no primeiro uso — é o que o R6 ("built once and reused, lazily")
e o §1a exigem. O snippet é ilustrativo; os requisitos §0 são normativos.

---

## FASE 0 — Fundação (não toca o caminho ao vivo, risco zero)

### ✅ Etapa 0.1 — Núcleo do Model B em `internal/api/source/`

**Arquivos:** `internal/api/source/source.go` (novo), `resolve.go`, `kind.go`, `definition.go` (ficam intocados por enquanto — não remover ainda).

- Definir `Descriptor` (renomeando/estendendo o atual `SourceDefinition` com um campo
  `Priority int` — resolução determinística, **não** por ordem de `init()`).
- Definir a interface `Source`:
  ```go
  type Source interface {
      Describe() Descriptor
      FetchEpisodes(ctx context.Context, a *models.Anime) ([]models.Episode, error)
      FetchStreamURL(ctx context.Context, e *models.Episode, a *models.Anime, q string) (string, error)
  }
  ```
- Adicionar `Register(s Source)` (chamado de `init()`) e uma nova `Resolve` que
  escaneia os `Source`s registrados via `Describe()` — **sem apagar** a `Resolve`
  atual baseada em `sourceDefs` ainda (dual-path por uma etapa, para não quebrar
  `resolve_test.go`).
- **DECIDIDO (2026-07-06):** `providers.ForKind`/`scraper.NewScraperManager`
  continuam sendo a camada lazy/cache por baixo. `NewScraperManager` já é um
  singleton global com `sync.Once`; os `Source`s registrados em `init()` ficam
  com `sm == nil` e um accessor `manager()` cai no singleton no primeiro fetch.
  Nada é construído eagerly no boot (R6 preservado).
- Testes novos: ordenação por `Priority`, `Register` idempotente/duplicado,
  `Resolve` com múltiplos matches.

**Verificação:**
```bash
go build ./...
go test ./internal/api/source/... -v -race -count=1
```

### ✅ Etapa 0.2 — Migrar os 4 providers vivos para self-describing sources

**Arquivo:** `internal/api/providers/source_providers.go` (AllAnime, AnimeFire,
Goyabu, SuperFlix — os únicos 4 blocos ativos; FlixHQ/SFlix/NineAnime estão
comentados como `TEMP-DISABLED`, não mexer neles agora).

- Um de cada vez: trocar `RegisterProvider(kind, factory)` (Model A) por
  `source.Register(&Source{...})` (Model B), implementando `Describe()` com os
  dados que já existem em `sourceDefs` para aquele `Kind` + um `Priority`
  explícito (ordem atual de `sourceDefs` vira os valores de `Priority`).
- Manter `FetchEpisodes`/`FetchStreamURL` exatamente com a lógica atual (só troca
  a casca de registro, não o corpo).
- `providers.Provider`/`registry.go` continuam existindo e funcionando em
  paralelo até a Fase 1 trocar o chamador — não apagar nada ainda.

**Verificação:**
```bash
go build ./...
go test ./internal/api/providers/... ./internal/api/source/... -v -race -count=1
```

---

## FASE 1 — Fio único de dispatch (⚠️ toca o caminho ao vivo)

### ✅ Etapa 1.1 — `context.Context` ponta a ponta (mecânico, sem trocar lógica)

**Arquivos:** `internal/playback/movie.go`, `internal/playback/common.go`,
`internal/player/playvideo.go`, `internal/handlers/media.go`,
`internal/player/scraper.go` (`GetVideoURLForEpisodeEnhanced`).

- Adicionar `ctx context.Context` como primeiro parâmetro em toda a cadeia de
  chamada até `GetVideoURLForEpisodeEnhanced`, propagando o `ctx` real do
  chamador (ou `context.Background()` só nos pontos que hoje não têm nenhum).
- **Não mudar o dispatch ainda** — só encanar o `ctx`. Isso separa o risco
  "mecânico" do risco "lógico" da Etapa 1.2.

**Verificação:**
```bash
go build ./...
go test ./... -short -race
```

### ✅ Etapa 1.2 — `GetVideoURLForEpisodeEnhanced` vira wrapper fino

**Arquivos:** `internal/player/scraper.go`, `internal/api/enhanced.go`.

- Trocar o corpo de `GetVideoURLForEpisodeEnhanced` (e o dispatch de
  `api/enhanced.go`) para: `source.Resolve(anime)` → `Source.FetchStreamURL(ctx, …)`.
- A cadeia antiga (`isAllAnimeSourcePlayer`/`isMovieOrTVSourcePlayer`/
  `isAnimeDriveSourcePlayer`, o `switch` de `enhanced.go`, o `ScraperManager`
  switch de `unified.go`) **continua no arquivo, sem ser chamada** — vira código
  morto por uma etapa (deletado na Fase 3), servindo de rede de segurança caso
  precise reverter rápido.
- Não mexer ainda na síntese `anime == nil` (fake AllAnime) nem no fallback
  silencioso — isso é a Fase 2.

**Verificação (obrigatório rodar os testes ao vivo aqui — é a etapa de maior risco):**
```bash
go build ./...
go test ./... -short -race
go test ./internal/scraper/ -run "TestSuperFlixStreamRevival_Live" -v
go test ./internal/scraper/ -run "RealSuperFlix" -v
```

---

## FASE 2 — Endurecer as bordas

### ✅ Etapa 2.1 — `ResolveURL` substitui a síntese fake-AllAnime

**Arquivo:** `internal/player/scraper.go` (bloco `anime == nil` que hoje monta um
`models.Anime` falso de AllAnime).

- Resolução só-por-URL passa a usar `source.ResolveURL(url)` + o `Source`
  correspondente, sem construir um `Anime` sintético.

**Verificação:** `go build ./... && go test ./... -short -race`

### ✅ Etapa 2.2 — `Unknown` explícito na borda do dispatch

**Arquivos:** `internal/player/scraper.go`, `internal/api/enhanced.go`.

- Quando `source.Resolve`/`ResolveURL` não casar nada: log explícito
  ("unrecognized source"), fallback best-effort para AllAnime **só se
  configurado** — nunca silencioso.

**Verificação:** `go build ./... && go test ./... -short -race`

---

## FASE 3 — Deletar & organizar

### ✅ Etapa 3.1 — Apagar as três camadas antigas + nomes finais do §2

**Arquivos:** `internal/player/scraper.go` (helpers `isXSourcePlayer`),
`internal/api/enhanced.go` (branching por `anime.Source`),
`internal/scraper/unified.go` (`switch` de adapters no `ScraperManager`),
`internal/api/source/` (definition.go/resolve.go antigos).

- Só entra depois que 1.2 + 2.1 + 2.2 estão ✅ e estáveis por pelo menos uma
  sessão. Apagar o código morto deixado como rede de segurança na Etapa 1.2.
- **OBRIGATÓRIO (conformidade §2 Model B):** com o `Resolve` antigo morto,
  renomear para a API exata do doc — `ResolveSource` → `Resolve(a *models.Anime)
  (Source, ResolvedSource)` e `ResolveSourceURL` → `ResolveURL`. Apagar
  `sourceDefs`/`SourceDefinition` (a lógica de matching migra para métodos do
  `Descriptor`). Os testes anti-drift morrem junto (não há mais dois caminhos).
- `go build` pega qualquer referência esquecida — se compilar, está seguro remover.

**Verificação:** `go build ./... && go test ./... -short -race`

### ✅ Etapa 3.2 — Criar `scraper/netx/`, mover plumbing compartilhado

**Arquivos a mover:** `ssrf.go`, `errors.go`, `source_diagnostic.go`,
`source_health.go`, `source_circuit.go` (+ seus `_test.go`) → `internal/scraper/netx/`.

- Atualizar imports em todo o repo (`grep -rl` pelos símbolos movidos).

**Verificação:** `go build ./... && go test ./... -short -race`

### ✅ Etapa 3.3 — Extrair SuperFlix → `scraper/providers/superflix/` (feita em 1 sessão)

**Arquivos (11 fontes + 14 testes hoje flat em `internal/scraper/`):**
`superflix.go`, `superflix_browser.go`, `superflix_cf.go`, `superflix_config.go`,
`superflix_streamcache.go`, `superflix_transport.go`, `superflix_tvmaze.go` + os
`_test.go` correspondentes.

- Mover para `internal/scraper/providers/superflix/`, dropar o prefixo
  `superflix_` dos nomes de arquivo dentro do novo pacote.
- `SuperFlixAdapter` continua em `manager.go` (era `unified.go`) e importa o
  subpacote — pacote leaf, sem ciclo de import.
- Se não couber numa sessão: mover os arquivos de infraestrutura primeiro
  (`transport`, `cf`, `config`, `streamcache`, `tvmaze`) e marcar 🔄; o núcleo
  (`superflix.go` + `superflix_browser.go`, os dois maiores) fica para a próxima.

**Verificação:** `go build ./... && go test ./... -short -race` +
`go test ./internal/scraper/providers/superflix/... -run "RealSuperFlix|Live" -v`

### ✅ Etapa 3.4 — Extrair `allanime/`, `animefire/`, `goyabu/`

Mesmo padrão da 3.3, um de cada vez: `allanime.go` (+test) →
`scraper/providers/allanime/`; `animefire.go` (+test) →
`scraper/providers/animefire/`; `goyabu.go` (+test) → `scraper/providers/goyabu/`.

**Verificação (rodar após cada um dos três, não só no final):**
`go build ./... && go test ./... -short -race`

### ✅ Etapa 3.5 — Renomear `unified.go` → `manager.go`; limpeza final de raiz

- Renomear arquivo, remover prefixos de pacote agora redundantes.
- Confirmar que a limpeza de raiz do §6.4 já está feita (já está: `TEST_*.md` em
  `docs/testing/`, `CLAUDE.md` já com o case certo, `ARCHITECTURE.md` na raiz).
- Adicionar `doc.go` (uma linha) a cada novo subpacote criado nas Etapas 3.2–3.4.

**Verificação:** `go build ./... && go vet ./... && go test ./... -short -race`

---

## FASE 4 — Capacidades sob demanda (Model C)

### ✅ Etapa 4.1 — `Seasoned` + `BrowserGated` no SuperFlix

**Arquivo:** `internal/api/source/source.go` (novas interfaces opcionais),
`internal/scraper/providers/superflix/` (implementação).

```go
type Seasoned     interface { Seasons(ctx context.Context, a *models.Anime) ([]Season, error) }
type BrowserGated interface { WarmUp(ctx context.Context) error }
```

- SuperFlix é o **primeiro e único** consumidor por enquanto — não introduzir
  `Searchable` ou outras capacidades sem um segundo source real que precise.
- Dispatcher faz type-assert (`if s, ok := src.(source.Seasoned); ok { … }`) e
  loga explicitamente quando uma capacidade não existe (R5: nunca um no-op silencioso).

**Verificação:** `go build ./... && go test ./... -short -race`

---

## FASE 5 (opcional / backlog) — Robustecimento inspirado no Curd (S1–S3)

> Não bloqueia a recomendação central (Model B + hook para C). Fazer só depois
> da Fase 4, se/quando o time quiser o kill-switch manual.

### ✅ Etapa 5.1 — S1 (kill-switch) implementado; S2 + S3 avaliados e dispensados

**Notas da ETAPA 5.1 (2026-07-08) — S1 feito, S2/S3 dispensados após avaliação:**

Avaliação do estado real antes de implementar (grounded no código):
- **S2 (host-services) — DISPENSADO.** O objetivo era impedir providers de
  importar "para cima". Verificado: os providers de produção importam SÓ
  `netx`/`util`/`models` — e `util` é leaf (fundo do DAG §6.2). Importar `util`
  não é importar para cima; o problema que o S2 resolveria não existe aqui.
  Adicionar `hostservices` seria indireção morta (§7 rec. 5: "adopt the
  patterns, not the heavy machinery").
- **S3 (wiring único) — DISPENSADO.** O Curd precisa de `load.go` porque cada
  source é pacote separado blank-importado. Aqui os 4 se auto-registram no
  `init()` de UM pacote (`api/providers/source_providers.go`), puxado por UM
  único blank import (`player/scraper.go`). O "set ativo num relance" já é um
  lugar só; mover a linha seria cosmético.
- **S1 (kill-switch) — IMPLEMENTADO** (único com ganho real: complemento manual
  do circuit breaker automático).

Implementação do S1:
- `util.SourceDisabled`/`SourceForceEnabled` (parse de `GOANIME_DISABLED_SOURCES`
  / `GOANIME_ENABLED_SOURCES`, case-insensitive e dot-forgiving via
  `canonSourceToken`). Fica no `util` (leaf) para busca E dispatch honrarem o
  mesmo switch sem ciclo de import.
- `Descriptor.DefaultDisabled` + `source.IsEnabled`/`Enabled`/`DisabledSources`
  + filtro em `registeredByPriority` (source desabilitado nunca resolve; log
  Debug por skip — R5). Player best-effort usa `source.Enabled`.
- Busca (`ScraperManager.searchAllScrapersConcurrent` + `searchSpecificScraper`)
  pula sources desabilitados com Warn.
- Confirmação visível no startup (`HandlePlaybackMode`): Warn listando sources
  desligados (R5: kill-switch nunca é surpresa silenciosa).
- Testes: +2 util, +4 source, +2 scraper. Live: `GOANIME_DISABLED_SOURCES=AllAnime,SuperFlix`
  → busca real retornou 21 resultados, nenhum das fontes desligadas.

**Verificação:** `go build ./... && go test ./... -short -race` — verde · lint 0 issues.

---

## FASE 6 — Fechar o gap de dispatch (episódios + busca no registry)

> Descoberto na auditoria de 2026-07-08: o registry cobria só o STREAM-play; a
> BUSCA (ScraperManager) e os EPISÓDIOS (switch de `enhanced.go`) ainda rodavam
> nos dispatchers antigos. Objetivo: unificar os dois no registry e então
> deletar o switch de `enhanced.go` + o `ScraperManager`.

### ✅ Etapa 6.1 — Episódios pelo registry (feito, verificado)

- Novo `providers.FetchEpisodes(ctx, anime)` (`api/providers/dispatch.go`):
  `source.Resolve` (honra kill-switch S1 + best-effort) → normaliza `anime.Source`
  quando vazio → `src.FetchEpisodes`. Equivalente ao switch de `GetAnimeEpisodesEnhanced`.
- SuperFlix provider `FetchEpisodes` corrigido: era stub-adapter (errava); agora
  delega a `api.GetSuperFlixEpisodes` via `superFlixEpisodesFn` (seletor de
  temporada UX-idêntico, sem mover código de UI → evita a inversão de dependência).
- Callers VIVOS roteados pro registry: `appflow/anime_data.go` (busca→play),
  `download/workflow.go` (3 chamadas), `playback/series.go` (change-anime).
- `GetAnimeEpisodesEnhanced` ainda VIVE para 2 callers api-internos
  (`series.go:IsSeriesEnhanced`, `allanime_smart.go:DownloadAllAnimeSmartRange`) —
  deletá-lo exige a inversão `api↔providers` (Etapa 6.3).
- Testes: +4 dispatch. Live: AllAnime 500, AnimeFire 219, Goyabu 500 episódios
  pelo registry. Suíte -race verde, lint 0 issues.

### ✅ Etapa 6.2 — Busca pelo registry (`Searchable` capability) — feito, verificado

- `source.Searchable { Search(ctx, query) }` (capability Model C) + helper
  `source.ActiveSources()` (sources habilitados por prioridade, respeita S1).
- Os 4 providers implementam `Search` delegando a
  `sm.SearchAnime(query, &type)` (= `searchSpecificScraper`) — REUSA breaker +
  kill-switch + diagnostics + tagging, **sem duplicar** o motor.
- `providers.SearchAll(ctx, query, kinds...)` (`dispatch.go`): fan-out
  concorrente sobre `ActiveSources()` filtrado por `Searchable`, com straggler
  grace (5s) + timeout agregado (15s); tolera falha por-source; kindFilter para
  busca de source específico.
- Ciclo `api↔providers` resolvido com SEAM de função: `api.SetSearchFetch`
  (chamado por `providers.init()`) — `SearchAnimeEnhanced` usa o seam quando
  ligado, senão cai no ScraperManager. Sem `api` importar `providers`.
- Verificação: suíte -race verde · lint 0 issues · +6 testes (fan-out, filtro,
  tolerância a falha, all-fail, no-searchable). Live: `SearchAll("naruto")` =
  64 resultados das 4 fontes, todos com tag de idioma.
- **Gap conhecido (aceito p/ 6.3):** `SearchAll` reusa o breaker via
  `searchSpecificScraper`, que NÃO tem per-source timeout + origin-probe (só o
  `searchAllScrapersConcurrent` tinha). Um source que trava não trippa o breaker
  por timeout (mas o straggler/timeout agregado sempre limita a espera do
  caller). Volta completo na 6.3 com adapters ctx-aware. Path PT-BR
  (`SearchAnimePTBR`) e `media_manager` seguem no ScraperManager até a 6.3.

### 🔄 Etapa 6.3 — Deletar os dispatchers antigos (parcial: switch de episódios DELETADO)

**6.3a ✅ (2026-07-08) — `GetAnimeEpisodesEnhanced` DELETADO.** O switch de
episódios (100 linhas, 29 checagens `anime.Source==`) morreu. Substituído por um
seam simétrico ao da busca: `api.SetEpisodesFetch` (ligado por `providers.init()`
→ `providers.FetchEpisodes`). Os 5 callers api-internos (series, allanime_smart,
download x3) usam `fetchEpisodesViaRegistry`. 15 testes do switch deletados
(cobertura equivalente já em `providers.TestResolve_LiveRegistry` +
`TestFetchEpisodes_*`); os 2 do SuperFlix já existiam apontando pra
`GetSuperFlixEpisodes`. Helpers de teste de download religam o seam via
`mockEpiScraper`. Live: download path lista 219 episódios pelo registry.
Suíte -race verde, lint 0 issues.

**6.3b ✅ (2026-07-08) — switch de stream `GetEpisodeStreamURL` DELETADO.**
- AllAnime provider tornado auto-contido: `FetchStreamURL` chama o adapter
  direto (`GetStreamURL` → client `GetEpisodeURL`) + publica o referer do
  metadata. Não delega mais ao api. Removidas `allAnimeEnhancedStreamFn`/`fallbackStreamFn`.
- Novo `providers.FetchStreamURL(ctx,...)` (dispatch.go): resolve + WarmUp +
  delega, espelhando `FetchEpisodes`. Stream seam `api.SetStreamFetch` ligado
  por `providers.init()`.
- Callers roteados: `player/download.go` (3 sites → `providers.FetchStreamURL`,
  loop AllAnime deduplicado), api-internos (enhanced.go DownloadEpisode*,
  allanime_smart) → `fetchStreamViaRegistry`.
- DELETADOS: `GetEpisodeStreamURL` (switch, 111 linhas), `GetEpisodeStreamURLEnhanced`,
  `GetAllAnimeEpisodeURLDirect`, `extractAllAnimeIDAPI`. Mantido `isAllAnimeSourceAPI`
  (usado por allanime_smart). ~24 testes obsoletos removidos (cobertura em
  `providers` + `source.Resolve`); helpers de download religam ambos os seams.
- Live: AllAnime + Goyabu resolvem stream pelo registry. Suíte -race verde, lint 0.

**6.3c 🔄 (2026-07-08) — busca 100% no registry; `ScraperManager` mantido como motor interno.**
- Caminho PT-BR (`SearchAnimeEnhanced` isPTBR) roteado para
  `providers.SearchAll(kinds=AnimeFire,Goyabu,SuperFlix)`. Live: 36 resultados,
  zero AllAnime. Nenhum caminho de busca do fluxo principal usa mais o fan-out
  do ScraperManager (só `SearchAnimePTBR`/`SearchAnime` como fallback quando o
  seam não está ligado — production-dead).
- **DECISÃO (avaliação honesta):** o objetivo do §5 ("colapsar os 3 dispatchers
  vivos num caminho declarativo") está ATINGIDO — episódios, stream E busca
  passam pelo registry. O `ScraperManager` NÃO é mais um dispatcher vivo: virou
  motor interno (fan-out concorrente + breaker + health + tagging) que os
  providers usam para episódios/stream via `GetScraper` e a busca reusa via
  `searchSpecificScraper`. Deletá-lo fisicamente exigiria: reescrever todos os
  métodos dos providers para client-direto, RELOCAR o breaker/tagging/timeout
  (search-only, vivem no manager), migrar `MediaManager` (handlers), e apagar a
  enum `ScraperType` — multi-sessão, nos caminhos mais usados, TTY-gated
  (seleção de busca, seletor de temporada, quality picker AnimeFire), com
  retorno arquitetural decrescente. Fica como cleanup opcional, não como
  requisito de dispatch.

**6.3d 🔄 (2026-07-08) — breaker + tagging PORTADOS; busca self-contained (providers seguram adapters).**
FEITO nesta sessão (parte conceitual mais difícil):
- **Breaker portado** → `netx/circuit.go` (`netx.CircuitBreaker`, genérico por
  string, +8 testes). Não depende mais do `ScraperManager`.
- **Tagging portado** → `providers/tagging.go` (`tagResults`/`cleanPTBRTitle`
  por `SourceKind`, +5 testes; comportamento idêntico ao legado).
- **Providers self-contained:** os 4 seguram o próprio adapter via
  `scraper.NewAdapter` + `lazyGetAdapter` (não usam mais `p.manager().GetScraper`).
  `Search` = adapter.SearchAnime + `tagResults` (não mais `searchViaManager`).
  Episódios/stream de AllAnime/AnimeFire/Goyabu = adapter-direto. SuperFlix
  episódios/stream ainda delegam ao api (`GetSuperFlix*`).
- **`SearchAll` ganhou o breaker próprio** (`searchBreaker = netx.NewCircuitBreaker()`,
  por `SourceKind`): pula source com breaker aberto, registra success/failure.
- Live: busca self-contained = 49 resultados, todos tagueados, das 4 fontes,
  **zero `ScraperManager` no path de busca**. Suíte -race verde, lint 0 issues.
- Bônus: revertido bump acidental do `go.mod` (go1.27 → go1.26.5) causado por
  `go run ...@latest`.

FALTA para o delete físico do `ScraperManager` (mecânico, mas toca MediaManager):
1. ✅ **FEITO** — `api/enhanced.go` limpo: `newScraperMgr` DELETADO (morto),
   `SearchAnimeEnhanced` simplificado (exige o seam; fallbacks `SearchAnime`/
   `SearchAnimePTBR` + `scraperManager` local removidos; import `scraper` órfão
   removido). Helpers de teste `injectScraper`/`injectMultiScraper` enxutos (só
   religam os seams episódio/stream; a parte `newScraperMgr` morta saiu).
   `TestSearchAnimeEnhanced_NoResultsReturnsError` religado via seam-stub.
   Suíte -race verde, lint 0.
2. ✅ **FEITO** — factory legado DELETADO: `provider.go` + `registry.go` +
   `registry_test.go` removidos; `RegisterProvider`/`ForKind`/`ForAnime`/
   `HasProvider`/`ResetForTesting`/`Provider` interface deletados. Campo `sm` +
   param `sm` de `lazyGetAdapter` removidos (providers 100% Model B via
   `source.Register`, seguram adapter próprio). Métodos `Kind()` órfãos
   removidos (`HasSeasons` fica — usado por `source.Seasoned`). Testes
   obsoletos (`*_ScraperNotFound`, `TestForKind_*`, `emptyManager`) removidos;
   `KindAndHasSeasons` reescritos via `Describe().Kind`. Suíte -race verde, lint 0.
3. ✅ **FEITO** — `MediaManager` DELETADO. `handlers/media.go` agora usa
   `registryMediaSource` (implementa a interface `mediaSource` via
   `providers.SearchAll` + `source.Resolve`→`FetchStreamURL`). `NewMediaHandler`
   nunca era chamado em prod (fluxo morto exercido só por testes com fake);
   `scraper.NewAdapter` já constrói os adapters direto, então o `MediaManager`
   (switch por source + wrapper do ScraperManager) saiu inteiro
   (`media_manager.go` + `_test.go`). Blocos comentados stale FlixHQ/SFlix/
   NineAnime removidos de `source_providers.go`. Testes obsoletos de
   `MediaManager` em `search_stream_test.go` removidos. Suíte -race verde, lint 0.
   **Resultado: o `ScraperManager` está fora de TODOS os fluxos principais**
   (busca/episódios/stream/media). Sobram só 2 touchpoints:
   - Pre-warm (`PreWarmScraperManager`, `cmd/main.go:92`) — aquece um singleton
     que ninguém mais usa (os providers constroem adapters via `NewAdapter`).
   - Health diagnostic (`CheckAllSourcesHealth`) — usado pelo CI
     (`TestSourceHealthLive`, tag `sourcehealth`).
4. **FALTA — delete físico do struct `ScraperManager`:**
   - Repontar/deletar `PreWarmScraperManager` (aquecer os adapters do registry,
     ou virar no-op + remover a chamada no `main.go`).
   - Migrar o health diagnostic (`source_health.go`) para o registry (health por
     `providers.Search` por source) + atualizar `TestSourceHealthLive`.
   - Deletar `ScraperManager` struct + `SearchAnime`/`searchSpecificScraper`/
     `searchAllScrapersConcurrent`/`SearchAnimePTBR`/`GetScraper`/breaker
     (`source_circuit.go`)/`getScraperDisplayName`/`tagResults`/`getLanguageTag`/
     `cleanPTBRTitle`/`searchWithTimeout` + os testes do manager. MANTER os TIPOS
     de adapter (os providers os constroem via `NewAdapter`) e `ScraperType`.
   - **TTY:** verificar seleção de busca + seletor de temporada + quality picker
     antes/depois (fluxo `handlers/media.go` é morto, mas a busca real não).

1. **Adapters ctx-aware** (`UnifiedScraper.SearchAnime`/`GetAnimeEpisodes`/
   `GetStreamURL` recebem `ctx`). Sem isso o per-source timeout + origin-probe
   não migram para `SearchAll` (gap conhecido da 6.2).
2. **Inverter `providers→api`**: mover os corpos de `enhanced.go` (switch de
   `GetEpisodeStreamURL`, `GetSuperFlixEpisodes` com o seletor de temporada) e
   `SearchAnimePTBR` para `providers`/pacotes por source. Isso arrasta deps de
   UI (spinner/tui/fuzzyfinder) para `providers` — avaliar mover a seleção para
   `appflow`.
3. **Realocar** `api/series.go:IsSeriesEnhanced` e
   `api/allanime_smart.go:DownloadAllAnimeSmartRange` para fora do `api` (são
   chamados só de `playback`/`download`), liberando o delete de
   `GetAnimeEpisodesEnhanced`.
4. **Migrar** `media_manager.go` e o path PT-BR para `providers.SearchAll`.
5. Então deletar: `GetAnimeEpisodesEnhanced`, o switch de `GetEpisodeStreamURL`,
   e o `ScraperManager` (com sua `ScraperType` machinery).

- **Requer TTY** para verificar seletor de temporada (SuperFlix) + seleção da
  busca antes do delete. Também requer soak-test de concorrência do breaker.

---

## Status atual

_(atualizar após cada etapa)_

| Fase | Etapa | Status | Data |
|---|---|:---:|---|
| 0 | 0.1 Núcleo Model B | ✅ | 2026-07-06 |
| 0 | 0.2 Migrar 4 providers | ✅ | 2026-07-06 |
| 1 | 1.1 ctx ponta a ponta | ✅ | 2026-07-06 |
| 1 | 1.2 Wrapper fino | ✅ | 2026-07-06 |
| 2 | 2.1 ResolveURL | ✅ | 2026-07-06 |
| 2 | 2.2 Unknown explícito | ✅ | 2026-07-06 |
| 3 | 3.1 Apagar camadas antigas + nomes §2 | ✅ | 2026-07-07 |
| 3 | 3.2 `scraper/netx/` | ✅ | 2026-07-07 |
| 3 | 3.3 Extrair SuperFlix | ✅ | 2026-07-07 |
| 3 | 3.4 Extrair allanime/animefire/goyabu | ✅ | 2026-07-08 |
| 3 | 3.5 Rename + doc.go | ✅ | 2026-07-08 |
| 4 | 4.1 Seasoned + BrowserGated | ✅ | 2026-07-08 |
| 5 | 5.1 S1 feito · S2/S3 dispensados | ✅ | 2026-07-08 |

**Próxima etapa:** 5.1 (backlog opcional — S1 kill-switch, S2 host-services, S3 wiring único).
O plano principal (Fases 0–4) está COMPLETO: Model B no ar + hook para C entregue.

**Notas da ETAPA 4.1 (2026-07-08) — FASE 4 COMPLETA, plano principal fechado:**
- Vocabulário Model C em `internal/api/source/capabilities.go`: interfaces
  opcionais `Seasoned` (`HasSeasons() bool`) e `BrowserGated` (`WarmUp(ctx) error`),
  descobertas por type-assert. Helpers `IsSeasoned`/`IsBrowserGated`/`WarmUp`
  encapsulam a assertion com log explícito (R5: capability ausente é visível,
  nunca no-op silencioso).
- SuperFlix é o ÚNICO consumidor (regra do §4): implementa `WarmUp` (fail-fast
  se `HeadlessEnvironment()` — sem display e sem `--sf-headless`; barato, sem
  solve eager) e já satisfazia `Seasoned` via `HasSeasons()→true`.
  `BrowserGated` é type-assert limpo (só SuperFlix); `Seasoned` é transitório
  (os outros carregam `HasSeasons()→false` da Provider antiga, mas `IsSeasoned`
  dá a resposta certa).
- Consumidor real no dispatch (`player/scraper.go`): após `source.Resolve`,
  loga `seasoned`/`browserGated` e chama `source.WarmUp(ctx, src)` antes de
  `FetchStreamURL` — SuperFlix headless falha rápido com mensagem clara em vez
  de travar num browser que nunca abre; fontes não-gated são no-op logado.
  Sem duplicar o preflight interno (WarmUp aborta antes de chegar nele).
- Testes: +4 (source helpers), +3 (providers: só SuperFlix é gated/seasoned;
  WarmUp headless/cancelamento), +2 (player dispatch: WarmUp falha→aborta antes
  do fetch; sucesso→prossegue). Suíte -race verde, lint 0 issues.
- NÃO introduzidas `Searchable`/`EpisodeLister` — sem segundo source que precise
  (evita over-engineering, §4).

**Notas da ETAPA 3.5 (2026-07-08) — FASE 3 COMPLETA:**
- `unified.go` → `manager.go` (e `unified_*_test.go` → `manager_*_test.go`).
- `allanime/client.go` (1561 linhas) dividido por responsabilidade: `client.go`
  (77 — struct/constructores) + `crypto.go` (284 — AES-GCM, decode de payloads)
  + `search.go` (148) + `episodes.go` (267) + `stream.go` (832 — pipeline
  GraphQL→sources→links→quality; denso mas coeso).
- `doc.go` dos subpacotes novos (netx + 4 providers) já criados nas etapas
  anteriores; §6.4 (raiz/docs) já estava feito. C4 completo (todos os pacotes
  com doc.go) permanece item do punch-list geral, fora desta fase.
- Suíte -race verde · lint 0 issues · vet limpo.

**Notas da ETAPA 3.4 (2026-07-08):**
- Extraídos para `providers/allanime|animefire|goyabu/` (cada um com `doc.go`):
  fontes viraram `client.go`; testes foram junto (`client_test.go`,
  `ctr_regression_test.go`, `stream_test.go`, `issue166_regression_test.go`).
- `AllAnimeClient.GetType()` DELETADO — só existia para satisfazer UnifiedScraper
  num cast que nunca sucedia (o manager sempre guardou adapters); manteria um
  ciclo provider→scraper. Adapters continuam donos de GetType.
- Descoberta relacionada: o cast `scraperInstance.(*AllAnimeClient)` em
  `api/enhanced.go` (caminho AniSkip de GetAnimeEpisodesEnhanced) era CÓDIGO
  MORTO em produção — substituído pelo caminho que sempre rodou, com NOTE
  explicando; reativar AniSkip ali é decisão consciente futura via
  `adapter.Client()`.
- `UserAgent` compartilhado (definido no allanime, usado por animefire/goyabu)
  → movido para `netx.UserAgent`; SuperFlix mantém o dele (UA atado ao browser
  que resolve o CF).
- `search_stream_test.go` dividido pelo dono: white-box de clients →
  `animefire/scrape_test.go` + `goyabu/sleep_test.go`; testes de
  MediaManager/Manager/Adapters ficaram, usando os novos seams
  `<pkg>.NewClientForTest(url)` (padrão do superflix, replicado nos 3).
- Consumidores atualizados: `unified.go`, `api/enhanced.go`,
  `api/allanime_enhanced.go`, `playback/allanime_navigation.go`,
  `player/playvideo.go`, `adapters_test.go`.
- `.vscode/settings.json` criado com `explorer.compactFolders: false` (pedido
  do usuário: árvore sem compactar `providers/superflix` numa linha).
- Suíte -race verde · lint 0 issues · vet limpo.

**Notas da ETAPA 3.3 (2026-07-07):**
- 7 fontes + 19 testes + 1 fixture movidos para `internal/scraper/providers/superflix/`
  com prefixo `superflix_` dropado dos nomes (§6.3): `client.go` (era superflix.go),
  `browser.go`, `cf.go`, `config.go`, `streamcache.go`, `transport.go`, `tvmaze.go`
  + `doc.go`. Fontes eram leaf puro (zero símbolos do scraper restante) → moveram
  todos de uma vez, sem seam temporário.
- `superflix_test.go` (2214 linhas) foi DIVIDIDO pelo dono real dos símbolos:
  testes de `SuperFlixAdapter`/`tagResults`/`ScraperManager`/`sortPTBRFirst` →
  `scraper/superflix_adapter_test.go` (ficam com unified.go); testes de guards →
  `netx/response_guard_test.go`; o resto moveu como `client_test.go`.
- `integration_test.go` era black-box (`scraper_test`) → moveu inteiro como
  `superflix_test` (o teste ThroughScraperManager importa scraper sem ciclo).
- Novo seam de teste exportado: `superflix.NewClientForTest(serverURL)` (padrão
  `NewScraperManagerForTest`) — usado pelos testes do adapter que ficaram.
- `skip_in_ci_test.go` do scraper (catch-all do C8) DELETADO — helper duplicado
  para o pacote superflix; todos os usuários estavam lá.
- `SuperFlixAdapter` permanece em `unified.go` importando o subpacote, como §6.6.
- Verificação: suíte -race verde · lint 0 issues · live: Revival PASS, Sniff
  PASS, SearchAndVerify PASS, host-pin PASS · GetEpisodes FAIL pré-existente
  (air_date upstream, mesma da 1.2).
- **Pós-polish (2026-07-08, ataca C2):** os dois god-files divididos por
  responsabilidade (§6.3, mesmo pacote = zero risco): `client.go` (1248) →
  `client.go` (client/config/constructores) + `types.go` (modelos+ToAnimeModel)
  + `search.go` + `stream.go` (pipeline bootstrap→getVideo) + `episodes.go`;
  `browser.go` (1388) → `browser.go` (solver core) + `setup.go`
  (install/profile/marker) + `sniff.go` (captura de stream do embed) +
  `gate.go` (helpers de página/Turnstile). Maior arquivo do pacote agora: 548
  linhas. Suíte -race verde, lint 0 issues.

**Notas da ETAPA 3.2 (2026-07-07):**
- Movidos para `internal/scraper/netx/` (com `doc.go`): `ssrf.go`, `errors.go`,
  `source_diagnostic.go` + testes. Exportados na mudança: `CheckHTTPStatus`,
  `CheckHTMLResponse`, `CheckChallengeDocument`, `ValidateStreamURL`,
  `IsDisallowedIP`, `SafeDialFunc`, `SafeScraperTransport` (eram unexported no
  god-package). Consumidores atualizados: pacote scraper inteiro,
  `player/download.go`, `player/scraper.go`, `pkg/goanime/errors.go`.
- **Desvio do §6.6 (registrado):** `source_health.go` e `source_circuit.go`
  NÃO foram para netx — são métodos/estado do `ScraperManager` (mapas por
  `ScraperType`); movê-los criaria ciclo netx↔scraper. São estado do manager,
  não plumbing puro. Revisitar quando o manager encolher (3.5) ou o breaker
  for re-chaveado por `source.SourceKind` (candidato natural: Fase 4/5).
- Testes divididos pelo dono real: `TestScraperManager_BaseURLForKnownTypes` →
  `scraper/manager_baseurl_test.go`; testes de integração circuit/health (usam
  `MockScraper`/`createTestManager`) → `scraper/diagnostic_integration_test.go`.
  Testes puros de diagnostic/probe/ssrf/errors foram com seus arquivos.
- Suíte -race verde · golangci-lint 0 issues.

**Notas da ETAPA 3.1 (2026-07-07):**
- **API final do §2 no ar:** `Resolve(a *models.Anime) (Source, ResolvedSource)`
  e `ResolveURL(url) (Source, ResolvedSource)` — assinaturas exatas do doc.
  `definition.go` (`sourceDefs`/`SourceDefinition`) DELETADO; matching agora são
  métodos do `Descriptor` (`matchNonExplicit`/`matchURL`), com testes dedicados.
  Registry é a única fonte de resolução; testes anti-drift removidos (não há
  mais dois caminhos) e substituídos por `TestResolve_LiveRegistry`/`TestResolveURL_LiveRegistry`
  em providers (validam os descriptors REAIS registrados via init()).
- **Deletados do player** (prod-dead): `isAllAnimeSourcePlayer`,
  `isAnimeDriveSourcePlayer`, `isLikelyAllAnimeID`, `isNumericString` e os
  regexes órfãos (`isNumericRe`, `hasLetterRe`) + seus testes.
- **Escopo re-anotado (honesto):** o branching de `api/enhanced.go` e o
  `ScraperManager`/adapters de `unified.go` NÃO puderam ser apagados aqui —
  são o motor por baixo dos providers desde a 1.2. Eles morrem nas Etapas
  3.3–3.5, quando os corpos migrarem para os pacotes por source.
  `isMovieOrTVSourcePlayer` também sobrevive como política transitória de
  erro/extração no wrapper (morre quando a política normalizar).
- Corrida de dados em teste corrigida: testes de `FetchStreamURL`
  (AnimeFire/Goyabu) mutam globals de util → sem `t.Parallel()`.
- Pré-requisito de "uma sessão de uso real" foi dispensado pelo usuário ao
  pedir a etapa; a rede de segurança agora é o git history do PR #1391.

**Notas da FASE 2 (2026-07-06):**
- 2.1: bloco `anime == nil` agora usa `source.ResolveSourceURL(episode.URL)`;
  o contexto mínimo (`Source: string(resolved.Kind)`) deriva do que o registry
  casou — a síntese hardcoded fake-AllAnime foi deletada. URLs HTTP sem dono
  continuam na extração legacy (decisão: mudança mínima; a extração legacy
  morre na Fase 3). Valor não-casado → erro explícito, nunca palpite.
- 2.2: borda do dispatch com `Unknown` nunca é silenciosa — `util.Warn` alto
  ("unrecognized source; dispatching best-effort AllAnime") e a env
  `GOANIME_STRICT_SOURCE=1|true` (novo `util.StrictSourceResolution()`)
  transforma o fallback em erro. Default mantém o best-effort (compatibilidade);
  o kill-switch por config completo é o S1 da Fase 5.
- Testes: +4 no player (`source_dispatch_test.go`), +1 no util. Suíte -race
  verde, golangci-lint 0 issues.

**Notas da ETAPA 1.2 (2026-07-06) — Phase 1 do §5 COMPLETA:**
- `GetVideoURLForEpisodeEnhanced` agora despacha via `source.ResolveSource` →
  `Source.FetchStreamURL(ctx, …)`. Blank import de `api/providers` em
  `player/scraper.go` popula o registry (consolidar num wiring file único = S3, backlog).
- **Paridade por delegação:** cada provider delega às MESMAS funções api que o
  caminho antigo chamava — AllAnime: `GetEpisodeStreamURLEnhanced` (seta referer
  global!) com fallback `GetEpisodeStreamURL`; SuperFlix: `GetSuperFlixStreamURL`
  (spinner, preflight browser, referer/legendas globais, erros amigáveis);
  AnimeFire/Goyabu: adapter direto com paridade completa (quality, empty-check,
  `ClearGlobalSubtitles`/`SetGlobalAnimeSource`). Vars injetáveis
  (`superFlixStreamFn` etc.) são o seam de teste; corpos migram para pacotes
  por source na Fase 3.
- **Política transitória no wrapper** (espelha o legado byte a byte, morre na
  Fase 2/3): `isMovieOrTVSourcePlayer` decide mensagens de erro + pula extração;
  `resolved.Kind == AllAnime` → erro sem fallback; demais → fallback silencioso
  `GetVideoURLForEpisode`. `Unknown` → best-effort AllAnime via `Registered`.
- api/enhanced.go NÃO foi rewired (viraria ciclo providers↔api); ele é o motor
  por baixo dos providers até a Fase 3 deletar o branching (step 7 do §5).
- Verificação: suíte curta -race verde · `TestSuperFlixStreamRevival_Live` PASS
  · `RealSuperFlix_SearchAndVerify` PASS · `RealSuperFlix_GetEpisodes` FALHA
  **pré-existente** (air_date vazio vindo do upstream; falha igual sem o diff —
  confirmado via stash).
- 4 testes novos de wiring em `player/source_dispatch_test.go` (dispatch via
  registry, política AllAnime/movie-TV, best-effort Unknown).

**Notas da ETAPA 1.1 (2026-07-06):**
- Cadeia real difere do doc: `handlers/media.go` NÃO chama o dispatch; o handler
  real é `handlers/playback.go` (`HandlePlaybackMode`). Cadeia encanada:
  `HandlePlaybackMode` (raiz, `context.Background()` com comentário apontando o
  futuro `signal.NotifyContext`) → `HandleSeries`/`HandleMovie` → `PlayEpisode`
  → `player.GetVideoURLForEpisodeEnhanced(ctx, …)`.
- `player/playvideo.go switchEpisode` roda no event-loop de teclas do mpv (sem
  ctx hoje) → usa `context.Background()` com comentário; tornar o loop
  context-aware fica para depois.
- `GetVideoURLForEpisodeEnhanced` honra `ctx.Err()` na entrada (testado com
  ctx cancelado); propagação downstream completa acontece na 1.2.
- Bônus: `PlayEpisode` agora passa o ctx real ao `metadata.Enricher` (antes era
  `context.Background()` hardcoded).

**Notas da FASE 0 (2026-07-06):**
- Novo `internal/api/source/source.go`: `Descriptor` (com `Priority`), interface
  `Source`, `Register`/`Registered`, `ResolveSource`/`ResolveSourceURL` (que
  espelham a semântica de `Resolve`/`ResolveURL`, incluindo fallback PT-BR),
  `SwapRegistryForTesting`. Matching reutiliza `SourceDefinition` via
  `Descriptor.definition()` — zero duplicação de lógica durante o dual-path.
- `source_providers.go`: os 4 providers vivos implementam `Describe()` +
  `manager()` e se registram também via `source.Register` no mesmo `init()`.
  Prioridades seguem a ordem de `sourceDefs`: AnimeFire=10, Goyabu=20,
  SuperFlix=30, AllAnime=40.
- Guardas anti-drift em `source_providers_test.go`
  (`TestResolveSource_AgreesWithResolve`, `TestResolveSourceURL_AgreesWithResolveURL`):
  enquanto os dois caminhos de resolução existirem, qualquer divergência quebra o teste.
- `sourceDefs`/`Resolve` antigos intactos; caminho ao vivo não tocado.
