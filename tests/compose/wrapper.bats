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

@test "docker socket: a rootful daemon leaves the variable unset" {
    export STUB_DOCKER_ENDPOINT=unix:///home/dev/.docker/run/docker.sock
    out=$(capture_with SHELLHUB_ENV=development)
    grep -qx 'SHELLHUB_DEV_DOCKER_SOCKET=' <<< "$out"
}

@test "docker socket: a rootless daemon resolves to the context's socket" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_DOCKER_ENDPOINT=unix:///run/user/1000/docker.sock
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=8080 SHELLHUB_SSH_PORT=2222)
    [[ "$out" == *"SHELLHUB_DEV_DOCKER_SOCKET=/run/user/1000/docker.sock"* ]]
}

@test "docker socket: DOCKER_HOST with a unix socket wins over the context" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_DOCKER_ENDPOINT=unix:///run/user/1000/docker.sock
    out=$(DOCKER_HOST=unix:///tmp/other.sock capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=8080 SHELLHUB_SSH_PORT=2222)
    [[ "$out" == *"SHELLHUB_DEV_DOCKER_SOCKET=/tmp/other.sock"* ]]
}

@test "docker socket: DOCKER_HOST with tcp leaves the variable unset" {
    export STUB_DOCKER_SECURITY="name=rootless"
    out=$(DOCKER_HOST=tcp://10.0.0.1:2375 capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=8080 SHELLHUB_SSH_PORT=2222)
    grep -qx 'SHELLHUB_DEV_DOCKER_SOCKET=' <<< "$out"
}

@test "docker socket: not resolved outside development" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_DOCKER_ENDPOINT=unix:///run/user/1000/docker.sock
    out=$(capture_with)
    grep -qx 'SHELLHUB_DEV_DOCKER_SOCKET=' <<< "$out"
}

@test "rootless guard: a privileged port without the capability aborts and names setcap" {
    export STUB_DOCKER_SECURITY="name=rootless name=seccomp,profile=builtin"
    run capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=80
    [ "$status" -ne 0 ]
    [[ "$output" == *"setcap cap_net_bind_service=ep"* ]]
}

@test "rootless guard: the ssh port counts too" {
    export STUB_DOCKER_SECURITY="name=rootless"
    run capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=8080 SHELLHUB_SSH_PORT=22
    [ "$status" -ne 0 ]
    [[ "$output" == *"port 22"* ]]
}

@test "rootless guard: the ssh port keeps its bind address out of the comparison" {
    export STUB_DOCKER_SECURITY="name=rootless"
    run capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=8080 SHELLHUB_SSH_PORT=127.0.0.1:22
    [ "$status" -ne 0 ]
    [[ "$output" == *"port 22"* ]]
}

@test "rootless guard: a missing getcap warns and lets the stack start" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_NO_GETCAP=1
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=80)
    [[ "$out" == *"getcap not found"* ]]
    [[ "$out" == *"COMPOSE_FILE="* ]]
}

@test "rootless guard: the capability on rootlesskit lets a privileged port through" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_GETCAP="cap_net_bind_service=ep"
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=80)
    [[ "$out" == *"COMPOSE_FILE="* ]]
}

@test "rootless guard: a sysctl that frees the port lets it through" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_SYSCTL_UNPRIVILEGED_PORT_START=0
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=80 SHELLHUB_SSH_PORT=22)
    [[ "$out" == *"COMPOSE_FILE="* ]]
}

@test "rootless guard: a sysctl above 1024 makes a higher port privileged too" {
    export STUB_DOCKER_SECURITY="name=rootless" STUB_SYSCTL_UNPRIVILEGED_PORT_START=2048
    run capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=1500 SHELLHUB_SSH_PORT=2222
    [ "$status" -ne 0 ]
    [[ "$output" == *"port 1500"* ]]
}

@test "rootless guard: unprivileged ports pass" {
    export STUB_DOCKER_SECURITY="name=rootless"
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=8080 SHELLHUB_SSH_PORT=2222)
    [[ "$out" == *"COMPOSE_FILE="* ]]
}

@test "rootless guard: a rootful daemon ignores privileged ports" {
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=80)
    [[ "$out" == *"COMPOSE_FILE="* ]]
}

@test "rootless guard: only in development" {
    export STUB_DOCKER_SECURITY="name=rootless"
    out=$(capture_with SHELLHUB_HTTP_PORT=80)
    [[ "$out" == *"COMPOSE_FILE="* ]]
}

@test "rootless guard: a failing docker info skips the guard" {
    export STUB_DOCKER_INFO_FAILS=1
    out=$(capture_with SHELLHUB_ENV=development SHELLHUB_HTTP_PORT=80)
    [[ "$out" == *"COMPOSE_FILE="* ]]
}
