package netx

// UserAgent is the shared browser User-Agent presented by the plain-HTTP
// scrapers (AniDB, AnimeFire, Goyabu). SuperFlix declares its own because
// its UA must match the browser that solves the Cloudflare challenge.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"

// AcceptLanguage is the Accept-Language that goes with UserAgent: Firefox's own
// q-ladder for a pt-BR desktop.
//
// It is a constant here, and shared, so the pairing cannot drift apart. The
// scrapers used to hardcode "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7" — Chrome's
// ladder — under a Firefox User-Agent. That combination cannot come from a real
// browser, and it is the stock line copied through countless scraping snippets,
// so it is exactly the kind of thing an anti-bot rule keys on.
//
// This matters more than it looks in this codebase. SuperFlix's CDN matches
// Accept-Language BY VALUE and 403s anything but one specific string (see
// superflix/cdn.go), and on 2026-09-21 its API host answered 429 to the Chrome
// ladder while serving every other value — a correlation that held across eight
// interleaved probes and then stopped reproducing a few hours later, so treat it
// as reputation-sensitive rather than a fixed rule. Either way, presenting a
// coherent browser costs nothing and removes a whole class of surprise.
const AcceptLanguage = "pt-BR,pt;q=0.8,en-US;q=0.5,en;q=0.3"

// ChromeAcceptLanguage is Chrome's q-ladder, for the few requests that
// deliberately present a Chrome User-Agent instead of the shared Firefox one.
//
// The string itself is not the problem — it is Chrome's genuine header. The
// problem was sending it under a Firefox UA, a pair no real browser produces.
// So there are two constants, and the rule is simply that the ladder must match
// the browser the UA claims (see useragent_pairing_test.go).
const ChromeAcceptLanguage = "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7"

// EnglishAcceptLanguage is the Firefox ladder for an English desktop, used by
// sources with no Portuguese catalog (AniDB) so the request does not claim a
// locale its content has nothing to do with.
const EnglishAcceptLanguage = "en-US,en;q=0.5"

// APIUserAgent identifies GoAnime to first-party JSON/GraphQL APIs (AniList) as
// an ordinary API client.
//
// It deliberately does NOT look like a browser — that is the whole point. AniList
// answers browser User-Agents with an HTTP 403 whose body reads "The AniList API
// has been temporarily disabled due to severe stability issues", while serving
// plain API clients normally. Verified live against graphql.anilist.co: a curl/Go
// UA (or none) returns 200; a Firefox or Chrome UA returns 403. See issue #184.
//
// Requests using it must also travel on a PLAIN net/http client: the shared surf
// clients impersonate Chrome and overwrite the User-Agent, so no header the
// caller sets can survive them.
const APIUserAgent = "GoAnime/1.0 (+https://github.com/alvarorichard/GoAnime)"
