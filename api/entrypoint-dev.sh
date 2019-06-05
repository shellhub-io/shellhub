#!/bin/sh

# This script generates a new SSH key pair for development environment

mv /vendor .

mkdir -p /var/run/secrets

if [ ! -f /var/run/secrets/api_private_key ]; then
    echo "Generating private key"
    openssl genrsa -out /var/run/secrets/api_private_key 2048
    openssl rsa -in /var/run/secrets/api_private_key -pubout -out /var/run/secrets/api_public_key
fi

refresh run
