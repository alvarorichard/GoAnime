# AnimesFHD API fixtures

Copied byte-for-byte on 2026-11-11 from the site owner's supplied public JSON
samples (the date is the fixture import date, not a new live capture).

- `search`: `/api/series?pagination=1&page=1&per_page=5&q=frieren`
- `popular`: `/api/series` popular listing
- `series`: `/api/series/1709160706373932?page=1&per_page=200`
- `episode`: `/api/episodes/22882`
- `movie`: `/api/series/2110072639862535?page=1&per_page=200`
- `movie_episode`: `/api/episodes/1049086484300956`

Pagination, redirects, invalid data and multiple versions are synthetic HTTP
responses in `client_test.go`; the supplied samples are not modified.
