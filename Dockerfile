FROM golang@sha256:3889b425f035be855a72fb4755265311293b6d414521f0a519d819df32222d83 AS builder # golang:1.26.6-alpine3.24

ARG VERSION
ARG SCANNER_IMG_VERSION
ARG SNIFFER_IMG_VERSION

COPY . /build
WORKDIR /build

RUN go mod download && \
    CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-X 'main.appName=NetAssert' -X 'main.version=${VERSION}' -X 'main.scannerImgVersion=${SCANNER_IMG_VERSION}' -X 'main.snifferImgVersion=${SNIFFER_IMG_VERSION}'" -v -o /netassertv2 cmd/netassert/cli/*.go && \
    ls -ltr /netassertv2

FROM gcr.io/distroless/static:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3 # gcr.io/distroless/base:nonroot
COPY --from=builder /netassertv2 /usr/bin/netassertv2

ENTRYPOINT [ "/usr/bin/netassertv2" ]
