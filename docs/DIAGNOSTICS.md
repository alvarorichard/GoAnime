# Diagnostics: debug logs and crash reports

When something goes wrong, the fastest fix starts with a log that says what
happened on *your* machine. `--debug` writes one. A crash writes one by itself.
Both are meant to be attached to a bug report **as they are**.

## Turning it on

```bash
goanime --debug "naruto"
```

GoAnime prints where the file is and how to follow it from another terminal.
Each run writes a new file; the newest 30 are kept and older ones are removed.

| System  | Folder                         |
| ------- | ------------------------------ |
| Windows | `%LOCALAPPDATA%\GoAnime\logs`  |
| macOS   | `~/Library/Logs/GoAnime`       |
| Linux   | `~/.local/share/goanime/logs`  |

## What the log starts with

Right after the header (which carries a **session id**, so a log and a crash
report from the same run can be matched) comes an environment report:

| Section     | What it says                                                                 |
| ----------- | ---------------------------------------------------------------------------- |
| Application | version, Go version, commit, build time, where the program runs from         |
| System      | OS and architecture, distribution, kernel, memory, CPUs, time zone, locale; WSL or container |
| Terminal    | `TERM`, `COLORTERM`, terminal program, multiplexer, SSH, shell, colour support, size |
| Environment | every `GOANIME_*` variable and any proxy (its password removed)              |
| Tools       | path and version of `mpv`, `yt-dlp`, `ffmpeg`, `ffprobe`, or "not found"     |
| Paths       | log, config and cache folders and the free space where logs are written      |
| Sources     | each provider: enabled, switched off, or off by default, and by which setting |
| cfsolve     | whether the Cloudflare helper is installed and whether it may download itself |

Then, in the background and without slowing the start, GoAnime asks each
enabled provider's homepage how it answers and logs one line each:

```
Provider check source=Goyabu result=challenge ms=1218 status=403 detail="cloudflare challenge (challenge)"
Provider check source=AnimeDrive result=ok ms=2462 status=200
```

`result` is `ok`, `challenge` (Cloudflare wants a check), `blocked`, `down`,
`slow`, `dns`, `refused`, `tls` or `error`. A provider that moved to another
domain also carries `moved_to=`.

Each line is the time and what happened. The few `INFO`, `WARN` and `ERROR`
lines are labelled so they stand out; the rest (almost all) are not. Set
`GOANIME_LOG_SOURCE=1` to add the source file and line of each message, which
helps when debugging GoAnime itself.

At the end of the file there is a one-line summary of the run: how long it
lasted and how many warnings and errors it logged.

## What never reaches the file

Everything written to the log passes through one filter first. It removes:

- credentials in URLs (`user:password@host`) and in `Authorization` and
  `Cookie` headers;
- query parameters and fields named like a secret: `token`, `secret`,
  `signature`, `ticket`, `password`, `session`, `clearance`, `api_key`,
  `key`, `sig`, and signed-URL parameters such as `X-Amz-Signature`;
- the one-time tickets the RedeFlix player hands out and JSON web tokens;
- your home folder, shown as `~`.

An ordinary search (`keyword=naruto`) is left alone, so the log stays useful.
The filter is applied to log lines, to the environment report, to output of
programs GoAnime runs and to crash reports.

It does **not** remove the titles you searched for or watched, because the bug
usually depends on them. Look through the file before sharing it if that
matters to you.

## Crash reports

If GoAnime panics it saves `crash_<date>_<time>_<id>.log` in the same folder,
**with or without `--debug`**, and prints the path. It holds the panic, the
stack trace and the environment report. The newest 10 are kept, separately from
the session logs.

## For contributors

- `internal/diag` collects the report, probes the providers and holds the
  redaction rules (`Redact`). It imports nothing from the app.
- `internal/util/logger.go` owns the log file. Every write goes through
  `logSink`, which redacts; do not write to the file any other way.
- `internal/debugreport` wires it to the program (sources, cfsolve, paths).
- Register work that should run when a debug session opens with
  `util.OnDebugSession`.
