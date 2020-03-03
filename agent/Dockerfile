# base stage
FROM golang:1.13-alpine3.11 AS base

RUN apk add --update git ca-certificates build-base bash util-linux && \
    go get -u github.com/Masterminds/glide

RUN ln -sf /bin/bash /bin/sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./agent/glide.yaml ./agent/glide.lock ./

RUN glide install

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

# builder stage
FROM base AS builder

ARG SHELLHUB_VERSION=latest

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./agent .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN glide install --skip-test

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

RUN go build -tags docker -ldflags "-X main.AgentVersion=${SHELLHUB_VERSION}"

# development stage
FROM base AS development

RUN apk add --update openssl openssh-client
RUN go get github.com/markbates/refresh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN glide install --skip-test
RUN cp -a $GOPATH/src/github.com/shellhub-io/shellhub/vendor /vendor

COPY ./agent/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/agent

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.11 AS production

WORKDIR /app
COPY --from=builder /agent /app/

ENTRYPOINT ./agent
