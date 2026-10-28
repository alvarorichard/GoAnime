# Contributing to GoAnime

Thanks for taking the time to contribute. This document covers what you need
to open a useful issue or get a pull request merged. The full development
guide, including code style and the Go 1.27 conventions the project relies on,
is in [docs/Development.md](docs/Development.md).

Issues and pull requests can be written in English or Portuguese.

## Reporting a bug

Search the [existing issues](https://github.com/alvarorichard/GoAnime/issues)
first. If nothing matches, open a
[bug report](https://github.com/alvarorichard/GoAnime/issues/new?template=bug_report.yml).
The form asks for the output of `goanime --version`, your operating system,
the command you ran, the source, and a debug log.
[Reporting a problem](README.md#reporting-a-problem) in the README explains how
to produce the log and where to find it on each system.

Most bugs are a source that stopped working because the site behind it
changed. The [source health workflow](https://github.com/alvarorichard/GoAnime/actions/workflows/source-health.yml)
checks every source daily, so look there to see whether it is already known.

Questions and ideas that are not bugs are better asked on
[Discord](https://discord.gg/FbQuf78D9G) or in
[Discussions](https://github.com/alvarorichard/GoAnime/discussions).

## Finding something to work on

Issues labeled
[good first issue](https://github.com/alvarorichard/GoAnime/labels/good%20first%20issue)
are small and self-contained. Fixing a broken source is the other common
starting point: each source is a separate package under
`internal/scraper/providers`, and
[docs/ADDING_A_SOURCE.md](docs/ADDING_A_SOURCE.md) explains how sources fit
into the rest of the code.

If you plan a larger change, open an issue first so the approach can be agreed
on before you write the code.

## Setting up

You need Go 1.27.3 or later, mpv, and a C compiler if you want progress tracking
(SQLite requires CGO).

```bash
git clone https://github.com/alvarorichard/GoAnime.git
cd GoAnime
git checkout dev
go build ./cmd/goanime
go test -short ./...
```

## Branches

Development happens on `dev`. Create your branch from `dev` and open the pull
request against `dev`. `main` only receives releases, so pull requests against
it will be retargeted.

## Before opening a pull request

CI runs the following on every pull request. Run them locally first:

```bash
gofmt -l .
go vet ./...
golangci-lint run ./...
gosec -exclude-generated ./...
govulncheck ./...
go test -short -race -count=1 ./...
```

Install instructions for each tool are in
[docs/Development.md](docs/Development.md#quality-checks). Any finding you
leave unfixed must be justified in the pull request description.

New features need tests, and bug fixes need a test that fails without the fix.
Documentation-only changes do not.

If you change user-facing behavior, update both `README.md` and
`README_pt-br.md`.

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
fix(animefire): handle the new episode list markup
feat: add a --last flag to resume the last watched title
docs: document GOANIME_HIANIME_AUDIO
```

## License

By contributing, you agree that your contributions are licensed under the
[MIT License](LICENSE).
