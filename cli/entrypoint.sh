#!/bin/sh

gomplate -f /etc/nginx/conf.d/shellhub.conf -o /etc/nginx/conf.d/shellhub.conf

exec "$@"
