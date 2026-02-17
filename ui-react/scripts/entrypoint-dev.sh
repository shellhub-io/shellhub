#!/bin/sh

cp -a /node_modules .
npm install

PREFIX=SHELLHUB

IFS=$'\n'
for line in $(env); do
    VAR=$(echo $line | cut -d'=' -f1)
    VALUE=$(echo $line | cut -d'=' -f2-)

    # Ignore variables that don't starts with 'prefix'
    echo $VAR | grep -q "^$PREFIX" || continue

    export "VITE_SHELLHUB_${VAR}=${VALUE}"
done

npm run dev:admin &
npm run dev:docs &
npm run dev:blog &
npm run dev:website
