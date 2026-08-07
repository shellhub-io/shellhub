#!/bin/sh

# Generates a JSON config file from SHELLHUB_* environment variables.
# The output is served as the /config.json endpoint.
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

# Booleans are rendered unquoted, so a value that is not true/false lands in
# config.json as raw JSON — enough to append a second "edition" key that wins
# at parse time and overrides the check above.
require_bool() {
    case "$2" in
        true|false) ;;
        *)
            echo "🚫 ERROR: invalid $1 '$2': must be true or false." >&2
            exit 1
            ;;
    esac
}

require_bool SHELLHUB_ANNOUNCEMENTS "${SHELLHUB_ANNOUNCEMENTS:-false}"
require_bool SHELLHUB_WEB_ENDPOINTS "${SHELLHUB_WEB_ENDPOINTS:-false}"

# Escape a value for use inside a JSON string literal. Values arrive verbatim
# from the deployment's environment, and a lone quote or backslash would make
# config.json unparseable, taking the whole console down. JSON also forbids
# raw control characters: tab/newline/return become escapes, and the rest are
# dropped since they have no meaning in a URL, token or version string.
#
# The $!{N;ba} guard matters: the more common :a;N;$!ba slurp makes sed print
# and exit on single-line input, skipping every substitution below it.
json_string() {
    printf '%s' "$1" |
        tr -d '\000-\010\013\014\016-\037' |
        sed \
            -e ':a' -e '$!{N;ba' -e '}' \
            -e 's/\\/\\\\/g' \
            -e 's/"/\\"/g' \
            -e "s/$(printf '\t')/\\\\t/g" \
            -e "s/$(printf '\r')/\\\\r/g" \
            -e 's/\n/\\n/g'
}

cat > "$OUTPUT" <<EOF
{
  "version": "$(json_string "${SHELLHUB_VERSION:-}")",
  "edition": "${EDITION}",
  "onboardingUrl": "$(json_string "${SHELLHUB_ONBOARDING_URL:-}")",
  "announcements": ${SHELLHUB_ANNOUNCEMENTS:-false},
  "webEndpoints": ${SHELLHUB_WEB_ENDPOINTS:-false},
  "stripePublishableKey": "$(json_string "${SHELLHUB_STRIPE_PUBLISHABLE_KEY:-}")",
  "chatwootWebsiteToken": "$(json_string "${SHELLHUB_CHATWOOT_WEBSITE_TOKEN:-}")",
  "chatwootBaseUrl": "$(json_string "${SHELLHUB_CHATWOOT_BASEURL:-}")"
}
EOF
