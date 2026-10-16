# syntax=docker/dockerfile:1

ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION

FROM cosmtrek/air:${AIR_VERSION} AS air
FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint

FROM golang:${GOLANG_VERSION} AS base

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./go.sum ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

COPY ./gateway/go.mod ./gateway/go.sum ./

RUN --mount=type=cache,id=shellhub-go-mod,target=/go/pkg/mod,sharing=shared \
    go mod download

FROM base AS builder

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./gateway .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN --mount=type=cache,id=shellhub-go-mod,target=/go/pkg/mod,sharing=shared \
    go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

RUN --mount=type=cache,id=shellhub-go-mod,target=/go/pkg/mod,sharing=shared \
    --mount=type=cache,id=shellhub-go-build,target=/root/.cache/go-build,sharing=shared \
    go build

FROM golang:${GOLANG_VERSION} AS tools

RUN apk add --no-cache git ca-certificates libgcc

COPY --from=air /go/bin/air /usr/local/bin/air
COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint

FROM tools AS development

COPY --from=base $GOPATH/src/github.com/shellhub-io/shellhub $GOPATH/src/github.com/shellhub-io/shellhub

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./gateway/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

RUN go mod download

ENTRYPOINT ["/entrypoint.sh"]

FROM alpine:${ALPINE_VERSION} AS production

RUN apk add --no-cache ca-certificates && \
    apk upgrade --no-cache libssl3 libcrypto3 # CVE-2026-14456: drop upgrade when alpine base includes the fix

VOLUME /data

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/gateway/gateway /gateway

ENTRYPOINT ["/gateway"]
