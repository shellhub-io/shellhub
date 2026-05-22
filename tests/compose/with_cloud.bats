#!/usr/bin/env bats
# Wrapper decisions that depend on the cloud/ sibling repo being present.
# Each test skips when ../cloud/ is absent so the file is safe to run anywhere.

load helpers

@test "dev-cloud: COMPOSE_FILE includes cloud/ base and enterprise.dev overlays" {
    require_cloud
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true)
    [[ "$out" == *"../cloud/docker-compose.yml"* ]]
    [[ "$out" == *"../cloud/docker-compose.enterprise.dev.yml"* ]]
}

@test "SHELLHUB_BILLING propagates to COMPOSE_PROFILES" {
    out=$(capture_with SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true SHELLHUB_BILLING=acme)
    [[ "$out" == *"COMPOSE_PROFILES=acme"* ]]
}

@test "user can override SHELLHUB_BILLING to change COMPOSE_PROFILES" {
    out=$(capture_with SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true SHELLHUB_BILLING=other)
    [[ "$out" == *"COMPOSE_PROFILES=other"* ]]
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

@test "symmetric peek: a flag declared only in cloud/.env is visible to the wrapper" {
    make_cloud_stub
    # Declare BILLING only in cloud/.env (not in the shellhub override).
    # With symmetric peek, the wrapper must source cloud/.env and see the
    # value, then export it as COMPOSE_PROFILES.
    printf 'SHELLHUB_BILLING=acme\n' > "$CLOUD_DIR_OVERRIDE/.env"
    out=$(capture_with SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true)
    [[ "$out" == *"COMPOSE_PROFILES=acme"* ]]
}

@test "override wins over cloud/.env for the same flag (last-wins ordering)" {
    make_cloud_stub
    printf 'SHELLHUB_BILLING=from-cloud\n' > "$CLOUD_DIR_OVERRIDE/.env"
    out=$(capture_with SHELLHUB_ENTERPRISE=true SHELLHUB_CLOUD=true SHELLHUB_BILLING=from-override)
    [[ "$out" == *"COMPOSE_PROFILES=from-override"* ]]
    [[ "$out" != *"COMPOSE_PROFILES=from-cloud"* ]]
}

@test "BILLING without CLOUD: profile exported but cloud overlay not loaded" {
    make_cloud_stub
    out=$(capture_with SHELLHUB_BILLING=acme)
    # Profile is exported regardless of CLOUD (orthogonal concerns).
    [[ "$out" == *"COMPOSE_PROFILES=acme"* ]]
    # But cloud/docker-compose.yml is NOT included (CLOUD is false), so any
    # service declared with profiles: [acme] in cloud/ won't actually run.
    [[ "$out" != *"../cloud/docker-compose.yml"* ]]
}
