# Adding and Removing a Source

> **Reference implementation: `Goyabu`.** It is the simplest complete source in
> the tree — no seasons, no browser gate, no quality argument. Read it, copy it,
> rename it. This document lists *where* to copy from and *what* to change; it
> deliberately contains no invented example code, because invented code in docs
> rots without anything failing.
>
> Verified against the tree at v1.8.6.

---

## Where a source lives

```
internal/scraper/providers/<name>/          leaf client — HTTP + parsing only
internal/scraper/manager.go                 UnifiedScraper adapter (ScraperType enum)
internal/api/providers/source_providers.go  the source.Source, registered in init()
internal/api/source/                        registry: Register / Resolve / ActiveSources
```

A source describes itself as **data** (`Describe() source.Descriptor`) and
registers itself in `init()`. Resolution, search fan-out, circuit breaker and
the kill-switch all read that data — **nothing in the dispatch path needs an
edit**. Everything else in the table below is bookkeeping the migration has not
yet eliminated (see [Known friction](#known-friction)).

---

## Touchpoints

Read the Goyabu column as "copy this, rename it".

| # | File | Where Goyabu does it | Required |
|---|---|---|:---:|
| 1 | `internal/scraper/providers/<name>/{doc.go,client.go}` | whole package | ✅ |
| 2 | `internal/scraper/manager.go` | import `:22` · `ScraperType` const `:32` · `NewAdapter` case `:54` · `scraperDisplayName` `:71` · `scraperLanguageTag` `:87` · adapter type `:253` | ✅ |
| 3 | `internal/api/source/kind.go` | `SourceKind` const `:14` · `scraperTypeMap` `:32` | ✅ |
| 4 | `internal/api/providers/source_providers.go` | provider block `:225–299` | ✅ |
| 5 | `internal/api/providers/tagging.go` | `sourceDisplayName` `:33` · `isPTBR` `:64` (+ `languageTag` if not PT-BR) | ✅ |
| 6 | `internal/api/providers/naming/naming.go` | `tagPattern` regex `:75` | ✅ |
| 7 | `internal/scraper/source_health.go` | `healthTargets` `:49` · `DefaultHealthCheckQuery` `:38` | 🟡 |
| 8 | `internal/api/enhanced.go` | `--source` case `:247` · `ptbr` group `:252` · backfill by name `:290` · backfill by URL `:302` · debug print `:320` · `sourceBreakdown` field `:939` · `countSourceBreakdown` case `:959` | 🟡 |
| 9 | `internal/api/anime.go` | `reSpaceDashNoise` source alternation `:540` | 🟡 |
| 10 | `internal/util/util.go` | `--source` help text `:461` — currently stale: lists `flixhq`, omits `goyabu` and `superflix` | ⚪ |
| 11 | `pkg/goanime/types/source.go` | public SDK enum — **breaking change** | ⚪ |
| 12 | tests | see [Tests](#tests) | ✅ |

Skip #6 and your tag is never stripped from titles → wrong AniList matches. Skip
#7 and the source is never health-probed → silent breakage. Skip #8 and
`--source <name>` reports an unknown source.

Nothing to wire for `init()`: the blank import already exists at
`internal/player/scraper.go:24`, and the provider lives in that same package.

---

## Adding

Build after each step — every one leaves the tree green.

**1. Leaf client.** Copy `internal/scraper/providers/goyabu/`. Keep the shape:
`New<Name>Client()` (no network I/O — it runs under `sync.Once`), `SearchAnime`,
`GetAnimeEpisodes`, `GetEpisodeStreamURL`, `NewClientForTest`, plus the
`decorateRequest` / `shouldRetry` / `sleep` / `resolveURL` helpers.
Non-negotiable: `util.NewFastClient()` for the HTTP client (SSRF-guarded),
regexes compiled at package level (`client.go:32`), errors via
`netx.NewParserError` / `NewBlockedChallengeError` / `NewHTTPStatusError` —
`DiagnoseError` and the circuit breaker classify by type. Import nothing from
`api/`, `player/`, `playback/`; this is a leaf.

**2. Adapter.** `manager.go`, the six sites in the table. `ScraperType` is an
`iota`: **append at the end, never insert.** If your stream needs a referer,
subtitles or an audio language, put them in `GetStreamURL`'s
`metadata map[string]string` under `referer` / `subtitles` / `subtitle_labels` /
`audio_lang` — the contract `SuperFlixAdapter` uses and the player reads.

**3. Kind.** `kind.go` — string constant plus the `scraperTypeMap` entry.

**4. Provider.** Copy `source_providers.go:225–298`. The only part that is design
rather than boilerplate is the descriptor:

```go
func (p *goyabuProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.Goyabu,
		Priority:    20,
		Explicit:    []string{"Goyabu"},
		Tags:        []string{"[goyabu]"},
		URLMatchers: []string{"goyabu"},
		ProbeURL:    "https://goyabu.io",
	}
}
```

| Field | Meaning | Skip when |
|---|---|---|
| `Kind` | identity — `Register` panics if empty | never |
| `Priority` | lower matched first among non-explicit candidates | never (0 is valid but greedy) |
| `Explicit` | values that may appear in `anime.Source`; checked before everything else | never — this is how a saved anime routes back |
| `Tags` | lowercase substrings of `anime.Name` | results carry no tag |
| `URLMatchers` | lowercase substrings of `anime.URL` | source uses opaque IDs |
| `MediaTypes` | `models.MediaType` values routed here | anime-only source |
| `ShortID` | accepts bare alphanumeric IDs | always — AllAnime-specific |
| `DefaultDisabled` | ships off unless `GOANIME_ENABLED_SOURCES` names it | shipping live |
| `ProbeURL` | homepage; HEAD-probed on search timeout to tell "site down" from "opaque hang" | GraphQL/opaque APIs, browser-gated sources |

Priorities in use: AnimeFire `10` · Goyabu `20` · SuperFlix `30` · AllAnime `40`.
Leave gaps of 10. Priority is ignored when `anime.Source` matches an `Explicit`
entry.

`FetchStreamURL` **must** open with `util.ClearGlobalSubtitles()` and
`util.SetGlobalAnimeSource(anime.Source)` (`source_providers.go:283–286`) — skip
them and the previous episode's subtitles leak into this one.

**Capabilities** are discovered by type assertion, not by a flag. Implement only
what is true: `HasSeasons() bool` → `source.Seasoned` · `WarmUp(ctx) error` →
`source.BrowserGated` (called before every stream fetch; see `superFlixProvider`)
· `Search(ctx, query)` → `source.Searchable`. **A source without `Search` is
silently excluded from the search fan-out** — it can still play by URL.

**5–11.** Mechanical; follow the table. One trap: `sourceDisplayName` must return
a string that appears in your `Descriptor.Explicit`, or a saved anime will not
resolve back to you. (`AnimeFire` returns `"Animefire.io"`, which is why its
`Explicit` lists both spellings.)

### Tests

One test per function, table-driven, `t.Parallel()`, `httptest.Server` for every
HTTP mock — never real network. Copy `goyabu/client_test.go`. Add the new kind to
the enumerating tests: `internal/api/providers/source_providers_test.go:147`,
`capabilities_test.go:23,39`, `internal/scraper/source_health_test.go:32,51`,
`internal/api/enhanced_results_test.go:62`. Pin the host with a dated assertion so
a domain rotation fails loudly. Gate live tests behind
`testing.Short() || os.Getenv("CI") != ""`.

---

## Removing

Pick the weakest level that solves the problem.

**Level 1 — runtime, no rebuild.** Correct first response to a source that broke
overnight:

```bash
GOANIME_DISABLED_SOURCES="Goyabu"              # comma-separated
GOANIME_DISABLED_SOURCES="goyabu,animefire.io" # case-insensitive, dot-forgiving
```

Drops it from `ActiveSources()`, `Resolve`, the fan-out and the best-effort
fallback. Each skip is logged at debug level — never silent.

**Level 2 — ship it disabled.** `DefaultDisabled: true` in the descriptor; users
opt back in with `GOANIME_ENABLED_SOURCES`. Code and tests stay, so reviving it
is one boolean. Use for fragile-but-not-dead sources.

**Level 3 — delete.** Only when the site is gone for good. **Follow this order** —
it removes references before the things they reference, so the build stays green
at every step:

```
1.  source_providers.go     delete the whole provider block
2.  tagging.go              sourceDisplayName/languageTag cases; drop from isPTBR
3.  naming/naming.go        drop from the tagPattern alternation
4.  enhanced.go             all 7 sites from the touchpoint table
5.  anime.go                drop from the reSpaceDashNoise alternation
6.  util.go                 drop from the --source help text
7.  source_health.go        healthTargets + DefaultHealthCheckQuery
8.  kind.go                 SourceKind const + scraperTypeMap entry
9.  manager.go              adapter type, NewAdapter case, both name switches,
                            import, ScraperType const   ⚠️ iota
10. rm -r internal/scraper/providers/<name>/
11. tests                   remove from every enumerating test
12. pkg/goanime/types/      only if publicly exposed (breaking change)
13. go mod tidy             if it pulled deps nobody else uses
```

> ⚠️ **`ScraperType` iota (step 9).** Deleting a middle constant renumbers
> everything after it. Safe *today* — no `ScraperType` value is persisted to disk
> or the wire; it is process-lifetime only. Confirm that still holds
> (`rg 'ScraperType' --glob '!*_test.go'`) before renumbering. If unsure, move the
> dead constant to the end of the block and mark it deprecated instead.

Confirm the removal is clean: `rg -i '<name>' --stats` returns zero hits outside
`CHANGELOG.md`.

---

## Troubleshooting

| Symptom | Cause | Fix |
|---|---|---|
| Source never appears in search results | no `Search` method → not `Searchable` → excluded from fan-out | implement `Search` |
| Source resolves for everything | `Priority: 0` + broad `URLMatchers` (e.g. `"tv"`) | raise priority, tighten matchers |
| Saved anime replays on the wrong source | `sourceDisplayName` string absent from `Descriptor.Explicit` | make them match, or list both spellings |
| Source tag leaks into AniList lookups | missing from `naming.go` `tagPattern` | add to the alternation |
| Previous episode's subtitles leak in | `util.ClearGlobalSubtitles()` missing | first statement of `FetchStreamURL` |
| Panic at startup: empty `Descriptor.Kind` | `Describe()` returns a zero `Kind` | set `Kind` |
| Source stops being tried after a few failures | circuit breaker opened — working as designed | fix the root cause; it auto-recovers after the cooldown |
| Timeout reported as a generic hang | `ProbeURL` empty | set it to the public homepage |
| Search slow for everyone | your `Search` blocks past the 12s per-source budget | fix the client; the breaker opens regardless |
| CI hangs on Windows | ungated live/TTY test | guard with `testing.Short()` / `CI` |

---

## Verification

```bash
go build ./... && go vet ./...
golangci-lint run ./...
gosec ./...
go test -short -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | tail -1        # must stay ≥ 66.0%

go run ./cmd/goanime --source <name> "naruto"
GOANIME_DEBUG=1 go run ./cmd/goanime "naruto"                 # your source in the fan-out
GOANIME_DISABLED_SOURCES=<Name> go run ./cmd/goanime "naruto" # kill-switch sees it
```

---

## Known friction

`ARCHITECTURE.md` §2 (R2) promises "adding a source = one file". Today it is
**12 files, ~20 edit sites**. Only #1, #3 and #4 are the architecture; the rest is
migration debt, with four named causes in descending order of cost:

1. **`UnifiedScraper` + `ScraperType` are a second registry** running parallel to
   the real one, and every adapter is a pure forwarding shim. If providers held
   their leaf client directly, touchpoints #2 and #3 vanish.
2. **Source identity is duplicated four times** — `Descriptor.Explicit`,
   `sourceDisplayName`, `scraperDisplayName`, `languageTag`/`scraperLanguageTag`.
   Moving `DisplayName` and `LanguageTag` into the `Descriptor` deletes two
   switches; the `"Animefire.io"` vs `"AnimeFire"` split is that duplication
   already producing a workaround.
3. **`healthTargets()` and the `ptbr` group are hardcoded lists** that should be
   derived from `ActiveSources()` and `Descriptor.Tags`.
4. **`enhanced.go` hand-maintains source names in five places**, including two
   backfill switches that duplicate what `source.Resolve` already computes.

Fixing 1–3 reduces adding a source to one new package plus one provider block.
Until then, trust this table over the architecture doc's ideal.

---

**See also:** [`ARCHITECTURE.md`](ARCHITECTURE.md) ·
[`architecture/ARCH_STAGES.md`](architecture/ARCH_STAGES.md) ·
[`Development.md`](Development.md) · [`testing/TEST_STRATEGY.md`](testing/TEST_STRATEGY.md)
