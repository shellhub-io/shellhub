#!/bin/sh
set -e

# Bundle the combined OpenAPI spec (all editions) and generate the typed
# client. The Dockerfile builder stage sets OPENAPI_SPEC_PATH to a spec
# it bundled in an earlier stage; skip the bundle step in that case.
if [ -z "$OPENAPI_SPEC_PATH" ]; then
  npx @redocly/cli@1.0.0-beta.100 bundle /openapi/spec/openapi.yaml -o /tmp/openapi.json --force
  OPENAPI_SPEC_PATH=/tmp/openapi.json
  export OPENAPI_SPEC_PATH
fi

cd "$(dirname "$0")/../apps/console"
exec npx openapi-ts
