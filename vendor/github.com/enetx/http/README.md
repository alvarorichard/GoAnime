# HTTP transports for Surf

A Go **1.27+** fork containing the HTTP packages used by Surf and their tests:

- `github.com/enetx/http`: HTTP/1, client/server types and shared configuration.
- `github.com/enetx/http/http2`: public adapters and frame aliases for the single
  native HTTP/2 engine in `internal/http2`.
- `cookiejar`, `httptrace`, `httptest`, and the request/response dump helpers in
  `httputil` used by Surf and the protocol tests.
- `headerorder`: the stateless header sorter shared with
  `github.com/enetx/quic-go/http3`.

HTTP/3 lives in the separate `github.com/enetx/quic-go` module. This module has no
QUIC or QPACK dependency. It preserves Surf's header and pseudo-header order,
ordered HTTP/2 settings, stream IDs, flow control and priority frames.
Profiling, CGI/FastCGI, reverse proxy, old Go compatibility and the old HTTP/2
bundle are outside this fork's scope.

Public-API tests live in `tests/`. Tests requiring private access remain beside
their source packages.

## Validation

```sh
go generate ./http2
go test -race -count=1 ./...
go vet ./...
python3 -m unittest discover -s tests/upstream
```

To validate local changes from Surf, temporarily replace this module:

```sh
go mod edit -replace=github.com/enetx/http=../http
go test -race -count=1 ./...
go mod edit -dropreplace=github.com/enetx/http
```

Published Surf versions use released module versions without local replacements.

## Updating upstream

[upstream.json](upstream.json) pins the Go and x/net source versions and commits.
[UPSTREAM.md](UPSTREAM.md) describes retained patches and the update procedure.
The helper requires Python 3.9+, Go and Git.

```sh
python3 scripts/upstream.py audit
python3 scripts/upstream.py prepare --go go1.27.3 --net v0.62.0
```

Preparation downloads/caches source and writes a separate candidate tree,
`changes.diff` and `summary.json`. It preserves this checkout. Review conflicts,
run validation and Surf's tests against the candidate, then transfer the changes.
