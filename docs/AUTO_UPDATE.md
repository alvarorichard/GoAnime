# Updating GoAnime

GoAnime can replace itself with the latest release:

```bash
goanime --update
```

The code lives in `internal/updater`.

## What it does

1. Asks the GitHub API for the latest release and compares its version with
   the running one.
2. If the release is newer, shows both versions and the first 300 characters
   of the release notes, and asks whether to update now.
3. On yes, downloads the release asset for the current platform:

   | Platform | Asset                                                                 |
   | -------- | --------------------------------------------------------------------- |
   | Windows  | `goanime-windows-<arch>.zip`, from which the executable is extracted   |
   | macOS    | `goanime-darwin-<arch>` when published, otherwise `goanime-darwin-universal` |
   | Linux    | `goanime-linux-<arch>`                                                |

4. Downloads `checksums-sha256.txt` from the same release and checks the
   asset's SHA-256 digest against it. If the release does not publish that
   file, the asset is not listed in it, or the digest differs, the update
   stops and nothing is installed.
5. Hashes the verified new executable and copies the current one to
   `<path>.backup`.
6. Replaces the executable (see the platform differences below).
7. Writes `<path>.update-pending` with the SHA-256 of the installed binary.

The backup is **not** deleted when the update command returns. On the next
start, `updater.CleanupAfterUpdate()` hashes the running executable and
compares it with the marker. Only a match removes the backup and the temporary
files. If the new binary does not start, or its hash differs, the backup is
kept so the previous version can be restored.

When there is nothing to update, it prints `You are running the latest
version!`. When you decline, it prints `Update cancelled by user`.

## Platform differences

On Linux and macOS the running executable is moved aside and the new one is
copied to the original path, which works while GoAnime is running. The new
version is used from the next start.

On Windows a running executable is locked. If GoAnime cannot move it, it hands
the replacement to a helper that runs after GoAnime exits, and prints `Update
will complete when you restart the application`.

### Windows: the post-exit helper

1. The verified new executable is copied to `goanime.exe.new` **next to the
   installed `goanime.exe`**. This is a stable path: the temporary download and
   extraction directory can be deleted before the helper runs.
2. `goanime.exe.update.bat` is written and started with the paths in
   `GOANIME_UPDATE_*` environment variables, so spaces, `&` and parentheses in
   `%LOCALAPPDATA%` do not break the batch file.
3. The helper waits for the GoAnime process ID to disappear, then:
   - checks the SHA-256 of `goanime.exe.new` with `certutil -hashfile`;
   - copies the current executable to `goanime.exe.backup` if no backup exists;
   - copies `.new` over `goanime.exe`, retrying up to 15 times;
   - verifies the SHA-256 of the installed file;
   - on success, writes `goanime.exe.update-pending` and deletes `.new` and
     itself;
   - on failure, restores `goanime.exe.backup`, deletes `.new`, and leaves the
     backup in place.

The helper deletes itself on success and on failure.

## Files next to the executable

| File | Purpose | Removed |
| --- | --- | --- |
| `goanime.exe.backup` | Previous working binary | After the new binary starts and its hash matches the marker |
| `goanime.exe.new` | Staged new binary | By the helper after success or failure |
| `goanime.exe.update-pending` | SHA-256 of the installed binary | After a verified start |
| `goanime.exe.update.bat` | Post-exit helper | By the helper |

If a backup is left behind because the update did not boot, close GoAnime and
rename `goanime.exe.backup` back to `goanime.exe`. The next start notices that
the running binary no longer matches the marker, keeps the backup and removes
the stale marker.

## Updating from 1.x on Windows

The 1.x updater writes the old helper, so it cannot install the corrected one
by itself. Windows users still on 1.x should download and run the v2.0.0
installer (or the portable ZIP) once. From v2.0.0 onward, `--update` uses the
flow above.

## Safety

- The release lookup and every download start at a GitHub address over HTTPS
  (`github.com`, `api.github.com`, `objects.githubusercontent.com` or
  `github-releases.githubusercontent.com`); any other starting URL is refused.
  GitHub redirects asset downloads to its own storage, and that redirect is
  followed.
- Every download is checked against the release's `checksums-sha256.txt`
  before anything is installed, and on Windows before the zip is extracted.
  This catches a corrupted, truncated or altered download. It does not protect
  against a compromised release, since the checksums come from the same place
  as the asset. Release signing and attestations are tracked separately.
- The Windows binaries are not Authenticode-signed. Windows SmartScreen may
  warn on first launch. See the README for how to verify a download and what
  to do with the warning.

## Installations managed by a package manager

`--update` replaces the executable in place. If you installed GoAnime from the
AUR, Scoop, winget or Homebrew, update it through that package manager instead,
so the manager keeps track of the installed version.
