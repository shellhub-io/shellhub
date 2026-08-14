# syntax=docker/dockerfile:1
#
# Build args:
#   EDITION=community (default) | enterprise
#   MJML_VERSION=4.10.1 (default)
#
# CE:  docker build --build-arg EDITION=community -f server/Dockerfile .
# EE:  docker build --build-arg EDITION=enterprise \
#        --build-context cloud-src=../cloud \
#        -f server/Dockerfile .
#
# The main build context is always the shellhub repo root.
# Cloud source is provided separately via --build-context (never in main context).

ARG EDITION=community
ARG MJML_VERSION=4.10.1

# ── cloud-src: empty by default, overridden via --build-context for EE ────────
FROM scratch AS cloud-src

# ── base stage ────────────────────────────────────────────────────────────────
FROM golang:1.25.13-alpine3.24 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates curl build-base openssh-client

WORKDIR /go/src/github.com/shellhub-io/shellhub
COPY ./go.mod ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./openapi/go.mod ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/server
COPY ./server/go.mod ./server/go.sum ./

RUN go mod download

# ── builder stage ─────────────────────────────────────────────────────────────
FROM base AS builder

ARG EDITION=community
ARG MJML_VERSION

RUN if [ "$EDITION" = "enterprise" ]; then \
        apk add --no-cache npm && npm install -g mjml@${MJML_VERSION}; \
    fi

COPY ./pkg     /go/src/github.com/shellhub-io/shellhub/pkg
COPY ./openapi /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./server     /go/src/github.com/shellhub-io/shellhub/server

WORKDIR /go/src/github.com/shellhub-io/shellhub/server

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

RUN --mount=type=bind,from=cloud-src,target=/go/src/github.com/shellhub-io/cloud \
    if [ "$EDITION" = "enterprise" ]; then \
        cd /go/src && go work init \
            ./github.com/shellhub-io/shellhub \
            ./github.com/shellhub-io/shellhub/openapi \
            ./github.com/shellhub-io/shellhub/server \
            ./github.com/shellhub-io/cloud && \
        GOWORK=/go/src/go.work go build -tags enterprise -o /server github.com/shellhub-io/cloud/cmd/server && \
        mjml /go/src/github.com/shellhub-io/cloud/templates/*.mjml -o /templates; \
    else \
        go build -o /server .; \
    fi

# ── development stage ─────────────────────────────────────────────────────────
FROM base AS development

ARG GOPROXY
ARG MJML_VERSION
ENV GOPROXY=${GOPROXY}

RUN apk add --update openssl docker-cli npm
RUN npm install -g mjml@${MJML_VERSION}
RUN CGO_ENABLED=0 go install github.com/air-verse/air@v1.62 && \
    CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.3 && \
    CGO_ENABLED=0 go install github.com/vektra/mockery/v3/...@v3.7.1

COPY ./server/entrypoint-dev.sh /entrypoint.sh

WORKDIR /go/src/github.com/shellhub-io/shellhub/server

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/entrypoint.sh"]

# ── production stage ──────────────────────────────────────────────────────────
FROM alpine:3.24.1 AS production

RUN apk add --no-cache curl openssh-client

COPY --from=builder /server /server
COPY --from=builder /templates /templates

ENTRYPOINT ["/server", "server"]
