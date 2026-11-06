#!/usr/bin/env bats

MERGE="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)/.github/actions/merge-default-branch/merge.sh"

commit() {
    printf '%s\n' "$2" > "$1"
    git add "$1"
    git -c user.name=test -c user.email=test@example.com commit --quiet -m "$1: $2"
}

setup() {
    export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1

    ORIGIN="$BATS_TEST_TMPDIR/origin"
    git init --quiet --initial-branch=master "$ORIGIN"
    cd "$ORIGIN"
    commit base.txt base
    git checkout --quiet -b paired
    commit paired.txt paired
    git checkout --quiet master
    commit master.txt master
}

clone() {
    CLONE="$BATS_TEST_TMPDIR/clone"
    git clone --quiet --depth 1 --branch "$1" "file://$ORIGIN" "$CLONE"
    cd "$CLONE"
}

@test "a shallow paired branch gets the default branch merged in" {
    clone paired

    run "$MERGE" paired master

    [ "$status" -eq 0 ] || { echo "$output" >&2; return 1; }
    [ "$(cat master.txt)" = master ]
    [ "$(cat paired.txt)" = paired ]
    [ "$(git rev-list --count --merges HEAD)" -eq 1 ]
    [ "$(git log -1 --format=%cn)" = 'github-actions[bot]' ]
}

@test "the default branch itself is left as checked out" {
    clone master
    before=$(git rev-parse HEAD)

    run "$MERGE" master master

    [ "$status" -eq 0 ]
    [ "$(git rev-parse HEAD)" = "$before" ]
    [ "$(git rev-parse --is-shallow-repository)" = true ]
}

@test "a paired branch that conflicts with the default branch fails and is left clean" {
    cd "$ORIGIN"
    commit base.txt master-side
    git checkout --quiet paired
    commit base.txt paired-side
    git checkout --quiet master
    clone paired
    before=$(git rev-parse HEAD)

    run "$MERGE" paired master

    [ "$status" -eq 1 ]
    [[ "$output" == *"paired does not merge cleanly with master"* ]]
    [ "$(git rev-parse HEAD)" = "$before" ]
    [ -z "$(git status --porcelain)" ]
}

@test "a call without both refs is refused" {
    clone paired

    run "$MERGE" paired

    [ "$status" -eq 2 ]
}
