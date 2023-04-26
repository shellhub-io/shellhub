FROM openresty/openresty:1.21.4.1-alpine AS base

RUN ["rm", "/etc/nginx/conf.d/default.conf"]

COPY --from=hairyhenderson/gomplate:v2.5.0-slim /gomplate /bin/gomplate

RUN apk add --no-cache perl curl && opm get bungle/lua-resty-template
RUN apk add inotify-tools
RUN apk add certbot certbot-nginx

RUN mkdir -p /etc/nginx/default.d

RUN mkdir /app

WORKDIR /app

COPY ./gateway/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]

FROM base as development 

FROM base as production

COPY ./gateway/nginx.conf /app/

COPY ./gateway/conf.d /app/conf.d
