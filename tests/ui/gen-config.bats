#!/usr/bin/env bats
# Tests for ui/scripts/gen-config.sh.

REPO_ROOT="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)"
GEN_CONFIG="$REPO_ROOT/ui/scripts/gen-config.sh"

# Run gen-config.sh with a pristine environment holding only the given
# VAR=value pairs, then print the config.json it produced. The empty
# environment is the point: it reproduces the container, which receives
# nothing the deployment did not explicitly pass in.
# Usage: gen_config VAR1=value VAR2=value ...
gen_config() {
    local out="$BATS_TEST_TMPDIR/config.json"
    env -i PATH="$PATH" "$@" sh "$GEN_CONFIG" "$out" || return
    cat "$out"
}

# Skip when jq is absent. Assertions about the *shape* of the output need a
# real parser; jq ships on the CI runner, so this only spares local devs.
require_jq() {
    command -v jq >/dev/null || skip "jq not installed"
}

@test "stripePublishableKey comes from SHELLHUB_STRIPE_PUBLISHABLE_KEY" {
    out=$(gen_config SHELLHUB_STRIPE_PUBLISHABLE_KEY=pk_test_123)
    [[ "$out" == *'"stripePublishableKey": "pk_test_123"'* ]]
}

@test "every value reads a SHELLHUB_-prefixed variable" {
    # Deployments only ever pass SHELLHUB_*, so an unprefixed reference is
    # never satisfied and silently renders empty. Guards shellhub#6865.
    # EDITION and OUTPUT are exempt because the script assigns them itself;
    # nothing else belongs in that list.
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

# Values arrive verbatim from the deployment's env, so the script has to treat
# them as data rather than as JSON source text.

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

@test "boolean flags render as JSON booleans" {
    require_jq
    out=$(gen_config SHELLHUB_ANNOUNCEMENTS=true)
    [ "$(printf '%s' "$out" | jq -er '.announcements | type')" = "boolean" ]
    [ "$(printf '%s' "$out" | jq -er '.webEndpoints | type')" = "boolean" ]
    [ "$(printf '%s' "$out" | jq -er .announcements)" = "true" ]
}

@test "a non-boolean flag aborts instead of injecting raw JSON" {
    # Unquoted fields are the opening: a crafted value can append a second
    # "edition" key that wins at parse time, silently overriding the guard
    # above and promoting a community deployment to cloud.
    run env -i PATH="$PATH" SHELLHUB_EDITION=community \
        SHELLHUB_ANNOUNCEMENTS='false, "edition": "cloud"' \
        sh "$GEN_CONFIG" "$BATS_TEST_TMPDIR/config.json"
    [ "$status" -ne 0 ]
    [ ! -f "$BATS_TEST_TMPDIR/config.json" ]
}
