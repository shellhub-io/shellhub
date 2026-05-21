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
    out=$(capture_with SHELLHUB_ENTERPRISE=true)
    [[ "$out" == *"docker-compose.enterprise.yml"* ]]
}

@test "enterprise: COMPOSE_ENV_FILES loads .env.enterprise" {
    out=$(capture_with SHELLHUB_ENTERPRISE=true)
    [[ "$out" == *".env.enterprise"* ]]
}

@test "CE: COMPOSE_ENV_FILES does not load .env.enterprise" {
    out=$(capture_with)
    [[ "$out" != *".env.enterprise"* ]]
}

@test "mongo database: COMPOSE_FILE includes mongo overlay (and not postgres)" {
    out=$(capture_with SHELLHUB_DATABASE=mongo)
    [[ "$out" == *"docker-compose.mongo.yml"* ]]
    [[ "$out" != *"docker-compose.postgres.yml"* ]]
}

@test "migrate database: includes both mongo and postgres overlays" {
    out=$(capture_with SHELLHUB_DATABASE=migrate)
    [[ "$out" == *"docker-compose.mongo.yml"* ]]
    [[ "$out" == *"docker-compose.postgres.yml"* ]]
}

@test "autossl: COMPOSE_FILE includes autossl overlay" {
    out=$(capture_with SHELLHUB_AUTO_SSL=true)
    [[ "$out" == *"docker-compose.autossl.yml"* ]]
}

@test "guard: SHELLHUB_CLOUD=true without SHELLHUB_ENTERPRISE=true aborts" {
    run capture_with SHELLHUB_CLOUD=true
    [ "$status" -ne 0 ]
    [[ "$output" == *"requires SHELLHUB_ENTERPRISE=true"* ]]
}
