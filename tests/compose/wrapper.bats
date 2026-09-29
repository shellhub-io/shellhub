#!/usr/bin/env bats
# Wrapper decisions that don't require the cloud/ sibling repo.

load helpers

@test "default: COMPOSE_FILE includes docker-compose.yml and the default database overlay" {
    out=$(capture_with)
    [[ "$out" == *"docker-compose.yml"* ]]
    [[ "$out" == *"docker-compose.postgres.yml"* ]]
}

@test "dev mode: COMPOSE_FILE adds dev and agent overlays" {
    out=$(capture_with SHELLHUB_ENV=development)
    [[ "$out" == *"docker-compose.dev.yml"* ]]
    [[ "$out" == *"docker-compose.agent.yml"* ]]
}

@test "prod CE: does not include dev/agent/enterprise overlays" {
    out=$(capture_with)
    [[ "$out" != *"docker-compose.dev.yml"* ]]
    [[ "$out" != *"docker-compose.agent.yml"* ]]
    [[ "$out" != *"docker-compose.enterprise.yml"* ]]
}

@test "enterprise + prod: COMPOSE_FILE includes enterprise overlay" {
    out=$(capture_with SHELLHUB_EDITION=enterprise)
    [[ "$out" == *"docker-compose.enterprise.yml"* ]]
}

@test "enterprise: COMPOSE_ENV_FILES loads .env.enterprise" {
    out=$(capture_with SHELLHUB_EDITION=enterprise)
    [[ "$out" == *".env.enterprise"* ]]
}

@test "CE: COMPOSE_ENV_FILES does not load .env.enterprise" {
    out=$(capture_with)
    [[ "$out" != *".env.enterprise"* ]]
}

@test "unknown database: warns and defaults to postgres" {
    out=$(capture_with SHELLHUB_DATABASE=unknown 2>&1)
    [[ "$out" == *"docker-compose.postgres.yml"* ]]
    [[ "$out" != *"docker-compose.mongo.yml"* ]]
}

@test "autossl: COMPOSE_FILE includes autossl overlay" {
    out=$(capture_with SHELLHUB_AUTO_SSL=true)
    [[ "$out" == *"docker-compose.autossl.yml"* ]]
}

@test "guard: invalid SHELLHUB_EDITION value aborts" {
    run capture_with SHELLHUB_EDITION=invalid
    [ "$status" -ne 0 ]
    [[ "$output" == *"invalid SHELLHUB_EDITION"* ]]
}

@test "precedence: env_override is loaded LAST so user wins over cloud defaults" {
    make_cloud_stub
    echo "SHELLHUB_FROM=cloud" > "$CLOUD_DIR_OVERRIDE/.env"
    out=$(capture_with SHELLHUB_EDITION=enterprise)
    files=$(echo "$out" | grep '^COMPOSE_ENV_FILES=' | sed 's|.*=||')
    # The last entry must be the override tmpfile (lives in BATS_TEST_TMPDIR).
    last=$(echo "$files" | awk -F',' '{print $NF}')
    [[ "$last" == "$BATS_TEST_TMPDIR/"* ]]
    # And cloud/.env must appear earlier in the chain.
    [[ "$files" == *"$CLOUD_DIR_OVERRIDE/.env,"* ]]
}

@test "extra compose file: read from the env override" {
    extra="$BATS_TEST_TMPDIR/compose.instance.yml"
    touch "$extra"
    out=$(capture_with EXTRA_COMPOSE_FILE="$extra")
    [[ "$out" == *":$extra"* ]]
}

@test "extra compose file: the process environment wins over the env override" {
    from_file="$BATS_TEST_TMPDIR/from-file.yml"
    from_env="$BATS_TEST_TMPDIR/from-env.yml"
    touch "$from_file" "$from_env"
    out=$(EXTRA_COMPOSE_FILE="$from_env" capture_with EXTRA_COMPOSE_FILE="$from_file")
    [[ "$out" == *":$from_env"* ]]
    [[ "$out" != *"$from_file"* ]]
}

@test "extra compose file: a missing file aborts" {
    run capture_with EXTRA_COMPOSE_FILE="$BATS_TEST_TMPDIR/missing.yml"
    [ "$status" -ne 0 ]
    [[ "$output" == *"missing.yml"* ]]
}

@test "external postgres: the postgres overlay is left out" {
    out=$(capture_with SHELLHUB_POSTGRES_EXTERNAL=true)
    [[ "$out" != *"docker-compose.postgres.yml"* ]]
}

@test "external postgres: false keeps the postgres overlay" {
    out=$(capture_with SHELLHUB_POSTGRES_EXTERNAL=false)
    [[ "$out" == *"docker-compose.postgres.yml"* ]]
}
