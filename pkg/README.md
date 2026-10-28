# pkg

Public Go packages of GoAnime.

- [goanime](goanime) is the API for searching sources, listing episodes and
  getting stream URLs. Its reference is in [goanime/README.md](goanime/README.md),
  and [docs/LIBRARY_USAGE.md](../docs/LIBRARY_USAGE.md) walks through it.
- [goanime/examples](goanime/examples) has complete programs.

Everything under `internal/` is private to GoAnime and can change without
notice. Only the packages here are meant to be imported by other projects.
