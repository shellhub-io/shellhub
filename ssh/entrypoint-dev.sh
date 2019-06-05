#!/bin/sh

# This script generates a new SSH private key for development environment

mkdir -p /var/run/secrets

if [ ! -f /var/run/secrets/ssh_private_key ]; then
    echo "Generating private key"
    openssl genrsa -out /var/run/secrets/ssh_private_key 2048
fi

refresh run
