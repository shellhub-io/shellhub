ARG GOLANG_VERSION=1.26.9
ARG ALPINE_VERSION=3.24

FROM --platform=$BUILDPLATFORM golang:${GOLANG_VERSION:-1.26}-alpine${ALPINE_VERSION:-3.24} AS base

ARG GOPROXY
ARG TARGETARCH
ARG TARGETOS=linux

RUN apk add --no-cache ca-certificates git

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

COPY ./agent/go.mod ./agent/go.sum ./

RUN go mod download

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

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GOARM=${TARGETVARIANT#v} go build -tags docker -ldflags "-s -w -X main.AgentVersion=${SHELLHUB_VERSION}" -o agent

FROM alpine:${ALPINE_VERSION:-3.24} AS runtime-utils

RUN apk add --no-cache util-linux setpriv ca-certificates

FROM base AS development

ARG GOPROXY
ENV GOPROXY=${GOPROXY}

RUN apk add --update openssl openssh-client util-linux setpriv
RUN CGO_ENABLED=0 go install github.com/air-verse/air@v1.62 && \
    CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0 && \
    CGO_ENABLED=0 go install github.com/vektra/mockery/v3/...@v3.8.0

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