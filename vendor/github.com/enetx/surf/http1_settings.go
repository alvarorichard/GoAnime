package surf

import (
	"fmt"

	"github.com/enetx/http"
	h3transport "github.com/enetx/surf/internal/http3"
	"github.com/enetx/surf/profiles"
)

// HTTP1Settings configures owned HTTP/1 request buffers and write packing.
// Browser defaults apply first; explicit fluent overrides apply last.
type HTTP1Settings struct {
	builder   *Builder
	overrides []func(*profiles.HTTP1Config)
}

// HTTP1Settings starts the nested HTTP/1 request-writing builder.
func (b *Builder) HTTP1Settings() *HTTP1Settings {
	if b.http1settings == nil {
		b.http1settings = &HTTP1Settings{builder: b}
	}

	return b.http1settings
}

// WriteBufferSize selects the outgoing HTTP/1 buffer size, zero through 1 MiB.
// Zero restores the native default.
func (h *HTTP1Settings) WriteBufferSize(size int) *HTTP1Settings {
	h.overrides = append(h.overrides, func(c *profiles.HTTP1Config) { c.WriteBufferSize(size) })
	return h
}

// BodyReadSize bounds body reads, zero through 1 MiB. Zero retains native copy
// optimizations. Explicit sizes also apply to file and streaming bodies.
func (h *HTTP1Settings) BodyReadSize(size int) *HTTP1Settings {
	h.overrides = append(h.overrides, func(c *profiles.HTTP1Config) { c.BodyReadSize(size) })
	return h
}

// HeaderBodyMergeLimit selects the largest combined header/in-memory-body write.
// Zero separates nonempty bodies; the valid range is zero through 1 MiB.
// Potentially blocking bodies and Expect/Continue retain their required flush.
func (h *HTTP1Settings) HeaderBodyMergeLimit(limit int64) *HTTP1Settings {
	h.overrides = append(h.overrides, func(c *profiles.HTTP1Config) { c.HeaderBodyMergeLimit(limit) })
	return h
}

// Set returns to the client builder.
func (h *HTTP1Settings) Set() *Builder { return h.builder }

func (b *Builder) applyHTTP1Settings() error {
	if b.configureHTTP1 == nil && b.http1settings == nil && b.cli.http1Config == nil {
		return nil
	}

	config := profiles.HTTP1Config{}
	if b.configureHTTP1 != nil {
		b.configureHTTP1(&config)
	}

	if b.http1settings != nil {
		for _, override := range b.http1settings.overrides {
			override(&config)
		}
	}

	if config.WriteBufferBytes < 0 || config.WriteBufferBytes > 1<<20 ||
		config.BodyReadBytes < 0 || config.BodyReadBytes > 1<<20 {
		return fmt.Errorf("HTTP/1 buffer sizes must be between zero and 1 MiB")
	}

	if limit := config.HeaderBodyMergeBytes; limit != nil {
		if *limit < 0 || *limit > 1<<20 {
			return fmt.Errorf("HTTP/1 header/body merge limit must be between zero and 1 MiB")
		}

		value := *limit
		config.HeaderBodyMergeBytes = &value
	}

	configureHTTP1Transports(b.cli.cli.Transport, config)
	b.cli.http1Config = &config

	return nil
}

func configureHTTP1Transports(transport http.RoundTripper, config profiles.HTTP1Config) {
	switch tr := transport.(type) {
	case *clientHintsTransport:
		configureHTTP1Transports(tr.Base, config)
	case *h3transport.Transport:
		configureHTTP1Transports(tr.Fallback, config)
	case *roundtripper:
		for _, tcp := range tr.networkTransports() {
			configureHTTP1Transports(tcp, config)
		}
	case *http.Transport:
		tr.WriteBufferSize = config.WriteBufferBytes
		if config.BodyReadBytes == 0 && config.HeaderBodyMergeBytes == nil {
			tr.HTTP1WritePolicy = nil
			return
		}

		tr.HTTP1WritePolicy = func(info http.HTTP1WriteInfo) http.HTTP1WriteOptions {
			options := http.HTTP1WriteOptions{BodyReadSize: config.BodyReadBytes}
			if limit := config.HeaderBodyMergeBytes; limit != nil && info.ContentLength != 0 {
				options.FlushHeaders = !info.InMemory || info.Chunked || info.ContentLength < 0 ||
					info.HeaderBytes > *limit || info.ContentLength > *limit-info.HeaderBytes
			}

			return options
		}
	}
}
