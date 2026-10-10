ARG GOLANG_VERSION
ARG AIR_VERSION
ARG GOLANGCI_LINT_VERSION
ARG REDOCLY_VERSION
ARG PRISM_VERSION
ARG PRETTIER_VERSION

FROM cosmtrek/air:${AIR_VERSION} AS air
FROM golangci/golangci-lint:${GOLANGCI_LINT_VERSION}-alpine AS golangci-lint

FROM golang:${GOLANG_VERSION} AS base

ARG REDOCLY_VERSION
ARG PRISM_VERSION
ARG PRETTIER_VERSION

RUN apk add --no-cache openssl nodejs npm openjdk11-jre git

RUN npm install -g @openapitools/openapi-generator-cli

RUN npm install -g @redocly/cli@${REDOCLY_VERSION:?}

RUN npm install -g @stoplight/prism-cli@${PRISM_VERSION:?}

RUN npm install -g prettier@${PRETTIER_VERSION:?}

COPY --from=air /go/bin/air /usr/local/bin/air
COPY --from=golangci-lint /usr/bin/golangci-lint /usr/local/bin/golangci-lint

FROM base AS development

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub

COPY ./go.mod ./

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/openapi

COPY ./openapi/go.mod ./

RUN go mod download

COPY ./openapi/entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
