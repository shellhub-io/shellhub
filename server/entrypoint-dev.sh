#!/bin/sh

mkdir -p /var/run/secrets

if [ ! -f /var/run/secrets/api_private_key ]; then
    echo "Generating API private key"
    openssl genpkey -algorithm RSA -out /var/run/secrets/api_private_key -pkeyopt rsa_keygen_bits:2048
    openssl rsa -in /var/run/secrets/api_private_key -pubout -out /var/run/secrets/api_public_key
fi

if [ ! -f /var/run/secrets/ssh_private_key ]; then
    echo "Generating SSH host key"
    openssl genpkey -algorithm RSA -out /var/run/secrets/ssh_private_key -pkeyopt rsa_keygen_bits:2048
fi

ln -sf /tmp/air/main /server

rm -f go.work go.work.sum

CLOUD_DIR="/go/src/github.com/shellhub-io/cloud"
WORKSPACE="/go/src/github.com/shellhub-io"

if [ -d "$CLOUD_DIR" ]; then
    echo "Cloud sources found at $CLOUD_DIR — building server-enterprise (EE)"

    if [ -d "$CLOUD_DIR/templates" ]; then
        echo "Compiling email templates from $CLOUD_DIR/templates"
        mjml "$CLOUD_DIR"/templates/*.mjml -o /templates || {
            echo "ERROR: MJML template compilation failed" >&2
            exit 1
        }
        echo "Email templates compiled successfully."
    fi

    go work init \
        "$WORKSPACE/shellhub" \
        "$WORKSPACE/shellhub/openapi" \
        "$WORKSPACE/shellhub/server" \
        "$WORKSPACE/cloud"

    exec air -build.cmd "go build -tags enterprise -o /tmp/air/main github.com/shellhub-io/cloud/cmd/server"
else
    exec air
fi
