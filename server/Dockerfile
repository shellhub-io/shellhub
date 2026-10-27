# syntax=docker/dockerfile:1

ARG DOCKERHUB_MIRROR=
ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION
ARG MOCKERY_VERSION
ARG DOCKER_VERSION
ARG MJML_VERSION
ARG EDITION=community

FROM scratch AS cloud-src

FROM ${DOCKERHUB_MIRROR}cosmtrek/air:${AIR_VERSION} AS air
FROM ${DOCKERHUB_MIRROR}golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM ${DOCKERHUB_MIRROR}vektra/mockery:${MOCKERY_VERSION} AS mockery
FROM ${DOCKERHUB_MIRROR}docker:${DOCKER_VERSION}-cli AS docker-cli

FROM ${DOCKERHUB_MIRROR}golang:${GOLANG_VERSION} AS base

WORKDIR /go/src/github.com/shellhub-io/shellhub
COPY ./go.mod ./go.sum ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./openapi/go.mod ./openapi/go.sum ./

WORKDIR /go/src/github.com/shellhub-io/shellhub/server
COPY ./server/go.mod ./server/go.sum ./

RUN --mount=type=cache,id=shellhub-go-mod,target=/go/pkg/mod,sharing=shared \
    go mod download

FROM base AS builder

ARG EDITION=community
ARG MJML_VERSION

RUN --mount=type=cache,id=shellhub-npm,target=/root/.npm,sharing=shared \
    if [ "$EDITION" = "enterprise" ]; then \
        apk add --no-cache npm && npm install -g mjml@${MJML_VERSION:?}; \
    fi

COPY ./pkg     /go/src/github.com/shellhub-io/shellhub/pkg
COPY ./openapi /go/src/github.com/shellhub-io/shellhub/openapi
COPY ./server     /go/src/github.com/shellhub-io/shellhub/server

WORKDIR /go/src/github.com/shellhub-io/shellhub/server

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

ARG GO_LDFLAGS

RUN --mount=type=bind,from=cloud-src,target=/go/src/github.com/shellhub-io/cloud \
    --mount=type=cache,id=shellhub-go-mod,target=/go/pkg/mod,sharing=shared \
    --mount=type=cache,id=shellhub-go-build,target=/root/.cache/go-build,sharing=shared \
    if [ "$EDITION" = "enterprise" ]; then \
        cd /go/src && go work init \
            ./github.com/shellhub-io/shellhub \
            ./github.com/shellhub-io/shellhub/openapi \
            ./github.com/shellhub-io/shellhub/server \
            ./github.com/shellhub-io/cloud && \
        GOWORK=/go/src/go.work go build -tags enterprise -ldflags "$GO_LDFLAGS" -o /server github.com/shellhub-io/cloud/cmd/server && \
        mjml /go/src/github.com/shellhub-io/cloud/templates/*.mjml --config.allowIncludes true -o /templates; \
    else \
        go build -ldflags "$GO_LDFLAGS" -o /server .; \
    fi

FROM ${DOCKERHUB_MIRROR}golang:${GOLANG_VERSION} AS tools

ARG MJML_VERSION

RUN apk add --no-cache git ca-certificates curl build-base openssh-client openssl npm
COPY --from=docker-cli /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker-cli /usr/local/libexec/docker/cli-plugins/docker-buildx /usr/local/libexec/docker/cli-plugins/docker-buildx
COPY --from=air /go/bin/air /usr/local/bin/air
COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=mockery /usr/local/bin/mockery /usr/local/bin/mockery
RUN npm install -g mjml@${MJML_VERSION:?}

FROM tools AS development

COPY --from=base /go/src/github.com/shellhub-io/shellhub /go/src/github.com/shellhub-io/shellhub

COPY ./server/entrypoint-dev.sh /entrypoint.sh

WORKDIR /go/src/github.com/shellhub-io/shellhub/server

RUN go mod download

RUN mkdir -p /templates
COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/entrypoint.sh"]

FROM ${DOCKERHUB_MIRROR}alpine:${ALPINE_VERSION} AS production

RUN apk add --no-cache curl && \
    apk upgrade --no-cache libssl3 libcrypto3 # CVE-2026-14456: drop upgrade when alpine base includes the fix

COPY --from=builder /server /server
COPY --from=builder /templates /templates

ENTRYPOINT ["/server", "server"]
