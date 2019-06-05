# base stage
FROM golang:alpine AS base

RUN apk add --update git ca-certificates && \
    go get -u github.com/Masterminds/glide

WORKDIR $GOPATH/src/github.com/gustavosbarreto/devicehub/ssh

ADD ./glide.yaml .
ADD ./glide.lock .

RUN glide install

# builder stage
FROM base AS builder

ADD . .

RUN glide -q install --skip-test && go build -o ssh && cp ssh /ssh

# development stage
FROM base AS development

RUN apk add --update openssl

COPY --from=builder /ssh /src/ssh

RUN go get github.com/markbates/refresh

WORKDIR $GOPATH/src/github.com/gustavosbarreto/devicehub/ssh

RUN glide -q install --skip-test

ADD entrypoint-dev.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.7 AS production

WORKDIR /app
COPY --from=builder /ssh /app/

ENTRYPOINT ./ssh
