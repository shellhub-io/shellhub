# base stage
FROM golang:1.17.4-alpine3.14 AS base

RUN apk add --no-cache git ca-certificates

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./
COPY ./api/go.mod ./api/go.mod

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/cli

COPY ./cli/go.mod ./cli/go.sum ./

RUN go mod download

# builder stage
FROM base AS builder

COPY ./api $GOPATH/src/github.com/shellhub-io/shellhub/api
COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./cli .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/cli

RUN go build

# development stage
FROM base AS development

RUN apk add --update openssl build-base docker-cli
RUN go install github.com/markbates/refresh@v1.11.1 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.42.1

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./cli/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/cli

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.15.0 AS production

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/cli/cli /cli

ENTRYPOINT /cli
