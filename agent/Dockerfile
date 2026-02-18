# Unified Dockerfile for ShellHub Agent
# Default build: production multiarch
# Development: use --target development

ARG GOLANG_VERSION=1.24
ARG ALPINE_VERSION=3.22

# Base stage for both development and production
FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION:-1.24}-alpine${ALPINE_VERSION:-3.22} AS base

ARG GOPROXY
ARG TARGETARCH
ARG TARGETOS=linux

# Install Go BUILD architecture dependencies
RUN apk add --no-cache ca-certificates

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

COPY ./agent/go.mod ./agent/go.sum ./

RUN go mod download

# Builder stage for production
FROM base AS builder

ARG SHELLHUB_VERSION=latest
ARG GOPROXY
ARG TARGETARCH
ARG TARGETVARIANT
ARG TARGETOS=linux

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./agent .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

# Cross-compile for target architecture
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} go build -tags docker -ldflags "-s -w -X main.AgentVersion=${SHELLHUB_VERSION}" -o agent

# Runtime utilities stage - CRITICAL: must use target platform
FROM --platform=$TARGETPLATFORM alpine:${ALPINE_VERSION:-3.22} AS runtime-utils

# Install runtime binaries for the TARGET architecture
RUN apk add --no-cache util-linux setpriv ca-certificates

# development stage
FROM base AS development

ARG GOPROXY
ENV GOPROXY ${GOPROXY}

RUN apk add --update openssl openssh-client util-linux setpriv
RUN go install github.com/air-verse/air@v1.62 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

#RUN cp -a $GOPATH/src/github.com/shellhub-io/shellhub/vendor /vendor

COPY ./agent/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

ENTRYPOINT ["/entrypoint.sh"]

# DEFAULT STAGE: Production multiarch build (scratch-based)
FROM scratch

ARG TARGETARCH

# Copy CA certificates from runtime-utils
COPY --from=runtime-utils /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy required utilities from runtime-utils - NOT from builder
COPY --from=runtime-utils /usr/bin/nsenter /usr/bin/
COPY --from=runtime-utils /bin/setpriv /bin/

# Copy shared libraries from runtime-utils
COPY --from=runtime-utils /usr/lib/libcap-ng.so.* /usr/lib/

# Copy musl loader from runtime-utils
COPY --from=runtime-utils /lib/ld-musl-*.so.1 /lib/

# Copy the agent binary
COPY --from=builder /go/src/github.com/shellhub-io/shellhub/agent/agent /bin/agent

ENTRYPOINT ["/bin/agent"]