FROM openresty/openresty:1.15.8.2-6-alpine as base

RUN ["rm", "/etc/nginx/conf.d/default.conf"]

COPY --from=hairyhenderson/gomplate:v2.5.0-slim /gomplate /bin/gomplate

RUN apk add --no-cache perl curl && opm get bungle/lua-resty-template

COPY gateway/nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY gateway/shellhub.conf /etc/nginx/conf.d/
COPY gateway/kickstart.sh /usr/local/openresty/nginx/html/
COPY gateway/entrypoint.sh /

FROM base as test

RUN apk add --no-cache alpine-sdk

COPY --from=golang:1.16.4-alpine3.13 /usr/local/go /usr/local/go

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"

WORKDIR $GOPATH/src/github.com/shellhub-io/shellhub/gateway/tests

COPY ./gateway/tests/go.mod ./gateway/tests/go.sum ./

RUN go mod download

COPY ./gateway/tests $GOPATH/src/github.com/shellhub-io/shellhub/gateway/tests

CMD ["go", "test"]

FROM base as production

ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]
