#!/bin/sh

cp -a /node_modules .
npm install

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

# Generate config.json so Vite can serve it during development.
mkdir -p apps/console/public
"$SCRIPTS_DIR/gen-config.sh" apps/console/public/config.json

# Generate the OpenAPI client from the combined spec, then keep regenerating
# it whenever any spec file changes so Vite HMR picks up the new types
# without needing to recreate the container.
npm run generate -w @shellhub/console
# chokidar-cli shells out to $SHELL, which isn't set in this alpine image.
SHELL=/bin/sh npx -y chokidar-cli@3.0.0 '/openapi/spec/**/*.yaml' --debounce 500 \
  -c 'npm run generate -w @shellhub/console' &

npm run dev:console
