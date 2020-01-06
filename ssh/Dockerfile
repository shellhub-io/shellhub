# base stage
FROM golang:alpine AS base

RUN apk add --update git ca-certificates openssh-client && \
    go get -u github.com/Masterminds/glide

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/ssh

ADD ./glide.yaml .
ADD ./glide.lock .

RUN glide install

# builder stage
FROM base AS builder

ADD . .

RUN glide -q install --skip-test && go build

# development stage
FROM base AS development

RUN apk add --update openssl
RUN go get github.com/markbates/refresh

RUN glide -q install --skip-test
RUN cp -a vendor /vendor

ADD entrypoint-dev.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.7 AS production

RUN apk add --update openssh-client

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/ssh/ssh /ssh

ENTRYPOINT /ssh
