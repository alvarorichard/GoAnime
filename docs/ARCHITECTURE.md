# GoAnime — Source Architecture (Decision Document)

> Status: **proposal for review**. This document states the requirements,
> diagnoses today's code, and proposes **three candidate architecture models**
> (A, B, C) with trade-offs and a recommendation, plus a phased migration plan.
> **No behavior is changed by this document** — it exists to be evaluated before
> any code moves.

---

## 0. Requirements

GoAnime streams from **rotating sources** (AllAnime, AnimeFire, Goyabu,
SuperFlix, …). Hosts and providers come and go. The architecture must be:

| # | Requirement | What it means concretely |
|---|-------------|--------------------------|
| **R1** | **Intuitive** | A reader finds "how does source X work" in one obvious place. No archaeology across 50 files. |
| **R2** | **Easy to add / remove a source** | Adding a source = one declarative unit + one implementation. Removing = delete it. No edits spread across the dispatch path, no `switch`/`if-else` ladders. |
| **R3** | **Longevity & easy maintenance** | One source of truth per source — no drift between a "table" and an "implementation". A maintainer changes one thing in one place; a host rotation is a one-line edit pinned by a dated test. Context-aware (`context.Context`) so timeouts / Ctrl-C cancel cleanly with no goroutine leaks. |
| **R4** | **Scalability** | Heterogeneous sources (anime vs movie/TV vs browser-gated) coexist without forcing every source to carry methods it doesn't need. New capabilities (subtitles, seasons, search) added without touching existing sources. |
| **R5** | **Robust** | One source failing never crashes or corrupts another. Failures are explicit (typed errors + diagnostics), not silent. A repeatedly-failing source is skipped (circuit breaker) instead of retried into the ground. An unrecognized input is surfaced, not guessed. |
| **R6** | **Efficient** | A source's machinery (HTTP client, headed browser, cookie jar) is built once and reused, lazily. Resolution runs once per anime. Cacheable work (the SuperFlix `(host, hash)` pair, cleared CF profile) is cached so the steady state does no redundant work and the hot path stays browser-free. |

Three cross-cutting rules fall out of these:

- **Declarative dispatch.** Resolution scans data, not `switch` cases.
- **Explicit `Unknown`.** An unrecognized source is logged and visible, never
  silently parsed as AllAnime.
- **Isolated failure.** Each source resolves and fails on its own; a per-source
  circuit breaker + diagnostics keep one bad host from poisoning the run.

These requirements are not arbitrary — they are precisely what respected Go
projects optimize for when they expose pluggable backends: `database/sql`
(drivers), `image` (decoders), Caddy (modules), Terraform (providers),
Prometheus (collectors). §7 maps each pattern to this design.

### How the model serves R5 (robust) and R6 (efficient)

These are **mostly mechanisms that already exist** — the architecture's job is to
keep them in one obvious place per source instead of scattered across the three
live dispatchers.

- **R5 Robust** — `internal/scraper/source_circuit.go` (per-source circuit
  breaker), `source_diagnostic.go` (typed `SourceDiagnostic` with a `Kind`),
  `source_health.go` (health probing), and the explicit `Unknown` from §0. Under
  the new model each `Source` owns its own failure surface; the dispatcher logs
  and moves on rather than letting an "alien" URL crash a parser.
- **R6 Efficient** — `providers` registry already builds each provider **once**,
  lazily, and caches it (`ForKind` double-checked locking). SuperFlix already
  caches the `(playerHost, videoHash)` pair on disk
  (`superflix_streamcache.go`) and reuses a persistent CF browser profile, so
  repeat plays are pure-HTTP and browser-free. The new model preserves this:
  resolution is one pass, the `Source` instance is reused, no dispatch layer
  re-does work another already did.

---

## 1. Current state: one good architecture, asleep

There are **two** parallel architectures in the tree today.

### 1a. The clean one — already built, **not wired to anything**

`internal/api/source/` and `internal/api/providers/` already implement a
declarative model:

- **`source.SourceDefinition`** (`internal/api/source/definition.go`):

  ```go
  type SourceDefinition struct {
      Kind        SourceKind
      Explicit    []string           // values in anime.Source
      Tags        []string           // "[animefire]" in anime.Name
      URLMatchers []string           // substrings in anime.URL
      MediaTypes  []models.MediaType // MediaType → source
      ShortID     bool               // accepts AllAnime-style short IDs
  }
  var sourceDefs = []SourceDefinition{ /* one entry per source */ }
  ```

- **`source.Resolve(anime)` / `source.ResolveURL(url)`** (`resolve.go`) — iterate
  `sourceDefs`, first match wins, fall through to `Unknown` **with a warning
  log**. No `switch`.

- **`providers.Provider`** (`provider.go`) — dispatch interface, **already
  context-first**:

  ```go
  type Provider interface {
      Kind() source.SourceKind
      FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error)
      FetchStreamURL(ctx context.Context, ep *models.Episode, anime *models.Anime, quality string) (string, error)
      HasSeasons() bool
  }
  ```

- **`providers` registry** (`registry.go`) — self-registering, lazy, cached:
  `RegisterProvider(kind, factory)` in `init()`, `ForKind(kind)` / `ForAnime(anime)`
  to dispatch.

**This subsystem satisfies R1–R4 already. Its problem: nothing calls it.** It is
dead code today.

### 1b. The live one — three stacked `if/else` dispatchers

The path that runs when a user plays an episode routes by source **three times**:

| Layer | File | How it dispatches |
|-------|------|-------------------|
| 1 | `internal/player/scraper.go` → `GetVideoURLForEpisodeEnhanced(episode, anime)` | `isAllAnimeSourcePlayer` / `isMovieOrTVSourcePlayer` / `isAnimeDriveSourcePlayer` chain; **no `context.Context`**; an `anime == nil` block that **synthesizes a fake AllAnime anime**; a **silent** `GetVideoURLForEpisode(url)` fallback |
| 2 | `internal/api/enhanced.go` → `GetEpisodeStreamURL` / `…Enhanced` | dispatches by `anime.Source` |
| 3 | `internal/scraper/unified.go` → `ScraperManager` | `switch`-based adapters (`SuperFlixAdapter`, `AllAnimeAdapter`, …) |

Live callers (`internal/playback/movie.go`, `internal/playback/common.go`,
`internal/player/playvideo.go`, `internal/handlers/media.go`) all enter layer 1.

**This duplication is the actual mess.** Adding a source today means touching the
`isXSourcePlayer` chain, the `enhanced.go` dispatch, **and** the `ScraperManager`
switch — three axes, the opposite of R2.

---

## 2. Candidate models

All three share the same dispatch shape and satisfy R1–R4. They differ in **where
a source's definition lives** and **how heterogeneous capabilities are handled**.

```
caller (playback / handlers)
        │  ctx context.Context
        ▼
      Resolve(anime) ─► Source / Kind     (declarative)
        ▼
      FetchStreamURL(ctx, ep, anime, q)
        ▼
   scraper/providers/<source>             (isolated scraping)

URL only (anime == nil) ─► ResolveURL(url) ─► same path
no match               ─► Unknown + explicit Warn; best-effort only if configured
```

### Model A — Central table + separate provider (what exists today)

Definition in a central table (`sourceDefs`), behavior in a separate `Provider`,
linked by `Kind`.

```go
// definition.go — central table
var sourceDefs = []SourceDefinition{ {Kind: SuperFlix, Tags: []string{"[superflix]"}, /* ... */} }

// provider registered separately
providers.RegisterProvider(SuperFlix, func(sm) Provider { return &superFlixProvider{ /* ... */ } })
```

- **Add a source:** 3 places — a row in `sourceDefs`, a `Kind` constant, a Provider file.
- **Pros:** already built; central table is easy to read top-to-bottom.
- **Cons:** **two sources of truth** (table ≠ provider) that drift over time
  (weak on R3); a forgotten row leaves an orphan provider.

### Model B — Self-describing source (RECOMMENDED)

Collapse definition **and** behavior into **one module per source**. The source
describes itself; the registry collects it. One source = one file = everything.

```go
// internal/api/source/source.go
type Descriptor struct {
    Kind        SourceKind
    Priority    int                // explicit, deterministic resolution order
    Explicit, Tags, URLMatchers []string
    MediaTypes  []models.MediaType
    ShortID     bool
}

type Source interface {
    Describe() Descriptor                                                   // Resolve reads from here
    FetchEpisodes(ctx context.Context, a *models.Anime) ([]models.Episode, error)
    FetchStreamURL(ctx context.Context, e *models.Episode, a *models.Anime, q string) (string, error)
}

func Register(s Source)                                                     // called from init()
func Resolve(a *models.Anime) (Source, ResolvedSource)                      // scans registry[i].Describe()
```

```go
// internal/scraper/providers/superflix/source.go — everything in one place
func init() { source.Register(&Source{client: NewClient()}) }

func (s *Source) Describe() source.Descriptor {
    return source.Descriptor{
        Kind: source.SuperFlix, Priority: 30,
        Tags: []string{"[superflix]"}, URLMatchers: []string{"superflix"},
    }
}
func (s *Source) FetchStreamURL(ctx context.Context, e *models.Episode, a *models.Anime, q string) (string, error) { /* ... */ }
```

- **Add a source:** **one file** (`Describe` + `Fetch*`) + one `Kind` constant. Zero edits to a table or the dispatch path.
- **Remove a source:** delete the file. Done.
- **Pros:** **single source of truth** (definition lives with behavior — cannot drift); maximum intuitiveness (R1: "show me SuperFlix" → one file); self-registering.
- **Cons:** resolution order must be deterministic — solved by an explicit
  `Priority` field in `Descriptor`, **not** by `init()` order (which Go does not
  guarantee across files).

### Model C — Model B + optional capabilities (scales to heterogeneous sources)

Model B at the core, but a **minimal** interface plus **optional** interfaces a
source implements only when relevant. The dispatcher type-asserts.

```go
type Source interface {                                       // minimal — every source has these
    Describe() Descriptor
    FetchStreamURL(ctx context.Context, e *models.Episode, a *models.Anime, q string) (string, error)
}

// optional capabilities — implement only what the source actually does
type EpisodeLister interface { FetchEpisodes(ctx context.Context, a *models.Anime) ([]models.Episode, error) }
type Seasoned      interface { Seasons(ctx context.Context, a *models.Anime) ([]Season, error) }     // SuperFlix movie/TV
type Searchable    interface { Search(ctx context.Context, query string) ([]models.Anime, error) }
type BrowserGated  interface { WarmUp(ctx context.Context) error }                                   // SuperFlix Turnstile
```

```go
if seasoned, ok := src.(source.Seasoned); ok {               // simple anime never carries season methods
    seasons, _ := seasoned.Seasons(ctx, anime)
}
```

- **Pros:** a simple anime source doesn't carry movie/season/browser methods;
  adding a new capability (e.g. subtitles) is a **new interface with zero impact
  on existing sources** — best on R4 (scalability) and R3 (longevity) for diverse
  sources. Generalizes the `HasSeasons()` that already exists on `Provider`.
- **Cons:** more concepts; a type-assert can hide "source doesn't support X"
  unless logged clearly. Risk of over-engineering if all sources end up similar.

---

## 3. Comparison

| Criterion | A (table + registry) | **B (self-describing)** | C (B + capabilities) |
|---|---|---|---|
| R1 Intuitive | medium (2 places) | **high (1 file)** | high |
| R2 Add / remove | medium (3 points) | **high (1 file)** | high |
| R3 Longevity / maintenance (anti-drift) | weak (table ≠ provider) | **strong** | strong |
| R4 Scalability (diverse sources) | medium | good | **excellent** |
| R5 Robust (isolated failure) | same¹ | same¹ | **best** (capability misses are explicit) |
| R6 Efficient (build-once, cache) | same² | same² | same² |
| Migration effort | 0 (exists) | low (refactor existing) | medium |
| Over-engineering risk | low | low | medium |

¹ R5 mechanisms (circuit breaker, diagnostics, explicit `Unknown`) are shared by
all three; **C** edges ahead because an unsupported capability is a typed
"source can't do X" instead of a silent no-op. ² R6 (lazy build-once + cache) is
identical across models — it lives in the registry and each source's scraper, not
in the dispatch shape.

---

## 4. Recommendation

**Adopt Model B now, with a hook toward C.**

- **B** meets all four requirements with the least complexity and the best
  longevity — one source of truth per source. It is a refactor of the code that
  already exists (`source` + `providers`), not a rewrite.
- **C's capabilities are added incrementally, only when a real source needs one.**
  SuperFlix already wants `Seasoned` + `BrowserGated`; introduce those interfaces
  when wiring SuperFlix, not before. This avoids paying C's complexity up front
  while keeping the door open.
- **A is the fallback** if minimizing change is the priority — it is already
  built and wiring it alone (without the B refactor) still collapses the three
  live dispatchers into one. The cost is the long-term table-vs-provider drift.

---

## 5. Migration plan (incremental, build-green between each, no commit)

> Each phase keeps the suite green and is independently revertible. Phases that
> touch the live playback path are validated with the live tests
> (`TestSuperFlixStreamRevival_Live`, the `…RealSuperFlix…` search tests) plus
> the short suite before the next phase starts. No flag day: the new path can
> shadow the old until the final deletion.

**Phase 0 — Foundation (does not touch the live path)**
1. Define `source.Source` + `Descriptor` (with `Priority`) + `Register` / `Resolve`
   (Model B). Have `Resolve` scan registered sources' `Describe()` instead of a
   static `sourceDefs`.
2. Migrate existing providers (`source_providers.go`) to self-describing sources,
   one at a time. `source` / `providers` tests stay green.

**Phase 1 — Wire as the single path**
3. `GetVideoURLForEpisodeEnhanced` becomes a thin wrapper over
   `source.Resolve(anime)` → `FetchStreamURL(ctx, …)`.
4. Thread `ctx` end to end (callers: `playback/movie.go`, `playback/common.go`,
   `player/playvideo.go`, `handlers/media.go`).

**Phase 2 — Harden the edges**
5. URL-only resolution moves behind `source.ResolveURL` + a source — delete the
   inline fake-AllAnime synthesis (R3).
6. `Unknown` becomes explicit at the dispatch boundary — log "unrecognized
   source", best-effort AllAnime only when configured (R4 / no silent fallback).

**Phase 3 — Delete & organize**
7. Remove the old layers once unreferenced: `isXSourcePlayer` helpers, the
   `enhanced.go` per-source branching, the `ScraperManager` switch.
8. Reorganize files (§6): SuperFlix subpackage first, then the others and `netx`.

**Phase 4 — Capabilities on demand (Model C)**
9. Introduce `Seasoned` / `BrowserGated` / `Searchable` only when a source needs
   it (SuperFlix is the first).

---

## 6. Repository organization (applies to B or C)

The goal: a layout where a newcomer guesses correctly where anything lives, and a
maintainer changes one thing in one place. The structure mirrors the data flow —
**entrypoint → dispatch → sources → shared plumbing** — top to bottom.

### 6.1 Top-level layout

```
goanime/
├── cmd/goanime/            # main(): flag parsing + wiring only, no logic
├── internal/               # all private application code (see 6.2)
├── pkg/goanime/            # PUBLIC library API: Client, types/, examples/
├── docs/                   # all long-form docs (see 6.4)
├── build/                  # build scripts (linux/macos/windows, installer)
├── aur/                    # Arch packaging
├── test/                   # cross-cutting integration tests (util/ today)
├── CHANGELOG.md
├── CLAUDE.md               # (rename from claude.md — convention)
├── README.md  README_pt-br.md
├── go.mod  go.sum
└── LICENSE
```

### 6.2 `internal/` — layered, dependency flows downward

```
internal/
├── cli/ or appflow/        # orchestration: search → select → play loop
├── handlers/               # user-action handlers (media, input)
├── playback/               # playback orchestration (movie/series/common)
├── player/                 # mpv control, download, video extraction
│
├── api/                    # high-level media operations
│   ├── source/             # ── DISPATCH CORE ──
│   │   ├── kind.go         #    SourceKind constants
│   │   ├── source.go       #    Source interface + Descriptor + Register/Resolve  (Model B)
│   │   └── resolve.go      #    Resolve(anime) / ResolveURL(url) / Unknown
│   ├── providers/          #    registry + per-source Provider wiring
│   │   ├── metadata/       #    AniList/TMDB enrichment
│   │   └── naming/         #    title normalization
│   ├── aniskip/  movie/    #    feature sub-areas
│   └── …
│
├── scraper/                # ── SCRAPING LAYER ──
│   ├── manager.go          #    ScraperManager + adapters (was unified.go)
│   ├── netx/               #    shared plumbing: ssrf, errors, http helpers,
│   │                       #    source_diagnostic / source_health / source_circuit
│   └── providers/          #    one self-contained directory PER SOURCE
│       ├── superflix/      #    client, browser, cf, transport, cache, tvmaze (+ tests)
│       ├── allanime/
│       ├── animefire/
│       └── goyabu/
│
├── models/                 # domain types (Anime, Episode, …) — no deps on layers above
├── tracking/  discord/     # side features (history, Rich Presence)
├── upscaler/  updater/     # side features
└── util/                   # leaf helpers (logging, debug) — depended on by all, depends on none
```

**Dependency direction (a clean DAG, no cycles):**

```
util / models            (leaf — no app deps)
   ▲
scraper/netx             (shared plumbing)
   ▲
scraper/providers/*      (per-source scraping)   ◄─ imported by ─┐
   ▲                                                            │
scraper/manager + api/source + api/providers   (dispatch) ──────┘
   ▲
playback / player / handlers / appflow          (orchestration)
   ▲
cmd/goanime                                      (entrypoint)
```

### 6.3 File & package conventions

- **One concern per file**, named after it: `resolve.go`, `circuit.go`,
  `cache.go` — not `superflix_streamcache.go` inside a god-package. Once a source
  is its own package, the `superflix_` prefix is redundant and drops away.
- **`doc.go` per package** — a short package comment stating its one job. (R1)
- **Tests live beside code**, same package for white-box, `package x_test` for
  black-box public-API tests. No `zz`-prefix ordering hacks; no `ci_skip_test.go`
  catch-alls — name a test file after what it tests.
- **Live / browser tests are env-gated** (`testing.Short()` skip, or
  `GOANIME_RECON=1`) so the default suite is deterministic and fast. (R6)
- **Entrypoint is thin**: `cmd/goanime/main.go` wires dependencies and calls into
  `internal/` — no business logic.

### 6.4 Root & docs cleanup (R1: intuitive)

The repo root currently carries loose planning docs and a mis-cased file:

```
TEST_PLAN.md  TEST_PLAN_FUNCTIONS.md  TEST_STAGES.md  TEST_STRATEGY.md   → docs/testing/
claude.md                                                                → CLAUDE.md
```

Move the four `TEST_*.md` into `docs/testing/` and rename `claude.md` →
`CLAUDE.md`. The root then shows only what a visitor expects: README,
CHANGELOG, LICENSE, build manifests.

### 6.5 How the structure serves each requirement

| Req | Reflected in the layout |
|---|---|
| **R1 Intuitive** | Tree mirrors data flow; one directory per source; `doc.go` per package; clean root. |
| **R2 Add / remove** | A new source is a new directory under `scraper/providers/` + a self-register — nothing else moves. |
| **R3 Longevity / maintenance** | Definition lives with behavior; shared plumbing in `netx` has one home; host pinned by dated test. |
| **R4 Scalability** | New sources are sibling directories; new capabilities are new interfaces in `api/source` — existing dirs untouched. |
| **R5 Robust** | `netx` centralizes circuit breaker + diagnostics + SSRF guard; each provider package fails in isolation. |
| **R6 Efficient** | Per-source state (browser, client, caches) is encapsulated and built once in its package; deterministic, gated tests. |

### 6.6 Migration order for the file move (Phase 3 of §5)

Lowest risk first, build-green between each:

1. Create `scraper/netx/`, move shared plumbing (`ssrf`, `errors`,
   `source_diagnostic/health/circuit`), update imports.
2. Extract **SuperFlix** (worst offender: 11 source + 14 test files flat) into
   `scraper/providers/superflix/`. Leaf provider → no import cycle; the
   `SuperFlixAdapter` stays in `manager.go` and imports the subpackage.
3. Extract `allanime/`, `animefire/`, `goyabu/` the same way.
4. Rename `unified.go` → `manager.go`; drop redundant `superflix_`/provider name
   prefixes inside the new packages.
5. Root/docs cleanup (§6.4).

---

## 7. Prior art — patterns from respected projects

Model B (self-describing + self-registering sources) is **not novel** — it is the
pattern the most respected Go projects use for pluggable backends. The design
rests on proven ground:

- **Go stdlib `database/sql`** — `sql.Register(name, driver)`; a driver
  self-registers in its `init()`, pulled in by a blank import
  (`import _ ".../mysql"`). Adding a DB backend touches **zero** `database/sql`
  code. → This is exactly `source.Register(...)` in §2 Model B. Validates
  self-registration.

- **Go stdlib `image`** — `image.RegisterFormat(name, magic, decode, …)`; each
  decoder declares the magic bytes that identify its format and registers itself.
  → Mirrors `Descriptor{URLMatchers, Tags, …}`: a source declares what identifies
  it and `Resolve` matches generically. Validates self-description + declarative
  matching.

- **Caddy (web server)** — every module implements `CaddyModule() ModuleInfo` and
  calls `caddy.RegisterModule(...)`. One self-contained unit that describes and
  registers itself; the core never switches on module type. → The clearest
  real-world Model B, in a codebase widely held up as exemplary Go.

- **Terraform / HashiCorp providers** — a provider is a separate, self-contained
  unit implementing a common interface, discovered via a registry. An ecosystem
  of hundreds of **rotating** providers is the canonical proof the pattern scales
  (R4) and stays maintainable (R3). GoAnime's rotating sources are the same
  problem at small scale.

- **Prometheus** — collectors self-register (`MustRegister`); the registry
  decouples "what produces data" from "what serves it" — the same source-vs-
  dispatch split.

- **`golang-standards/project-layout`** (`cmd/` `internal/` `pkg/`) — the
  community convention GoAnime already follows. Honest caveat: it is **not** an
  official Go spec, but `cmd/`, `internal/` (compiler-enforced) and a thin `pkg/`
  public API are endorsed by the Go team and used across every project above.
  Keep it; don't over-nest.

### Recommendations distilled from the above

1. **Self-register via `init()` + blank import**, like `database/sql`. The wiring
   file does `import _ ".../providers/superflix"` and the source appears — no
   central list to edit (beats Model A's table on R2/R3).
2. **Declare identity in data, match generically**, like `image.RegisterFormat`'s
   magic bytes — `Descriptor.URLMatchers/Tags`, never `if`/`switch`.
3. **Keep the core ignorant of concrete sources**, like Caddy's core never
   knowing module names — the dispatcher knows only the `Source` interface.
4. **Thin public `pkg/` facade**, like the cited libraries expose small stable
   APIs — keep `pkg/goanime` minimal, everything else `internal/`.
5. **Don't cargo-cult scale.** GoAnime has ~4 sources, not Terraform's hundreds.
   Adopt the *patterns*, not the heavy machinery — no gRPC plugins, no codegen.
   Excellence without complexity.

> **Nuance:** blank-import self-registration (rec. 1) trades an explicit central
> list for import side-effects. That is idiomatic Go (stdlib does it), but the set
> of active sources becomes "whatever the wiring file imports". Keep **all** source
> imports in ONE obvious file (e.g. `internal/api/providers/register.go`) so the
> active set stays readable in one glance — the best of both worlds.

### Prior art, shipped: Curd v2.0.0 (sibling project, identical problem)

Curd (`Wraient/curd`) is a CLI anime player in the same niche as GoAnime. In its
**v2.0.0** release (Jun 2026) it shipped a **compile-time provider registry** that
is, point for point, **Model B plus the optional-capability layer of Model C**.
This is the strongest validation in this document: not a stdlib analogy but a
direct peer that hit the same "rotating sources" wall and resolved it exactly the
way §2–§4 recommend — in production, today.

**Mapping — this proposal ⇄ what Curd actually shipped:**

| This document | Curd v2.0.0 equivalent |
|---|---|
| Model B: self-describing source, `Register(s)` in `init()` | `providers.Register(Meta, factory)` in each `register.go` |
| `Descriptor{Kind, Tags, URLMatchers, …}` (identity as data) | `Meta{Name, Aliases, Referrer, DefaultDisabled, …}` |
| One source = one self-contained unit (R1/R2) | one package `internal/providers/<name>/` per source |
| Model C: minimal interface + optional capabilities via type-assert (`Seasoned`, `Searchable`, `BrowserGated`) | optional `ModeResolver`, `HintResolver`, `IDResolver`, discovered by type assertion |
| Rec. 1 nuance: all self-registers in ONE wiring file | `internal/loadproviders/load.go` (one blank-import list) |
| §6.2 DAG: providers never import upward | hard rule "providers must not import `internal`"; they use a `curdhost` hooks package |
| §7 rec. 5: no gRPC / `.so` plugins, single binary | stated non-goal; compile-time modules only |

**What to adopt from Curd — additions on top of B/C:**

| # | Steal | Why it helps | Req |
|---|---|---|---|
| **S1** | **Runtime disable via config** — a `DisabledProviders=[…]` list plus a per-source `DefaultDisabled` flag in the `Descriptor`, turning a source off **without a rebuild or release**. | A host rotation or a suddenly-broken source becomes a one-line config edit, not a wait-for-release. It is the manual, user-facing complement to the automatic circuit breaker: the breaker reacts to failures at runtime; the config kill-switch lets a human pre-empt a known-bad source. | R3 R5 |
| **S2** | **An explicit host-services layer** (Curd's `curdhost`) — a small package of typed hooks (`HTTPClient()`, `Log()`, `StoragePath()`, …) that source packages consume **instead of importing the layers above them**. | §6.2 draws the no-cycle DAG; this is how to *enforce* it concretely. The shared HTTP client / cookie jar / cache path (R6, build-once) reach a leaf source package through this seam with no import cycle, and the seam is trivially mockable in tests. | R3 R5 R6 |
| **S3** | **One wiring file, confirmed** — Curd keeps every source's blank import in a single `load.go`. | Validates the §7 rec. 1 "nuance" with a shipping peer: the active-source set stays readable in one glance. Adopt it as settled, not as an open question. | R1 R2 |

**What NOT to copy from Curd — GoAnime is already ahead here:**

- **Context.** Curd's `Provider` interface has **no `context.Context`**
  (`SearchAnime(query, mode)`, `GetEpisodeURL(cfg, id, ep)`). The context-first
  interface in §1a/§2 is strictly better for clean timeout / Ctrl-C cancellation
  with no goroutine leaks. **Keep `ctx` on every method; do not regress to Curd's
  signature.** (R3)
- **Failure-handling depth.** Curd's whole answer to a broken source is the manual
  config kill-switch (S1). GoAnime already has typed `SourceDiagnostic`,
  `source_circuit.go`, `source_health.go`, and an explicit `Unknown` (§0). Treat
  S1 as a **complement** to those, not a replacement — adopt the config switch
  *and* keep the structured, automatic isolation. (R5)

**Net:** Curd proves the chosen direction ships and scales for this exact domain.
Fold in S1–S3, keep GoAnime's context-first interface and structured failure
isolation, and the result is strictly a **superset** of both designs.

---

## 8. Risk notes

- Only Phase 1 (steps 3–4) is behavior-sensitive on the playback path; everything
  after is dead-code deletion and file moves.
- White-box tests reaching package-private symbols move **with** their package
  during §6, or convert to black-box (`package x_test`) where they assert public
  behavior.
- Host rotation stays a one-line change pinned by a dated test
  (`internal/scraper/superflix.go: SuperFlixBase`,
  `TestSuperFlixBase_PointsToLiveHost`) — independent of which model is chosen.
- Source resolution order is data, not `init()` order: keep it deterministic via
  `Descriptor.Priority` (Model B/C) or `sourceDefs` order (Model A).

---

## 9. Current-state assessment (honest review)

A frank read of the repository as a reviewer / "big dev" would see it. Numbers are
measured, not estimated (Go 1.26.3, 33 packages, ~32.7k src LOC, ~40.8k test LOC).

### What already stands up to scrutiny (keep, be proud of)

| Area | Evidence |
|---|---|
| **Test investment** | Test LOC (**40.8k**) exceeds source LOC (**32.7k**) — rare and respectable. Live tests are env-gated so the default suite stays deterministic. |
| **CI/CD** | `.github/workflows/` has `ci.yml` (test + build-binaries + windows-installer), `coverage.yml`, `release.yml`. Multi-platform. |
| **Static analysis** | `.golangci.yml` + `gosec` in use; security findings are triaged (real path-traversal fixed, false positives documented with `#nosec` + rationale). |
| **Layout baseline** | Already follows `cmd/` + `internal/` (compiler-enforced) + a public `pkg/goanime` with `examples/` and a `doc.go`. |
| **Hygiene** | **Zero** `TODO`/`FIXME`/`HACK` markers in Go source. |
| **Modern toolchain** | Go 1.26.3. |

### What a critic would flag (fix to silence doubt)

| # | Issue | Evidence | Requirement hurt |
|---|---|---|---|
| C1 | **Dead parallel architecture** | The clean `source`+`providers` registry has **zero** live callers; three `if/else` dispatchers run instead (§1). | R1 R2 R3 |
| C2 | **God files** | 9 files at 1.2k–2.3k LOC: `player/download.go` (2275), `player/player.go` (1796), `scraper/allanime.go` (1555), `player/playvideo.go` (1461), `player/scraper.go` (1416), `downloader/downloader.go` (1344), `util/util.go` (1217), `scraper/superflix_browser.go` (1214), `scraper/superflix.go` (1192). | R1 |
| C3 | **God package** | `internal/scraper` holds 4 providers + shared plumbing + manager flat (50+ files). | R1 R4 |
| C4 | **Package docs** | Only **1 of 33** packages has a `doc.go`. | R1 |
| C5 | **gofmt drift** | ~10 source files are not `gofmt`-clean → formatting isn't enforced in practice. | R3 |
| C6 | **Dependency weight** | **73** direct dependencies for a CLI — supply-chain surface and build cost. | R3 R6 |
| C7 | **Missing OSS health files** | No `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`, or `ARCHITECTURE.md` (this doc fills the last). | R1 |
| C8 | **Naming hacks / root clutter** | `zzchooser_test.go` (zz-sort hack), `ci_skip_test.go` (catch-all), `claude.md` (mis-cased), four `TEST_*.md` loose at root. | R1 |

### Verdict

The **engineering discipline is already strong** (tests, CI, security triage) —
this is not a weak project. The weakness is **structural**: a good architecture
was built and left unplugged while the old one accreted duplication. Fixing C1
(wire the registry, delete the three dispatchers) is the single highest-leverage
change; it directly resolves R1/R2/R3 and is the thing a reviewer notices first.

### Prioritized punch-list (independent of the A/B/C decision)

1. **C1** — wire the registry as the one dispatch path; delete the three old
   dispatchers (the migration plan, §5).
2. **C3 + C2** — split `scraper` into `providers/*` + `netx`; break the 2k-line
   god files along their seams (download, player) into focused files.
3. **C5** — make `gofmt` (and `golangci-lint`) a hard CI gate; fix the drift once.
4. **C4** — add a one-line `doc.go` to each package.
5. **C7** — add `CONTRIBUTING.md`, `SECURITY.md`, `CODE_OF_CONDUCT.md`; keep this
   `ARCHITECTURE.md` linked from the README.
6. **C8** — rename the hack files; move `TEST_*.md` to `docs/testing/`; `claude.md`
   → `CLAUDE.md`.
7. **C6** — audit direct deps; drop anything a small CLI doesn't need.
8. **S1 + S2 + S3** (from §7, validated by Curd v2.0.0) — while wiring the
   registry (C1), add a `DisabledProviders` config kill-switch (S1) and a
   `curdhost`-style host-services package (S2), and keep all self-registers in one
   wiring file (S3). All three are low-cost and harden R3/R5/R6 without touching
   user-visible behavior. Keep the context-first interface — do **not** adopt
   Curd's `ctx`-less signatures.

None of these require changing user-visible behavior; all are gated by the green
suite and reversible.
