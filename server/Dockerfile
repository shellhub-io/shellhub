# base stage
FROM golang:1.23.7-alpine3.21 AS base

ARG GOPROXY

RUN apk add --update git ca-certificates build-base openssh-client

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/server

COPY ./server/go.mod ./server/go.sum ./

RUN go mod download

RUN apk add curl

# builder stage
FROM base AS builder

ARG GOPROXY

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./server .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/server

RUN go build -tags internal_api

# development stage
FROM base AS development

ARG GOPROXY
ENV GOPROXY ${GOPROXY}

RUN apk add --update openssl
RUN go install github.com/air-verse/air@v1.52.3 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.23.0 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.63.2 && \
    go install github.com/vektra/mockery/v2/...@v2.20.0

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./server/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/server

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.21.3 AS production

RUN apk add curl

RUN apk add --update openssh-client

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/server/server /server

ENTRYPOINT /server
