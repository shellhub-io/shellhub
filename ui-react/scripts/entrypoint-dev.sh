#!/bin/sh

cp -a /node_modules .
npm install

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

# Generate config.json so Vite can serve it during development.
mkdir -p apps/console/public
"$SCRIPTS_DIR/gen-config.sh" apps/console/public/config.json

npm run dev:console
