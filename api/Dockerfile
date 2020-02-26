# base stage
FROM golang:alpine AS base

RUN apk add --update git ca-certificates && \
    go get -u github.com/Masterminds/glide

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./api/glide.yaml ./api/glide.lock ./

RUN glide install --skip-test

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

# builder stage
FROM base AS builder

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./api .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN glide install --skip-test

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

RUN go build

# development stage
FROM base AS development

RUN apk add --update openssl
RUN go get github.com/markbates/refresh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN glide install --skip-test
RUN cp -a $GOPATH/src/github.com/shellhub-io/shellhub/vendor /vendor

COPY ./api/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/api

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.11 AS production

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/api/api /api

ENTRYPOINT /api
