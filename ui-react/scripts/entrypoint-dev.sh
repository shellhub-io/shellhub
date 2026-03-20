#!/bin/sh

cp -a /node_modules .
npm install

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

# Generate config.json so Vite can serve it during development.
mkdir -p apps/console/public
"$SCRIPTS_DIR/gen-config.sh" apps/console/public/config.json

# Generate OpenAPI client from the combined spec (all editions)
echo "Bundling OpenAPI spec..."
npx @redocly/cli@1.0.0-beta.100 bundle /openapi/spec/openapi.yaml -o /tmp/openapi.json --force
echo "Generating OpenAPI client..."
OPENAPI_SPEC_PATH=/tmp/openapi.json npx -w @shellhub/console openapi-ts

npm run dev:console
