# base stage
FROM golang:alpine AS base

RUN apk add --update git ca-certificates && \
    go get -u github.com/Masterminds/glide

WORKDIR $GOPATH/src/github.com/gustavosbarreto/devicehub/ssh-server

ADD ./glide.yaml .
ADD ./glide.lock .

RUN glide install

# builder stage
FROM base AS builder

ADD . .

RUN glide -q install --skip-test && go build -o ssh-server && cp ssh-server /ssh-server

# development stage
FROM base AS development

RUN apk add --update openssl

COPY --from=builder /ssh-server /src/ssh-server

RUN go get github.com/markbates/refresh

WORKDIR $GOPATH/src/github.com/gustavosbarreto/devicehub/ssh-server

RUN glide -q install --skip-test

ADD entrypoint-dev.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]

# production stage
FROM alpine:3.7 AS production

WORKDIR /app
COPY --from=builder /ssh-server /app/

ENTRYPOINT ./ssh-server
