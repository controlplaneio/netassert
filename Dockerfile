FROM golang:1.24-alpine AS builder

ARG VERSION

COPY . /build
WORKDIR /build

RUN go mod download && \
    CGO_ENABLED=0 GO111MODULE=on go build -ldflags="-X 'main.appName=NetAssert' -X 'main.version=${VERSION}'" -v -o /netassertv2 cmd/netassert/cli/*.go && \
    ls -ltr /netassertv2

FROM gcr.io/distroless/base:nonroot
COPY --from=builder /netassertv2 /usr/bin/netassertv2

ENTRYPOINT [ "/usr/bin/netassertv2" ]
