# base stage
FROM golang:1.17.6-alpine3.14 AS base

# target build flags
ARG DB_NAME
ARG GOPROXY
ARG BILLING
ARG CLOUD
ARG CACHE
ARG ENTERPRISE

RUN apk add --no-cache git ca-certificates

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

COPY ./api/go.mod ./api/go.sum ./

RUN go mod download

# builder stage
FROM base AS builder

ARG DB_NAME
ARG GOPROXY
ARG BILLING
ARG CLOUD
ARG CACHE
ARG ENTERPRISE

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./api .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

RUN go build

# development stage
FROM base AS development

# override envs in case of build args
ARG GOPROXY
ENV GOPROXY ${GOPROXY}

ARG DB_NAME
ENV MONGO_DB_NAME ${DB_NAME}

ARG ENTERPRISE
ENV SHELLHUB_ENTERPRISE ${ENTERPRISE}

ARG CLOUD
ENV SHELLHUB_CLOUD ${CLOUD}

ARG BILLING
ENV SHELLHUB_BILLING ${BILLING}

ARG CACHE
ENV SHELLHUB_STORE_CACHE ${CACHE}

# install deps

RUN apk add --update openssl build-base docker-cli
RUN go install github.com/markbates/refresh@v1.11.1 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1 && \
    go install github.com/vektra/mockery/v2/...@v2.9.2

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./api/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.15.0 AS production

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/api/api /api

ENTRYPOINT /api server
