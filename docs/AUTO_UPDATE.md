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
5. Copies the running executable to `<path>.backup`, puts the new one in its
   place and deletes the backup. If the replacement fails, the backup is
   restored.

When there is nothing to update, it prints `You are running the latest
version!`. When you decline, it prints `Update cancelled by user`.

## Platform differences

On Linux and macOS the running executable is moved aside and the new one is
copied to the original path, which works while GoAnime is running. The new
version is used from the next start.

On Windows a running executable is locked. If GoAnime cannot move it, it
writes an update script that finishes the replacement after GoAnime exits,
and prints `Update will complete when you restart the application`.

## Safety

- The release lookup and every download start at a GitHub address over
  HTTPS (`github.com`, `api.github.com`, `objects.githubusercontent.com` or
  `github-releases.githubusercontent.com`); any other starting URL is refused.
  GitHub redirects asset downloads to its own storage, and that redirect is
  followed.
- Every download is checked against the release's `checksums-sha256.txt`
  before anything is installed, and on Windows before the zip is extracted.
  This catches a corrupted, truncated or altered download. It does not
  protect against a compromised release, since the checksums come from the
  same place as the asset; that would need signed releases.

## Installations managed by a package manager

`--update` replaces the executable in place. If you installed GoAnime from the
AUR, update it through your AUR helper instead, so the package manager keeps
track of the installed version.
