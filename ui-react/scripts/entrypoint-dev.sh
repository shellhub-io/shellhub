#!/bin/sh

cp -a /node_modules .
npm install

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

# Generate config.json so Vite can serve it during development.
mkdir -p apps/console/public
"$SCRIPTS_DIR/gen-config.sh" apps/console/public/config.json

# Generate OpenAPI client from cloud spec (always cloud, regardless of dev mode)
echo "Bundling OpenAPI cloud spec..."
npx @redocly/cli@1.0.0-beta.100 bundle /openapi/spec/cloud-openapi.yaml -o /tmp/openapi.json --force
echo "Generating OpenAPI client..."
OPENAPI_SPEC_PATH=/tmp/openapi.json npx -w @shellhub/console openapi-ts

npm run dev:console
