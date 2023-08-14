# base stage
FROM golang:1.20.4-alpine3.16 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

COPY ./api/go.mod ./api/go.sum ./

RUN go mod download

RUN apk add curl

# builder stage
FROM base AS builder

ARG GOPROXY

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
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
RUN go install github.com/markbates/refresh@v1.11.1 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.53.3 && \
    go install github.com/vektra/mockery/v2/...@v2.20.0

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./api/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.18.3 AS production

RUN apk add curl

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/api/api /api

RUN mkdir /templates

COPY ./api/templates /templates

ENTRYPOINT /api server
