FROM golang:alpine AS builder
WORKDIR /go/src/github.com/wzshiming/socks5/
COPY . .
ENV CGO_ENABLED=0
RUN go install ./cmd/socks5

FROM alpine
EXPOSE 1080
COPY --from=builder /go/bin/socks5 /usr/local/bin/
ENTRYPOINT [ "/usr/local/bin/socks5" ]
