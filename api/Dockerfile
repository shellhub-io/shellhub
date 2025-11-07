# base stage
FROM golang:1.24-alpine3.22 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/openapi

COPY ./openapi/go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

COPY ./api/go.mod ./api/go.sum ./

RUN go mod download

RUN apk add curl

# builder stage
FROM base AS builder

ARG GOPROXY

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./openapi $GOPATH/src/github.com/shellhub-io/shellhub/openapi
COPY ./api .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

RUN go build

# development stage
FROM base AS development

ARG GOPROXY
ENV GOPROXY ${GOPROXY}

RUN apk add --update openssl build-base docker-cli
RUN go install github.com/air-verse/air@v1.62 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.25 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 && \
    go install github.com/vektra/mockery/v2/...@v2.53.2

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./api/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

RUN mkdir -p /templates

COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.22.2 AS production

RUN apk add curl

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/api/api /api

RUN mkdir -p /templates

COPY ./install.sh /templates/install.sh

ENTRYPOINT ["/api", "server"]
