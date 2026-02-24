# syntax=docker/dockerfile:1
#
# Build args:
#   EDITION=community (default) | enterprise
#
# CE:  docker build --build-arg EDITION=community -f api/Dockerfile .
# EE:  docker build --build-arg EDITION=enterprise \
#        --build-context cloud-src=../cloud \
#        -f api/Dockerfile .
#
# The main build context is always the shellhub repo root.
# Cloud source is provided separately via --build-context (never in main context).

ARG EDITION=community

# ── cloud-src: empty by default, overridden via --build-context for EE ────────
FROM scratch AS cloud-src

# ── base stage ────────────────────────────────────────────────────────────────
FROM golang:1.24.13-alpine3.22 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates curl

WORKDIR /go/src/github.com/shellhub-io/shellhub
COPY ./go.mod ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./openapi/go.mod ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/api
COPY ./api/go.mod ./api/go.sum ./

RUN go mod download

# ── builder stage ─────────────────────────────────────────────────────────────
FROM base AS builder

ARG EDITION=community

COPY ./pkg     /go/src/github.com/shellhub-io/shellhub/pkg
COPY ./openapi /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./api     /go/src/github.com/shellhub-io/shellhub/api

WORKDIR /go/src/github.com/shellhub-io/shellhub/api

RUN --mount=type=bind,from=cloud-src,target=/go/src/github.com/shellhub-io/cloud \
    if [ "$EDITION" = "enterprise" ]; then \
        cd /go/src && go work init \
            ./github.com/shellhub-io/shellhub \
            ./github.com/shellhub-io/shellhub/openapi \
            ./github.com/shellhub-io/shellhub/api \
            ./github.com/shellhub-io/cloud && \
        cd /go/src/github.com/shellhub-io/shellhub/api && \
        GOWORK=/go/src/go.work go build -tags enterprise -o /api .; \
    else \
        go build -o /api .; \
    fi

# ── development stage ─────────────────────────────────────────────────────────
FROM base AS development

ARG GOPROXY
ENV GOPROXY=${GOPROXY}

RUN apk add --update openssl build-base docker-cli
RUN go install github.com/air-verse/air@v1.62 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 && \
    go install github.com/vektra/mockery/v2/...@v2.53.2

COPY ./api/entrypoint-dev.sh /entrypoint.sh

WORKDIR /go/src/github.com/shellhub-io/shellhub/api

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/entrypoint.sh"]

# ── production stage ──────────────────────────────────────────────────────────
FROM alpine:3.23.3 AS production

RUN apk add --no-cache curl

COPY --from=builder /api /api

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/api", "server"]
