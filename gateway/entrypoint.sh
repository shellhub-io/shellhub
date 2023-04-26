#!/bin/sh

# Same tunning settings applied by NGINX Ingress Controller
# https://github.com/kubernetes/ingress-nginx/blob/844a02c276788e293480c080fe09f4d242545c82/internal/ingress/controller/nginx.go#L512

WORKER_PROCESSES="${WORKER_PROCESSES:-auto}"
MAX_WORKER_OPEN_FILES="${MAX_WORKER_OPEN_FILES:-0}"
MAX_WORKER_CONNECTIONS="${MAX_WORKER_CONNECTIONS:-16384}"

if [ "$WORKER_PROCESSES" == "auto" ]; then
    WORKER_PROCESSES=$(nproc)
fi

if [ "$MAX_WORKER_OPEN_FILES" == "0" ]; then
    MAX_WORKER_OPEN_FILES=$(($(ulimit -Sn) - 1024))

    if [ "$MAX_WORKER_OPEN_FILES" -lt "1024" ]; then
	MAX_WORKER_OPEN_FILES=1024
    fi
fi

if [ "$MAX_WORKER_CONNECTIONS" == "0" ]; then
    MAX_WORKER_CONNECTIONS=$(($MAX_WORKER_OPEN_FILES * 3 / 4))
fi

export WORKER_PROCESSES
export MAX_WORKER_OPEN_FILES
export MAX_WORKER_CONNECTIONS

export HOST_IP=$(ip -4 route show default | awk '{ print $3 }')

generate() {
	gomplate -f /app/nginx.conf -o /usr/local/openresty/nginx/conf/nginx.conf
	gomplate -f /app/conf.d/shellhub.conf -o /etc/nginx/conf.d/shellhub.conf
}

watch () {
  while inotifywait -q -r -e close_write "/app/nginx.conf" "/app/conf.d/" > /dev/null;
  do
	  generate
	  nginx -s reload
  done
}

generate
watch &

exec "$@"
