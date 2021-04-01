#!/bin/sh

# Same tunning settings applied by NGINX Ingress Controller
# https://github.com/kubernetes/ingress-nginx/blob/844a02c276788e293480c080fe09f4d242545c82/internal/ingress/controller/nginx.go#L512

if [ "$WORKER_PROCESSES" == "" ]; then
    WORKER_PROCESSES=$(nproc)
fi

if [ "$MAX_WORKER_OPEN_FILES" == "" ]; then
    MAX_OPEN_FILES=$(ulimit -n)
    MAX_WORKER_OPEN_FILES=$((($MAX_OPEN_FILES / $WORKER_PROCESSES) - 1024))

    if [ "$MAX_WORKER_OPEN_FILES" -lt "1024" ]; then
	MAX_WORKER_OPEN_FILES=1024
    fi
fi

if [ "$MAX_WORKER_CONNECTIONS" == "" ]; then
    MAX_WORKER_CONNECTIONS=$(($MAX_WORKER_OPEN_FILES * 3 / 4))
fi

export WORKER_PROCESSES
export MAX_WORKER_OPEN_FILES
export MAX_WORKER_CONNECTIONS

gomplate -f /usr/local/openresty/nginx/conf/nginx.conf -o /usr/local/openresty/nginx/conf/nginx.conf
gomplate -f /etc/nginx/conf.d/shellhub.conf -o /etc/nginx/conf.d/shellhub.conf

exec "$@"
