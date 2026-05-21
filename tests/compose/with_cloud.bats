#!/usr/bin/env bats
# Wrapper decisions that depend on the cloud/ sibling repo being present.
# Each test skips when ../cloud/ is absent so the file is safe to run anywhere.

load helpers

@test "dev-cloud: COMPOSE_FILE includes cloud/ base, dev, and enterprise.dev overlays" {
    require_cloud
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true)
    [[ "$out" == *"../cloud/docker-compose.yml"* ]]
    [[ "$out" == *"../cloud/docker-compose.dev.yml"* ]]
    [[ "$out" == *"../cloud/docker-compose.enterprise.dev.yml"* ]]
}

@test "prod-cloud: COMPOSE_FILE includes cloud/base and shellhub/enterprise (no dev overlays)" {
    require_cloud
    out=$(capture_with SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true)
    [[ "$out" == *"docker-compose.enterprise.yml"* ]]
    [[ "$out" == *"../cloud/docker-compose.yml"* ]]
    [[ "$out" != *"../cloud/docker-compose.dev.yml"* ]]
    [[ "$out" != *"../cloud/docker-compose.enterprise.dev.yml"* ]]
}

@test "prod-cloud: COMPOSE_ENV_FILES loads cloud/.env" {
    require_cloud
    out=$(capture_with SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true)
    [[ "$out" == *"../cloud/.env"* ]]
}

@test "dev-enterprise: COMPOSE_FILE includes cloud/enterprise.dev.yml even without CLOUD=true" {
    require_cloud
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_ENTERPRISE=true)
    [[ "$out" == *"../cloud/docker-compose.enterprise.dev.yml"* ]]
}

@test "dev-enterprise without CLOUD: does not load cloud base/dev overlays" {
    require_cloud
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_ENTERPRISE=true)
    [[ "$out" != *"../cloud/docker-compose.yml"* ]]
    [[ "$out" != *"../cloud/docker-compose.dev.yml"* ]]
}

@test "enterprise: COMPOSE_ENV_FILES loads cloud/.env when cloud/ is present" {
    require_cloud
    out=$(capture_with SHELLHUB_ENTERPRISE=true)
    [[ "$out" == *"../cloud/.env"* ]]
}
