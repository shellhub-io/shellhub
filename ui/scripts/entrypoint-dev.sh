#!/bin/sh

# Install dependencies only when package.json or package-lock.json changes
STAMP="./node_modules/.install_hash"
HASH=$(md5sum ./package.json ./package-lock.json 2>/dev/null | md5sum | cut -d' ' -f1)
if [ ! -d ./node_modules/.bin ] || [ "$(cat "$STAMP" 2>/dev/null)" != "$HASH" ]; then
    npm install
    echo "$HASH" > "$STAMP"
fi

PREFIX=SHELLHUB

IFS=$'\n'
for line in $(env); do
    VAR=$(echo $line | cut -d'=' -f1)
    VALUE=$(echo $line | cut -d'=' -f2-)

    # Ignore variables that don't starts with 'prefix'
    echo $VAR | grep -q "^$PREFIX" || continue

    export "VUE_APP_${VAR}=${VALUE}"
done

yarn dev
