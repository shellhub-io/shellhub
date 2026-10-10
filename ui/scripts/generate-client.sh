#!/bin/sh
set -e

if [ -z "$OPENAPI_SPEC_PATH" ]; then
  npx "@redocly/cli@${REDOCLY_VERSION:?REDOCLY_VERSION is unset, the ui service takes it from versions.env}" bundle /openapi/spec/openapi.yaml -o /tmp/openapi.json --force
  OPENAPI_SPEC_PATH=/tmp/openapi.json
  export OPENAPI_SPEC_PATH
fi

cd "$(dirname "$0")/../apps/console"
exec npx openapi-ts
