#!/bin/sh

# Generates a JSON config file from SHELLHUB_* environment variables.
# The output is served by nginx as the /v2/ui/config endpoint.
#
# Usage: gen-config.sh <output-file>

OUTPUT="${1:?usage: gen-config.sh OUTPUT_FILE}"

cat > "$OUTPUT" <<EOF
{
  "version": "${SHELLHUB_VERSION:-}",
  "enterprise": ${SHELLHUB_ENTERPRISE:-false},
  "cloud": ${SHELLHUB_CLOUD:-false}
}
EOF
