FROM openresty/openresty:1.15.8.2-6-alpine
RUN apk add --no-cache --update perl curl && opm install ledgetech/lua-resty-http
RUN ["rm", "/etc/nginx/conf.d/default.conf"]

COPY --from=hairyhenderson/gomplate:v2.5.0-slim /gomplate /bin/gomplate

COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY shellhub.conf /etc/nginx/conf.d/
COPY entrypoint.sh /

ENTRYPOINT ["/entrypoint.sh"]

CMD ["/usr/local/openresty/bin/openresty", "-g", "daemon off;"]
