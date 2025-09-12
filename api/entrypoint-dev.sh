#!/bin/sh

# This script generates a new SSH key pair for development environment

mkdir -p /var/run/secrets

if [ ! -f /var/run/secrets/api_private_key ]; then
    echo "Generating private key"
    openssl genpkey -algorithm RSA -out /var/run/secrets/api_private_key -pkeyopt rsa_keygen_bits:2048
    openssl rsa -in /var/run/secrets/api_private_key -pubout -out /var/run/secrets/api_public_key
fi

ln -sf $PWD/api /api

air
