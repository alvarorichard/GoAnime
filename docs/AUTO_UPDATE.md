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

4. Copies the running executable to `<path>.backup`, puts the new one in its
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

- Release data and assets are downloaded over HTTPS, and only from
  `github.com`, `api.github.com`, `objects.githubusercontent.com` and
  `github-releases.githubusercontent.com`.
- The download is not yet verified against the `checksums-sha256.txt` file
  published with each release.

## Installations managed by a package manager

`--update` replaces the executable in place. If you installed GoAnime from the
AUR, update it through your AUR helper instead, so the package manager keeps
track of the installed version.
