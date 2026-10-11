# Build options

Progress tracking (resuming playback and remembering watched episodes) stores
its data in SQLite through [go-sqlite3](https://github.com/mattn/go-sqlite3),
which needs CGO. Everything else in GoAnime works without it.

The release binaries are built with tracking. `go build` and `go install`
include it automatically when a C compiler is available, and leave it out
when none is found.

## With tracking

You need a C compiler: gcc or clang on Linux and macOS, or MinGW-w64 on
Windows. SQLite itself is compiled into the binary, so no system SQLite
library is required.

```bash
CGO_ENABLED=1 go build ./cmd/goanime
```

For a stripped Linux binary, `build/buildlinux-with-sqlite.sh` does the same.
Run it from the `build` directory:

```bash
cd build
./buildlinux-with-sqlite.sh
```

## Without tracking

```bash
CGO_ENABLED=0 go build ./cmd/goanime
```

The other scripts in `build/` produce builds without tracking:

```bash
cd build
./buildlinux.sh     # Linux
./buildmacos.sh     # macOS, amd64 and arm64
./buildwindows.sh   # Windows
```

A build without tracking prints this notice when you search for something to
watch, and otherwise works normally:

```text
Notice: Anime progress tracking disabled (CGO not available)
Episode progress and resume features will not be available.
```

## Checking a build

```bash
goanime --version
```

The output ends in `(with SQLite tracking)` or `(without SQLite tracking)`.
