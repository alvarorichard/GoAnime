.DEFAULT_GOAL := generate

export YTDLP_VERSION := 2026.10.26

license:
	curl -sL https://liam.sh/-/gh/g/license-header.sh | bash -s

clean:
	rm -f ./cmd/export-ytdlp/export-${YTDLP_VERSION}.json

fetch:
	cd ./cmd/codegen && go mod tidy
	cd ./cmd/gen-jsonschema && go mod tidy
	go mod tidy

up:
	cd ./cmd/codegen && go get -u -t ./... && go mod tidy
	cd ./cmd/gen-jsonschema && go get -u -t ./... && go mod tidy
	cd ./_examples && go get -u -t ./... && go mod tidy
	go get -u -t ./... && go mod tidy

commit: generate
	git add --all \
		Makefile \
		cmd/export-ytdlp \
		*.gen.go *.gen_test.go \
		optiondata/*.gen.go
	git commit -m "chore(codegen): generate updated cli bindings"

patch:
	uv -q run \
		--python 3.13 \
		--no-project \
		--with "yt-dlp==${YTDLP_VERSION}" \
		--script ./cmd/export-ytdlp/export.py \
		${YTDLP_VERSION} > ./cmd/export-ytdlp/export-${YTDLP_VERSION}.json

test: fetch
	GORACE='exitcode=1 halt_on_error=1' go test -v -race -timeout 5m -count 3 ./...

test-e2e:
	go test -tags=e2e -run '^TestExtractedInfoE2E' -v -timeout 15m ./...

generate: license fetch patch
	rm -rf \
		*.gen.go *.gen_test.go \
		optiondata/*.gen.go
	cd ./cmd/codegen && go run . ../export-ytdlp/export-${YTDLP_VERSION}.json ../../
	gofmt -e -s -w .
	cd ./cmd/gen-jsonschema && go run . ../../optiondata/
	go vet .
	go test -v ./...
