# golang:1.26-alpine
FROM golang@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS builder

ARG VERSION
ARG SCANNER_IMG_VERSION
ARG SNIFFER_IMG_VERSION

COPY . /build
WORKDIR /build

RUN go mod download && \
    CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-X 'main.appName=NetAssert' -X 'main.version=${VERSION}' -X 'main.scannerImgVersion=${SCANNER_IMG_VERSION}' -X 'main.snifferImgVersion=${SNIFFER_IMG_VERSION}'" -v -o /netassertv2 cmd/netassert/cli/*.go && \
    ls -ltr /netassertv2

# gcr.io/distroless/base:nonroot
FROM gcr.io/distroless/base@sha256:97b9d04bed1c754b756c3c4b6a04915c22fb0b5d96a59944eb3bf78c26e6e157
COPY --from=builder /netassertv2 /usr/bin/netassertv2

ENTRYPOINT [ "/usr/bin/netassertv2" ]
