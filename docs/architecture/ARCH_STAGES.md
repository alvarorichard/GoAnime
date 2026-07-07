# GoAnime — Migração de Arquitetura por Fases (Model B → hook para C)

> **LEIA ISTO PRIMEIRO** ao continuar a migração de arquitetura.
> **Decisão de referência:** `ARCHITECTURE.md` §4 (Recommendation), §5 (Migration plan), §6 (Repo layout), §7 (prior art).
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

### ⬜ Etapa 1.1 — `context.Context` ponta a ponta (mecânico, sem trocar lógica)

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

### ⬜ Etapa 1.2 — `GetVideoURLForEpisodeEnhanced` vira wrapper fino

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

### ⬜ Etapa 2.1 — `ResolveURL` substitui a síntese fake-AllAnime

**Arquivo:** `internal/player/scraper.go` (bloco `anime == nil` que hoje monta um
`models.Anime` falso de AllAnime).

- Resolução só-por-URL passa a usar `source.ResolveURL(url)` + o `Source`
  correspondente, sem construir um `Anime` sintético.

**Verificação:** `go build ./... && go test ./... -short -race`

### ⬜ Etapa 2.2 — `Unknown` explícito na borda do dispatch

**Arquivos:** `internal/player/scraper.go`, `internal/api/enhanced.go`.

- Quando `source.Resolve`/`ResolveURL` não casar nada: log explícito
  ("unrecognized source"), fallback best-effort para AllAnime **só se
  configurado** — nunca silencioso.

**Verificação:** `go build ./... && go test ./... -short -race`

---

## FASE 3 — Deletar & organizar

### ⬜ Etapa 3.1 — Apagar as três camadas antigas

**Arquivos:** `internal/player/scraper.go` (helpers `isXSourcePlayer`),
`internal/api/enhanced.go` (branching por `anime.Source`),
`internal/scraper/unified.go` (`switch` de adapters no `ScraperManager`).

- Só entra depois que 1.2 + 2.1 + 2.2 estão ✅ e estáveis por pelo menos uma
  sessão. Apagar o código morto deixado como rede de segurança na Etapa 1.2.
- `go build` pega qualquer referência esquecida — se compilar, está seguro remover.

**Verificação:** `go build ./... && go test ./... -short -race`

### ⬜ Etapa 3.2 — Criar `scraper/netx/`, mover plumbing compartilhado

**Arquivos a mover:** `ssrf.go`, `errors.go`, `source_diagnostic.go`,
`source_health.go`, `source_circuit.go` (+ seus `_test.go`) → `internal/scraper/netx/`.

- Atualizar imports em todo o repo (`grep -rl` pelos símbolos movidos).

**Verificação:** `go build ./... && go test ./... -short -race`

### ⬜ Etapa 3.3 — Extrair SuperFlix → `scraper/providers/superflix/` (maior etapa — provável 🔄 em 2 sessões)

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

### ⬜ Etapa 3.4 — Extrair `allanime/`, `animefire/`, `goyabu/`

Mesmo padrão da 3.3, um de cada vez: `allanime.go` (+test) →
`scraper/providers/allanime/`; `animefire.go` (+test) →
`scraper/providers/animefire/`; `goyabu.go` (+test) → `scraper/providers/goyabu/`.

**Verificação (rodar após cada um dos três, não só no final):**
`go build ./... && go test ./... -short -race`

### ⬜ Etapa 3.5 — Renomear `unified.go` → `manager.go`; limpeza final de raiz

- Renomear arquivo, remover prefixos de pacote agora redundantes.
- Confirmar que a limpeza de raiz do §6.4 já está feita (já está: `TEST_*.md` em
  `docs/testing/`, `CLAUDE.md` já com o case certo, `ARCHITECTURE.md` na raiz).
- Adicionar `doc.go` (uma linha) a cada novo subpacote criado nas Etapas 3.2–3.4.

**Verificação:** `go build ./... && go vet ./... && go test ./... -short -race`

---

## FASE 4 — Capacidades sob demanda (Model C)

### ⬜ Etapa 4.1 — `Seasoned` + `BrowserGated` no SuperFlix

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

### ⬜ Etapa 5.1 — `DisabledProviders` (S1) + seam de host-services (S2) + wiring único (S3)

- S1: lista de sources desabilitadas via config, mais um `DefaultDisabled` no
  `Descriptor` — desligar um source quebrado sem rebuild.
- S2: pacote `hostservices` (ou nome equivalente) com hooks tipados
  (`HTTPClient()`, `Log()`, `StoragePath()`) que os pacotes de source consomem
  em vez de importar camadas acima deles.
- S3: consolidar todos os `import _ ".../providers/..."` num único arquivo de
  wiring (ex.: `internal/api/providers/register.go`).

**Verificação:** `go build ./... && go test ./... -short -race`

---

## Status atual

_(atualizar após cada etapa)_

| Fase | Etapa | Status | Data |
|---|---|:---:|---|
| 0 | 0.1 Núcleo Model B | ✅ | 2026-07-06 |
| 0 | 0.2 Migrar 4 providers | ✅ | 2026-07-06 |
| 1 | 1.1 ctx ponta a ponta | ⬜ | — |
| 1 | 1.2 Wrapper fino | ⬜ | — |
| 2 | 2.1 ResolveURL | ⬜ | — |
| 2 | 2.2 Unknown explícito | ⬜ | — |
| 3 | 3.1 Apagar camadas antigas | ⬜ | — |
| 3 | 3.2 `scraper/netx/` | ⬜ | — |
| 3 | 3.3 Extrair SuperFlix | ⬜ | — |
| 3 | 3.4 Extrair allanime/animefire/goyabu | ⬜ | — |
| 3 | 3.5 Rename + doc.go | ⬜ | — |
| 4 | 4.1 Seasoned + BrowserGated | ⬜ | — |
| 5 | 5.1 S1+S2+S3 (opcional) | ⬜ | — |

**Próxima etapa:** 1.1 — `context.Context` ponta a ponta (mecânico, sem trocar lógica).

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
