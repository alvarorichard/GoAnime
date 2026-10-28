<!--
Pull requests go to the dev branch, not main. If the base branch above says
main, change it to dev before opening.
-->

## What this changes

<!-- What the change does and why. Link the issue it fixes, e.g. "Fixes #123". -->

## How it was tested

<!-- Commands you ran, sources you tried, and on which operating system. -->

## Checklist

- [ ] The base branch is `dev`.
- [ ] `go vet ./...`, `golangci-lint run ./...`, `gosec ./...` and `govulncheck ./...` pass, or each remaining finding is explained above.
- [ ] `go test -short -race ./...` passes, and new behavior has tests.
- [ ] `README.md` and `README_pt-br.md` are updated if user-facing behavior changed.
