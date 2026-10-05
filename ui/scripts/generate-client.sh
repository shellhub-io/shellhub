#!/bin/sh
set -e

if [ -z "$OPENAPI_SPEC_PATH" ]; then
  npx @redocly/cli@2.31.5 bundle /openapi/spec/openapi.yaml -o /tmp/openapi.json --force
  OPENAPI_SPEC_PATH=/tmp/openapi.json
  export OPENAPI_SPEC_PATH
fi

cd "$(dirname "$0")/../apps/console"
exec npx openapi-ts
