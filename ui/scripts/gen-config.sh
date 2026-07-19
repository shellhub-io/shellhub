#!/bin/sh

# Generates a JSON config file from SHELLHUB_* environment variables.
# The output is served by nginx as the /config.json endpoint.
#
# Usage: gen-config.sh <output-file>

OUTPUT="${1:?usage: gen-config.sh OUTPUT_FILE}"

EDITION=$(printf '%s' "${SHELLHUB_EDITION:-community}" | tr '[:upper:]' '[:lower:]' | tr -d '[:space:]')

case "$EDITION" in
    community|enterprise|cloud) ;;
    *)
        echo "🚫 ERROR: invalid SHELLHUB_EDITION '$EDITION': must be community, enterprise, or cloud." >&2
        exit 1
        ;;
esac

cat > "$OUTPUT" <<EOF
{
  "version": "${SHELLHUB_VERSION:-}",
  "edition": "${EDITION}",
  "onboardingUrl": "${SHELLHUB_ONBOARDING_URL:-}",
  "announcements": ${SHELLHUB_ANNOUNCEMENTS:-false},
  "webEndpoints": ${SHELLHUB_WEB_ENDPOINTS:-false},
  "stripePublishableKey": "${STRIPE_PUBLISHABLE_KEY:-}",
  "chatwootWebsiteToken": "${SHELLHUB_CHATWOOT_WEBSITE_TOKEN:-}",
  "chatwootBaseUrl": "${SHELLHUB_CHATWOOT_BASEURL:-}"
}
EOF
