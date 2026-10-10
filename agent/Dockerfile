ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION
ARG MOCKERY_VERSION

FROM --platform=$BUILDPLATFORM cosmtrek/air:${AIR_VERSION} AS air
FROM --platform=$BUILDPLATFORM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM --platform=$BUILDPLATFORM vektra/mockery:${MOCKERY_VERSION} AS mockery

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION} AS toolchain

RUN apk add --no-cache ca-certificates git

FROM toolchain AS base

ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

COPY ./agent/go.mod ./agent/go.sum ./

RUN go mod download

FROM base AS builder

ARG SHELLHUB_VERSION=latest
ARG TARGETARCH
ARG TARGETVARIANT
ARG TARGETOS=linux

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./agent .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} go build -tags docker -ldflags "-s -w -X main.AgentVersion=${SHELLHUB_VERSION}" -o agent

FROM alpine:${ALPINE_VERSION} AS runtime-utils

RUN apk add --no-cache util-linux setpriv ca-certificates

FROM toolchain AS tools

RUN apk add --no-cache openssl openssh-client util-linux setpriv
COPY --from=air /go/bin/air /usr/local/bin/air
COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint
COPY --from=mockery /usr/local/bin/mockery /usr/local/bin/mockery

FROM tools AS development

COPY --from=base $GOPATH/pkg/mod $GOPATH/pkg/mod
COPY --from=base $GOPATH/src/github.com/shellhub-io/shellhub $GOPATH/src/github.com/shellhub-io/shellhub

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

ARG USERNAME
ARG PASSWORD

RUN [ -z "$USERNAME" ] || { adduser -D "$USERNAME" && echo "$USERNAME:$PASSWORD" | chpasswd; }

COPY ./agent/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

ENTRYPOINT ["/entrypoint.sh"]

FROM scratch

ARG TARGETARCH

COPY --from=runtime-utils /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=runtime-utils /usr/bin/nsenter /usr/bin/
COPY --from=runtime-utils /bin/setpriv /bin/

COPY --from=runtime-utils /usr/lib/libcap-ng.so.* /usr/lib/
COPY --from=runtime-utils /usr/lib/libeconf.so.* /usr/lib/

COPY --from=runtime-utils /lib/ld-musl-*.so.1 /lib/

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/agent/agent /bin/agent

ENTRYPOINT ["/bin/agent"]
