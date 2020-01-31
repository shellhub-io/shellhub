# base stage
FROM golang:alpine AS base

RUN apk add --update git ca-certificates openssh-client && \
    go get -u github.com/Masterminds/glide

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./ssh/glide.yaml ./ssh/glide.lock ./

RUN glide install

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

# builder stage
FROM base AS builder

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./ssh .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN glide -q install --skip-test

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

RUN go build

# development stage
FROM base AS development

RUN apk add --update openssl
RUN go get github.com/markbates/refresh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN glide install --skip-test
RUN cp -a $GOPATH/src/github.com/shellhub-io/shellhub/vendor /vendor

COPY ./ssh/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.11 AS production

RUN apk add --update openssh-client

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/ssh/ssh /ssh

ENTRYPOINT /ssh
