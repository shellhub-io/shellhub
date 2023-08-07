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

wait_for_acme_webserver() {
  for i in `seq 30` ; do
    nc -z localhost 80 > /dev/null 2>&1

    if [ $? -eq 0 ] ; then
        return
    fi

    sleep 1
  done

  echo "Timed out waiting for ACME webserver" >&2

  exit 1
}

# The certificate generation is only available in production mode and if the SHELLHUB_AUTO_SSL is set to true.
if [ "$SHELLHUB_ENV" != "development" ] && [ "$SHELLHUB_AUTO_SSL" == "true" ]; then
    if [ -z "$SHELLHUB_DOMAIN" ]; then
        echo "SHELLHUB_DOMAIN cannot be empty"
        exit 1
    fi

    if ! echo "$SHELLHUB_DOMAIN" | grep -qE '^([a-z0-9]+(-[a-z0-9]+)*\.)+[a-z]{2,}$'; then
        echo "SHELLHUB_DOMAIN must be a valid domain name"
        exit 1
    fi

    # If the certificate is not generated yet, generate it.
    if [ ! -f /etc/letsencrypt/live/$SHELLHUB_DOMAIN/fullchain.pem ]; then
        echo "Generating SSL certificate"

        ACME_WEBSERVER_ROOT="/var/www/letsencrypt"
        ACME_CHALLENGE_DIR="$ACME_WEBSERVER_ROOT/.well-known/acme-challenge"

        mkdir -p $ACME_CHALLENGE_DIR

        # We need to ensure that acme challenge webserver is running before running certbot,
        # as we are utilizing the webroot mode, which relies on a running local webserver
        ACME_WEBSERVER_PID=$(cd $ACME_WEBSERVER_ROOT; python -m http.server 80 > /dev/null 2>&1 & echo $!)
        wait_for_acme_webserver

        certbot certonly --non-interactive --agree-tos --register-unsafely-without-email --webroot --webroot-path $ACME_WEBSERVER_ROOT --preferred-challenges http -n -d $SHELLHUB_DOMAIN
        if [ $? -ne 0 ]; then
            echo "Failed to generate SSL certificate"
            exit 1
        fi

        echo "SSL certificate successfully generated"

        curl https://ssl-config.mozilla.org/ffdhe2048.txt > /etc/letsencrypt/live/$SHELLHUB_DOMAIN/dhparam.pem 2> /dev/null
        if [ $? -ne 0 ]; then
            echo "Failed to download Mozilla's DH parameters"
            exit 1
        fi

        echo "Mozilla's DH parameters successfully downloaded"

        kill $ACME_WEBSERVER_PID
    fi

    # Loop every 24 hours to check if certificate is about to expire.
    # It is safe to run the renew each 24 hours because the command will only renew if the certificate is about to expire.
    # About to expire is defined as: "if a certificate is going to expire in less than 30 days, it will be renewed."
    # https://eff-certbot.readthedocs.io/en/stable/using.html#renewing-certificates
    while sleep "24h" ; do
        # If certificate is already generated, check if it is about to expire. If so, renew it.
        # You may want to renew near the 60 day mark, to provide enough time for any possible problems that may arise.
        # https://letsencrypt.org/docs/faq/#what-is-the-lifetime-for-let-s-encrypt-certificates-for-how-long-are-they-valid

        echo "Checking if SSL certificate needs to be renewed"
        certbot renew
        if [ $? -ne 0 ]; then
            echo "Failed to renew SSL certificate"
            exit 1
        fi

        nginx -s reload

        echo "SSL certificate successfully renewed"
    done &
fi

generate() {
    gomplate -f /app/nginx.conf -o /usr/local/openresty/nginx/conf/nginx.conf
    gomplate -f /app/conf.d/shellhub.conf -o /etc/nginx/conf.d/shellhub.conf
}

if [ "$SHELLHUB_ENV" == "development" ]; then
    while inotifywait -q -r -e close_write "/app/nginx.conf" "/app/conf.d/" > /dev/null; do
        generate
        nginx -s reload
    done &
fi

generate

echo "Starting NGINX"
exec "$@"
