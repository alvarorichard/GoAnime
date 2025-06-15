# Build Options

GoAnime can be built with or without SQLite support for tracking anime progress.
Here's how to choose the right build for your needs:

## Standard Build (No SQLite)

The standard build is compiled without CGO and doesn't include SQLite tracking
features. This build:

- Has smaller binary size
- Works on all platforms without dependencies
- Doesn't require any system libraries
- No progress tracking features

To create a standard build:

```bash
cd build
./buildlinux.sh   # For Linux
./buildmacos.sh   # For macOS
./buildwindows.sh # For Windows
```

## Build With SQLite Support

This build includes full anime progress tracking features using SQLite:

- Remembers playback position for all episodes
- Allows resuming from where you left off
- Tracks watched episodes
- Requires SQLite development libraries on the system
- Larger binary size due to CGO dependencies

To create a build with SQLite support:

```bash
cd build
./buildlinux-with-sqlite.sh   # For Linux
```

### System Requirements for SQLite Support

For builds with SQLite support, you'll need:

- GCC or compatible C compiler
- SQLite development libraries:
  - Ubuntu/Debian: `sudo apt-get install libsqlite3-dev`
  - Fedora: `sudo dnf install sqlite-devel`
  - Arch Linux: `sudo pacman -S sqlite`
  - macOS: `brew install sqlite3`

## Troubleshooting

If you get an error like:

```text
panic: database initialization failed: schema creation failed: Binary was 
compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
```

It means you're using a standard build without SQLite support. You have two options:

1. Build with SQLite support using the instructions above
2. Continue using the standard build - anime progress tracking will be disabled,
   but all other features will work

## How to Check Your Build

To check if your build supports SQLite tracking:

```bash
./goanime --version
```

SQLite-enabled builds will show "with SQLite tracking" in the version information.
