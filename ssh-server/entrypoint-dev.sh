#!/bin/sh

# This script generates a new SSH private key for development environment

mkdir -p /var/run/secrets

if [ ! -f /var/run/secrets/ssh_server_private_key ]; then
    echo "Generating private key"
    openssl genrsa -out /var/run/secrets/ssh_server_private_key 2048
fi

export SSH_SERVER_PRIV_KEY_PATH=/var/run/secrets/ssh_server_private_key

refresh run
