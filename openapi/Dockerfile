FROM golang:1.26.8-alpine3.24 AS base

ARG GOPROXY
ENV GOPROXY=${GOPROXY}

RUN apk add --no-cache openssl nodejs npm openjdk11-jre git

RUN npm install -g @openapitools/openapi-generator-cli

RUN npm install -g @redocly/cli@2.31.5

RUN npm install -g @stoplight/prism-cli@4.6.1

RUN npm install -g prettier@3.9.9

RUN CGO_ENABLED=0 go install github.com/air-verse/air@v1.62 && \
    CGO_ENABLED=0 go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0

FROM base AS development

ARG GOPROXY

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/openapi

COPY ./openapi/go.mod ./

RUN go mod download

COPY ./openapi/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
