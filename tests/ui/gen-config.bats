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
