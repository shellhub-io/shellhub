#!/bin/sh

PREFIX=$1

echo "window.env = {"

env | while read line; do
    VAR=$(echo $line | cut -d'=' -f1)
    VALUE=$(echo $line | cut -d'=' -f2-)

    # Ignore variables that don't starts with 'prefix'
    echo $VAR | grep -q "^$PREFIX" || continue

    echo -e "\tVUE_APP_${VAR}: \"${VALUE}\","
done

echo "};"
