package netx

// UserAgent is the shared browser User-Agent presented by the plain-HTTP
// scrapers (AniDB, AnimeFire, Goyabu). SuperFlix declares its own because
// its UA must match the browser that solves the Cloudflare challenge.
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/121.0"

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
