# syntax=docker/dockerfile:1

ARG EDITION=community
ARG MJML_VERSION=4.10.1

FROM scratch AS cloud-src

FROM golang:1.26.9-alpine3.24 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates curl build-base openssh-client

WORKDIR /go/src/github.com/shellhub-io/shellhub
COPY ./go.mod ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./openapi/go.mod ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/server
COPY ./server/go.mod ./server/go.sum ./

RUN go mod download

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

ARG GO_LDFLAGS

RUN --mount=type=bind,from=cloud-src,target=/go/src/github.com/shellhub-io/cloud \
    if [ "$EDITION" = "enterprise" ]; then \
        cd /go/src && go work init \
            ./github.com/shellhub-io/shellhub \
            ./github.com/shellhub-io/shellhub/openapi \
            ./github.com/shellhub-io/shellhub/server \
            ./github.com/shellhub-io/cloud && \
        GOWORK=/go/src/go.work go build -tags enterprise -ldflags "$GO_LDFLAGS" -o /server github.com/shellhub-io/cloud/cmd/server && \
        mjml /go/src/github.com/shellhub-io/cloud/templates/*.mjml -o /templates; \
    else \
        go build -ldflags "$GO_LDFLAGS" -o /server .; \
    fi

FROM base AS development

ARG GOPROXY
ARG MJML_VERSION
ENV GOPROXY=${GOPROXY}

RUN apk add --update openssl npm
COPY --from=docker:29.8.1-cli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker:29.8.1-cli /usr/local/libexec/docker/cli-plugins/docker-buildx /usr/local/libexec/docker/cli-plugins/docker-buildx
RUN npm install -g mjml@${MJML_VERSION}
RUN CGO_ENABLED=0 go install github.com/air-verse/air@v1.62 && \
    CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 && \
    CGO_ENABLED=0 go install github.com/vektra/mockery/v3/...@v3.8.0

COPY ./server/entrypoint-dev.sh /entrypoint.sh

WORKDIR /go/src/github.com/shellhub-io/shellhub/server

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/entrypoint.sh"]

FROM alpine:3.24.2 AS production

RUN apk add --no-cache curl openssh-client && \
    apk upgrade --no-cache libssl3 libcrypto3 # CVE-2026-14456: drop upgrade when alpine base includes the fix

COPY --from=builder /server /server
COPY --from=builder /templates /templates

ENTRYPOINT ["/server", "server"]
