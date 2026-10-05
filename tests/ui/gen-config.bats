#!/usr/bin/env bats

REPO_ROOT="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)"
GEN_CONFIG="$REPO_ROOT/ui/scripts/gen-config.sh"

gen_config() {
    local out="$BATS_TEST_TMPDIR/config.json"
    env -i PATH="$PATH" "$@" sh "$GEN_CONFIG" "$out" || return
    cat "$out"
}

require_jq() {
    command -v jq >/dev/null || skip "jq not installed"
}

@test "stripePublishableKey comes from SHELLHUB_STRIPE_PUBLISHABLE_KEY" {
    out=$(gen_config SHELLHUB_STRIPE_PUBLISHABLE_KEY=pk_test_123)
    [[ "$out" == *'"stripePublishableKey": "pk_test_123"'* ]]
}

@test "every value reads a SHELLHUB_-prefixed variable" {
    # shellcheck disable=SC2016 # matching a literal $, not expanding it
    unprefixed=$(grep -oE '\$\{?[A-Z_][A-Z0-9_]*' "$GEN_CONFIG" |
        sed -E 's/^\$\{?//' |
        grep -vE '^(SHELLHUB_[A-Z0-9_]*|EDITION|OUTPUT)$' |
        sort -u)

    if [ -n "$unprefixed" ]; then
        echo "unprefixed variables in gen-config.sh: $unprefixed" >&2
        return 1
    fi
}

@test "an unset value renders an empty string" {
    out=$(gen_config)
    [[ "$out" == *'"stripePublishableKey": ""'* ]]
}

@test "edition defaults to community" {
    out=$(gen_config)
    [[ "$out" == *'"edition": "community"'* ]]
}

@test "edition is lowercased and stripped of whitespace" {
    out=$(gen_config 'SHELLHUB_EDITION= Cloud ')
    [[ "$out" == *'"edition": "cloud"'* ]]
}

@test "an invalid edition aborts without writing the config" {
    run env -i PATH="$PATH" SHELLHUB_EDITION=bogus \
        sh "$GEN_CONFIG" "$BATS_TEST_TMPDIR/config.json"
    [ "$status" -ne 0 ]
    [ ! -f "$BATS_TEST_TMPDIR/config.json" ]
}

@test "a quote in a value survives as data" {
    require_jq
    url='https://x.test/?q="hi"&r=1'
    out=$(gen_config SHELLHUB_ONBOARDING_URL="$url")
    [ "$(printf '%s' "$out" | jq -er .onboardingUrl)" = "$url" ]
}

@test "a backslash in a value survives as data" {
    require_jq
    token='tok\en\\x'
    out=$(gen_config SHELLHUB_CHATWOOT_WEBSITE_TOKEN="$token")
    [ "$(printf '%s' "$out" | jq -er .chatwootWebsiteToken)" = "$token" ]
}

@test "a tab or newline in a value survives as data" {
    require_jq
    url="$(printf 'a\tb\nc')"
    out=$(gen_config SHELLHUB_ONBOARDING_URL="$url")
    [ "$(printf '%s' "$out" | jq -er .onboardingUrl)" = "$url" ]
}

@test "a control character in a value does not corrupt the JSON" {
    require_jq
    out=$(gen_config SHELLHUB_ONBOARDING_URL="$(printf 'a\002b')")
    [ "$(printf '%s' "$out" | jq -er .onboardingUrl)" = "ab" ]
}

@test "a boolean flag renders as a JSON boolean, false when unset" {
    require_jq
    out=$(gen_config)
    [ "$(printf '%s' "$out" | jq -r '.webEndpoints | type')" = "boolean" ]
    [ "$(printf '%s' "$out" | jq -r .webEndpoints)" = "false" ]
    out=$(gen_config SHELLHUB_WEB_ENDPOINTS=true)
    [ "$(printf '%s' "$out" | jq -er '.webEndpoints | type')" = "boolean" ]
    [ "$(printf '%s' "$out" | jq -er .webEndpoints)" = "true" ]
}

@test "a non-boolean flag aborts instead of injecting raw JSON" {
    run env -i PATH="$PATH" SHELLHUB_EDITION=community \
        SHELLHUB_WEB_ENDPOINTS='false, "edition": "cloud"' \
        sh "$GEN_CONFIG" "$BATS_TEST_TMPDIR/config.json"
    [ "$status" -ne 0 ]
    [ ! -f "$BATS_TEST_TMPDIR/config.json" ]
}
