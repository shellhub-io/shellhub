FROM nginx:1.25.3-alpine AS base

RUN ["rm", "/etc/nginx/conf.d/default.conf"]

COPY --from=hairyhenderson/gomplate:v2.5.0-slim /gomplate /bin/gomplate

RUN apk add inotify-tools
RUN apk add certbot certbot-nginx
RUN apk add curl

RUN mkdir -p /etc/nginx/default.d

RUN mkdir /app

VOLUME /etc/letsencrypt

WORKDIR /app

COPY ./gateway/entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["/usr/sbin/nginx", "-g", "daemon off;"]

FROM base as development 

FROM base as production

RUN apk add curl

COPY ./gateway/nginx.conf /app/

COPY ./gateway/conf.d /app/conf.d
