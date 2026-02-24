#!/bin/sh

# This script generates a new SSH key pair for development environment

mkdir -p /var/run/secrets

if [ ! -f /var/run/secrets/api_private_key ]; then
    echo "Generating private key"
    openssl genpkey -algorithm RSA -out /var/run/secrets/api_private_key -pkeyopt rsa_keygen_bits:2048
    openssl rsa -in /var/run/secrets/api_private_key -pubout -out /var/run/secrets/api_public_key
fi

ln -sf $PWD/api /api

# If the cloud repo is mounted at the expected container path, run air
# with -tags enterprise (EE). Otherwise run a plain CE build.
CLOUD_DIR="/go/src/github.com/shellhub-io/cloud"
WORKSPACE="/go/src/github.com/shellhub-io"

if [ -d "$CLOUD_DIR" ]; then
    echo "Cloud sources found at $CLOUD_DIR — building api-enterprise (EE)"

    # Compile email templates from MJML source into /templates.
    # NOTE: Templates are compiled once at container startup.
    # Restart the container to recompile after editing .mjml files.
    if [ -d "$CLOUD_DIR/templates" ]; then
        echo "Compiling email templates from $CLOUD_DIR/templates"
        mjml "$CLOUD_DIR"/templates/*.mjml -o /templates || {
            echo "ERROR: MJML template compilation failed" >&2
            exit 1
        }
        echo "Email templates compiled successfully."
    fi

    # Create go.work so the unified build can resolve both shellhub and cloud modules.
    go work init \
        "$WORKSPACE/shellhub" \
        "$WORKSPACE/shellhub/openapi" \
        "$WORKSPACE/shellhub/api" \
        "$WORKSPACE/cloud"

    exec air -c .air.enterprise.toml
else
    exec air
fi
