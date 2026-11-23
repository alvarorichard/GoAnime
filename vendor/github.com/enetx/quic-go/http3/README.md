# HTTP/3

[![Documentation](https://img.shields.io/badge/docs-quic--go.net-red?style=flat)](https://quic-go.net/docs/)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/enetx/quic-go/http3)](https://pkg.go.dev/github.com/enetx/quic-go/http3)

This package implements HTTP/3 ([RFC 9114](https://datatracker.ietf.org/doc/html/rfc9114)), including QPACK ([RFC 9204](https://datatracker.ietf.org/doc/html/rfc9204)) and HTTP Datagrams ([RFC 9297](https://datatracker.ietf.org/doc/html/rfc9297)).
It aims to provide feature parity with the standard library's HTTP/1.1 and HTTP/2 implementation.

Detailed documentation can be found on [quic-go.net](https://quic-go.net/docs/).

This fork tracks quic-go **v0.65.0** and uses `github.com/enetx/http`.
It preserves configurable request header order and ordered HTTP/3 settings.
See [UPSTREAM.md](UPSTREAM.md) for the fork audit, retained changes and Surf API usage.

Run the upstream and fork regression suites with `go test -race ./...` and
`go vet ./...`.
