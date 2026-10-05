#!/bin/sh

cp -a /node_modules .
npm install

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

mkdir -p apps/console/public
"$SCRIPTS_DIR/gen-config.sh" apps/console/public/config.json

npm run generate -w @shellhub/console
SHELL=/bin/sh npx -y chokidar-cli@3.0.0 '/openapi/spec/**/*.yaml' --debounce 500 \
  -c 'npm run generate -w @shellhub/console' &

npm run dev -w @shellhub/website &
rm -f apps/docs/.astro/dev.json
npm run dev -w @shellhub/docs &

npm run dev:console
