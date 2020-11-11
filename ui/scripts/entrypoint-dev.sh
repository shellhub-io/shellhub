#!/bin/sh

cp -a /node_modules .

PREFIX=SHELLHUB

IFS=$'\n'
for line in $(env); do
    VAR=$(echo $line | cut -d'=' -f1)
    VALUE=$(echo $line | cut -d'=' -f2-)

    # Ignore variables that don't starts with 'prefix'
    echo $VAR | grep -q "^$PREFIX" || continue

    export "VUE_APP_${VAR}=${VALUE}"
done

yarn serve
