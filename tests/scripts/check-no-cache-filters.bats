#!/usr/bin/env bats

CHECK="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)/.github/scripts/check-no-cache-filters.sh"

setup() {
    REPO="$BATS_TEST_TMPDIR/repo"
    mkdir -p "$REPO/.github/workflows" "$REPO/app"
    printf 'FROM golang AS builder\nFROM alpine AS production\n' > "$REPO/app/Dockerfile"
    printf 'FROM alpine AS Runtime-Utils\n' > "$REPO/Dockerfile"
    printf 'jobs:\n  build:\n    steps:\n      - uses: actions/checkout@v4\n' > "$REPO/.github/workflows/build.yml"
}

step() {
    cat >> "$REPO/.github/workflows/build.yml"
}

app_step() {
    step <<YAML
      - uses: docker/build-push-action@v6
        with:
          file: app/Dockerfile
          no-cache-filters: $1
YAML
}

check() {
    run bash "$CHECK" "$REPO"
}

assert_passes() {
    [ "$status" -eq 0 ] || { echo "expected the check to pass, got $status:" >&2; echo "$output" >&2; return 1; }
}

assert_fails_with() {
    [ "$status" -eq 1 ] || { echo "expected exit 1, got $status:" >&2; echo "$output" >&2; return 1; }
    [[ "$output" == *"$1"* ]] || { echo "expected '$1' in:" >&2; echo "${output:-<nothing>}" >&2; return 1; }
}

@test "a stage the Dockerfile defines passes" {
    app_step production
    check
    assert_passes
}

@test "a comma list of stages the Dockerfile defines passes" {
    app_step 'builder, production'
    check
    assert_passes
}

@test "a stage the Dockerfile does not define fails" {
    app_step prod
    check
    assert_fails_with "stage prod, which app/Dockerfile does not define"
}

@test "one missing stage in a comma list fails" {
    app_step 'builder,prod'
    check
    assert_fails_with "stage prod,"
}

@test "a stage name that only prefixes a real one fails" {
    app_step produc
    check
    assert_fails_with "stage produc,"
}

@test "a multi-line list of stages the Dockerfile defines passes" {
    step <<'YAML'
      - uses: docker/build-push-action@v6
        with:
          file: app/Dockerfile
          no-cache-filters: |
            builder
            production
YAML
    check
    assert_passes
}

@test "a multi-line list with a missing stage fails" {
    step <<'YAML'
      - uses: docker/build-push-action@v6
        with:
          file: app/Dockerfile
          no-cache-filters: |
            builder
            prod
YAML
    check
    assert_fails_with "stage prod,"
}

@test "a step without file or context checks the root Dockerfile, ignoring case" {
    step <<'YAML'
      - uses: docker/build-push-action@v6
        with:
          no-cache-filters: runtime-utils
YAML
    check
    assert_passes
}

@test "a step without file checks the Dockerfile in its context" {
    step <<'YAML'
      - uses: docker/build-push-action@v6
        with:
          context: app
          no-cache-filters: production
YAML
    check
    assert_passes
}

@test "a step without file does not fall back to the root Dockerfile" {
    step <<'YAML'
      - uses: docker/build-push-action@v6
        with:
          context: app
          no-cache-filters: runtime-utils
YAML
    check
    assert_fails_with "which app/Dockerfile does not define"
}

@test "a Dockerfile that does not exist fails" {
    app_step production
    rm "$REPO/app/Dockerfile"
    check
    assert_fails_with "builds app/Dockerfile, which does not exist"
}

@test "a workflow without no-cache-filters passes" {
    printf 'jobs:\n  lint:\n    steps:\n      - run: true\n' > "$REPO/.github/workflows/lint.yml"
    app_step production
    check
    assert_passes
}

@test "a workflow yq cannot parse fails" {
    app_step production
    printf 'jobs: [unclosed\n' > "$REPO/.github/workflows/broken.yml"
    check
    [ "$status" -ne 0 ] || { echo "expected the check to fail, got:" >&2; echo "$output" >&2; return 1; }
}

checkout_path_job() {
    cat > "$REPO/.github/workflows/publish.yml" <<YAML
jobs:
  publish:
    steps:
      - uses: actions/checkout@v4
        with:
          path: $1
      - uses: actions/checkout@v4
        with:
          repository: example/other
          path: other
      - uses: docker/build-push-action@v6
        with:
          file: $2
          context: $1
          no-cache-filters: $3
YAML
}

@test "a step in a job that checks the repository out under a path resolves the file inside it" {
    checkout_path_job src src/app/Dockerfile production
    check
    assert_passes
}

@test "a missing stage under a checkout path names the file inside the repository" {
    checkout_path_job src src/app/Dockerfile prod
    check
    assert_fails_with "stage prod, which app/Dockerfile does not define"
}

@test "a step without file under a checkout path checks the Dockerfile in its context" {
    cat > "$REPO/.github/workflows/publish.yml" <<'YAML'
jobs:
  publish:
    steps:
      - uses: actions/checkout@v4
        with:
          path: src
      - uses: docker/build-push-action@v6
        with:
          context: src/app
          no-cache-filters: production
YAML
    check
    assert_passes
}

@test "another repository's checkout path does not resolve the file" {
    cat > "$REPO/.github/workflows/publish.yml" <<'YAML'
jobs:
  publish:
    steps:
      - uses: actions/checkout@v4
        with:
          repository: example/other
          path: other
      - uses: docker/build-push-action@v6
        with:
          file: other/app/Dockerfile
          no-cache-filters: production
YAML
    check
    assert_fails_with "builds other/app/Dockerfile, which does not exist"
}

@test "the checkout path of another job does not resolve the file" {
    cat > "$REPO/.github/workflows/build.yml" <<'YAML'
jobs:
  publish:
    steps:
      - uses: actions/checkout@v4
        with:
          path: src
  build:
    steps:
      - uses: actions/checkout@v4
      - uses: docker/build-push-action@v6
        with:
          file: src/app/Dockerfile
          no-cache-filters: production
YAML
    check
    assert_fails_with "builds src/app/Dockerfile, which does not exist"
}

@test "a file that starts with ./ resolves inside the checkout path" {
    checkout_path_job ./src ./src/app/Dockerfile production
    check
    assert_passes
}

@test "a context that starts with ./ resolves inside the checkout path" {
    cat > "$REPO/.github/workflows/publish.yml" <<'YAML'
jobs:
  publish:
    steps:
      - uses: actions/checkout@v4
        with:
          path: ./src
      - uses: docker/build-push-action@v6
        with:
          context: ./src/app
          no-cache-filters: production
YAML
    check
    assert_passes
}

@test "a checkout path with a trailing slash resolves the file inside it" {
    checkout_path_job src/ src/app/Dockerfile production
    check
    assert_passes
}
