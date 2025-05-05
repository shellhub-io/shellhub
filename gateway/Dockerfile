# base stage
FROM golang:1.23.6-alpine3.20 AS base

ARG GOPROXY

RUN apk add --no-cache git ca-certificates libgcc curl certbot certbot-nginx certbot-dns certbot-dns-digitalocean

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

COPY ./gateway/go.mod ./gateway/go.sum ./

RUN go mod download

# builder stage
FROM base AS builder

COPY ./pkg $GOPATH/src/github.com/shellhub-io/shellhub/pkg
COPY ./gateway .

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

RUN go build

# development stage
FROM base AS development

COPY --from=openresty/openresty:1.25.3.1-5-alpine-apk /usr/local/openresty /usr/local/openresty

RUN mkdir /etc/shellhub-gateway

RUN mkdir -p /var/run/openresty /etc/letsencrypt && \
    curl -sSL https://ssl-config.mozilla.org/ffdhe2048.txt -o /etc/shellhub-gateway/dhparam.pem

RUN apk add --update openssl build-base
RUN go install github.com/air-verse/air@v1.52.3 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.23.0 && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@v2.1.6 && \
    go install github.com/vektra/mockery/v2/...@v2.20.0

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

RUN go mod download

COPY ./gateway/entrypoint-dev.sh /entrypoint.sh

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.21.3 AS production

RUN apk add libgcc curl certbot certbot-nginx certbot-dns certbot-dns-digitalocean

COPY --from=openresty/openresty:1.25.3.1-5-alpine-apk /usr/local/openresty /usr/local/openresty

RUN mkdir /etc/shellhub-gateway

RUN mkdir -p /var/run/openresty /etc/letsencrypt && \
    curl -sSL https://ssl-config.mozilla.org/ffdhe2048.txt -o /etc/shellhub-gateway/dhparam.pem

COPY --from=builder /go/src/github.com/shellhub-io/shellhub/gateway/gateway /gateway

RUN mkdir /templates

COPY ./gateway/nginx /templates

ENTRYPOINT /gateway
