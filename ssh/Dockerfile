# base stage
FROM golang:1.23-alpine3.19 AS base

ARG GOPROXY

RUN apk add --update git ca-certificates build-base openssh-client

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

COPY ./ssh/go.mod ./ssh/go.sum ./

RUN go mod download

RUN apk add curl

# builder stage
FROM base AS builder

ARG GOPROXY

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./ssh .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

RUN go build -tags internal_api

# development stage
FROM base AS development

ARG GOPROXY
ENV GOPROXY ${GOPROXY}

RUN apk add --update openssl
RUN go install github.com/air-verse/air@v1.52.3 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.23.0 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 && \
    go install github.com/vektra/mockery/v2/...@v2.20.0

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./ssh/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.22.1 AS production

RUN apk add curl

RUN apk add --update openssh-client

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/ssh/ssh /ssh

ENTRYPOINT /ssh
