ARG DOCKERHUB_MIRROR=
ARG GOLANG_VERSION
ARG ALPINE_VERSION
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION
ARG MOCKERY_VERSION

FROM --platform=$BUILDPLATFORM ${DOCKERHUB_MIRROR}cosmtrek/air:${AIR_VERSION} AS air
FROM --platform=$BUILDPLATFORM ${DOCKERHUB_MIRROR}golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint
FROM --platform=$BUILDPLATFORM ${DOCKERHUB_MIRROR}vektra/mockery:${MOCKERY_VERSION} AS mockery

FROM --platform=$BUILDPLATFORM ${DOCKERHUB_MIRROR}golang:${GOLANG_VERSION} AS toolchain

RUN apk add --no-cache ca-certificates git

FROM toolchain AS base

ARG TARGETARCH
ARG TARGETOS=linux

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./go.sum ./

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

FROM ${DOCKERHUB_MIRROR}alpine:${ALPINE_VERSION} AS runtime-utils

RUN apk add --no-cache util-linux setpriv ca-certificates libgcc

FROM runtime-utils AS runtime-rootfs

RUN <<EOF
set -eu
set -- /etc/ssl/certs/ca-certificates.crt /usr/bin/nsenter /bin/setpriv \
    /usr/lib/libcap-ng.so.* /usr/lib/libeconf.so.* /usr/lib/libgcc_s.so.1 /lib/ld-musl-*.so.1

for file; do
    mkdir -p "/rootfs${file%/*}"
    cp -a "$file" "/rootfs$file"
done
cp -L /etc/os-release /etc/alpine-release /rootfs/etc/

apk info -q --who-owns "$@" >/tmp/owners
mkdir -p /rootfs/lib/apk/db
awk 'NR == FNR { wanted["P:" $0]; next }
     { for (i = 1; i <= NF; i++) if ($i in wanted) { printf "%s\n\n", $0; next } }' \
    /tmp/owners RS= FS='\n' /lib/apk/db/installed >/rootfs/lib/apk/db/installed
test "$(grep -c '^P:' /rootfs/lib/apk/db/installed)" -eq "$(sort -u /tmp/owners | wc -l)"
EOF

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

COPY --from=runtime-rootfs /rootfs/ /

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/agent/agent /bin/agent

ENTRYPOINT ["/bin/agent"]
