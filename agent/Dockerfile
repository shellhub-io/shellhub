# base stage
FROM golang:1.16.4-alpine3.13 AS base

ARG GOPROXY

RUN apk add --update git ca-certificates build-base bash util-linux setpriv

RUN ln -sf /bin/bash /bin/sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

COPY ./agent/go.mod ./agent/go.sum ./

RUN go mod download

# builder stage
FROM base AS builder

ARG SHELLHUB_VERSION=latest
ARG GOPROXY

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./agent .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

RUN go build -tags docker -ldflags "-X main.AgentVersion=${SHELLHUB_VERSION}"

# development stage
FROM base AS development

ARG GOPROXY
ENV GOPROXY ${GOPROXY}

RUN apk add --update openssl openssh-client
RUN go get github.com/markbates/refresh && go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.37.1

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

#RUN cp -a $GOPATH/src/github.com/shellhub-io/shellhub/vendor /vendor

COPY ./agent/entrypoint-dev.sh /entrypoint.sh
COPY ./.revive.toml /

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.11 AS production

WORKDIR /app
COPY --from=builder /agent /app/

ENTRYPOINT ./agent
