FROM openresty/openresty:1.15.8.2-6-alpine

RUN ["rm", "/etc/nginx/conf.d/default.conf"]

COPY nginx.conf /usr/local/openresty/nginx/conf/nginx.conf
COPY shellhub.conf /etc/nginx/conf.d/
