#!/bin/sh

# This script generates the key pairs used by the development environment

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

# air rebuilds to /tmp/air/main on every change. Symlinking it at /server keeps
# the binary at the same path in development and production, which is what
# bin/cli execs to reach the admin commands.
ln -sf /tmp/air/main /server

# go work init fails when the file already exists, so drop any left from an
# earlier run before choosing a build below.
rm -f go.work go.work.sum

# If the cloud repo is mounted at the expected container path, run air
# with -tags enterprise (EE). Otherwise run a plain CE build.
CLOUD_DIR="/go/src/github.com/shellhub-io/cloud"
WORKSPACE="/go/src/github.com/shellhub-io"

if [ -d "$CLOUD_DIR" ]; then
    echo "Cloud sources found at $CLOUD_DIR — building server-enterprise (EE)"

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
        "$WORKSPACE/shellhub/server" \
        "$WORKSPACE/cloud"

    # Use the cloud entry point, which blank-imports cloud extension packages
    # before calling the community server.
    exec air -build.cmd "go build -tags enterprise -o /tmp/air/main github.com/shellhub-io/cloud/cmd/server"
else
    exec air
fi
