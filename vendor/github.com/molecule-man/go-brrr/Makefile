.PHONY: init
init:
	git submodule update --init
	go install golang.org/x/perf/cmd/benchstat@latest
	go install github.com/campoy/embedmd@latest
	$(MAKE) lib/libbrotli_cref.a

lib/libbrotli_cref.a: testdata/build_libbrotli_cref.sh \
	$(wildcard brotli-ref/c/common/*.c) \
	$(wildcard brotli-ref/c/dec/*.c) \
	$(wildcard brotli-ref/c/enc/*.c) \
	$(wildcard brotli-ref/c/common/*.h) \
	$(wildcard brotli-ref/c/dec/*.h) \
	$(wildcard brotli-ref/c/enc/*.h) \
	$(wildcard brotli-ref/c/include/brotli/*.h)
	testdata/build_libbrotli_cref.sh

# Keep the s390x C library separate from the host library.
S390X_LIB ?= /tmp/go-brrr-s390x/lib
S390X_GOCACHE ?= /tmp/go-brrr-s390x/gocache
.PHONY: test-s390x
test-s390x:
	mkdir -p $(S390X_LIB) $(S390X_GOCACHE)
	docker build -q -t go-brrr-s390x - < scripts/s390x.Dockerfile >/dev/null
	docker run --rm \
		-v $(CURDIR):/src -v $(S390X_LIB):/src/lib \
		-v $(shell go env GOROOT):/goroot:ro \
		-v $(shell go env GOMODCACHE):/gomodcache -v $(S390X_GOCACHE):/gocache \
		-e GOROOT=/goroot -e GOMODCACHE=/gomodcache -e GOCACHE=/gocache -e GOFLAGS=-buildvcs=false \
		-e GOTOOLCHAIN=local -e TAGS='$(TAGS)' -e HOME=/tmp \
		-u $(shell id -u):$(shell id -g) -w /src go-brrr-s390x \
		sh -c 'PATH=/goroot/bin:$$PATH scripts/test-s390x.sh'
