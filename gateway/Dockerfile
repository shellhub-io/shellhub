# base stage
FROM golang:1.25.13-alpine3.24 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates libgcc

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

COPY ./gateway/go.mod ./gateway/go.sum ./

RUN go mod download

# builder stage
FROM base AS builder

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./gateway .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

RUN go build

# development stage
FROM base AS development

RUN CGO_ENABLED=0 go install github.com/air-verse/air@v1.62 && \
    CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.3

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./gateway/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.24.1 AS production

RUN apk add --no-cache ca-certificates libgcc && \
    apk upgrade --no-cache libssl3 libcrypto3 # CVE-2026-14456: drop upgrade when alpine base includes the fix

# The account key and the certificates. Declared here rather than in the compose
# file so the volume exists however the image is run: without it, every recreate
# of the container asks the CA for everything again, and Let's Encrypt caps how
# many times a week it will answer.
VOLUME /data

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/gateway/gateway /gateway

ENTRYPOINT ["/gateway"]
