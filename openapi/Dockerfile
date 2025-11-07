FROM golang:1.25-alpine3.22 AS base

ARG GOPROXY
ENV GOPROXY ${GOPROXY}

RUN apk add --no-cache openssl build-base nodejs npm openjdk11-jre git

RUN npm install -g @openapitools/openapi-generator-cli

RUN npm install -g @redocly/cli@1.0.0-beta.100

RUN npm install -g @stoplight/prism-cli@4.6.1

RUN npm install -g prettier@2.8.7

RUN go install github.com/air-verse/air@v1.62 && \
    go install github.com/go-delve/delve/cmd/dlv@v1.25 && \
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 && \
    go install github.com/vektra/mockery/v2/...@v2.53.2

FROM base AS development

ARG GOPROXY

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

# As OpenAPI directory doens't depend on other modules in the repo, we don't need to copy go.sum.
# COPY ./go.sum ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/openapi

COPY ./openapi/go.mod ./

# As OpenAPI directory doens't depend on other modules in the repo, we don't need to copy go.sum.
# COPY ./openapi/go.sum ./

RUN go mod download

COPY ./openapi/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
