# Helpers for tests/compose/*.bats
#
# Strategy: stub `docker` in PATH so the wrapper's `exec docker compose`
# captures the resulting COMPOSE_FILE and COMPOSE_ENV_FILES values instead of
# actually invoking Compose. Lets us assert on the wrapper's decisions
# (which overlays it chains, which env files it loads, when it aborts)
# without starting containers or even needing a working docker daemon.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

unset EXTRA_COMPOSE_FILE DOCKER_HOST SHELLHUB_DEV_DOCKER_SOCKET ROOTLESSKIT COMPOSE_FILE COMPOSE_ENV_FILES

# Usage: capture_with VAR1=value VAR2=value ...
capture_with() {
    local stub_dir="$BATS_TEST_TMPDIR/stub"
    if [ ! -x "$stub_dir/docker" ]; then
        mkdir -p "$stub_dir"
        cat > "$stub_dir/docker" <<'EOF'
#!/bin/sh
case "$1" in
    context)
        echo "${STUB_DOCKER_ENDPOINT:-unix:///var/run/docker.sock}"
        ;;
    info)
        [ -z "${STUB_DOCKER_INFO_FAILS:-}" ] || exit 1
        echo "[${STUB_DOCKER_SECURITY:-name=seccomp,profile=builtin name=cgroupns}]"
        ;;
    compose)
        if [ "$2" = config ]; then
            [ -n "${COMPOSE_FILE:-}" ] || exit 1
            echo "services:"
            echo "  gateway:"
            echo "    ports:"
            for port in ${STUB_PUBLISHED:-}; do
                echo "      - mode: ingress"
                case "$port" in
                    *:*) echo "        host_ip: ${port%:*}" ;;
                esac
                echo "        target: 1"
                echo "        published: \"${port##*:}\""
                echo "        protocol: tcp"
            done
            exit 0
        fi
        echo "COMPOSE_FILE=$COMPOSE_FILE"
        echo "COMPOSE_ENV_FILES=$COMPOSE_ENV_FILES"
        echo "COMPOSE_PROFILES=$COMPOSE_PROFILES"
        echo "SHELLHUB_DEV_DOCKER_SOCKET=$SHELLHUB_DEV_DOCKER_SOCKET"
        ;;
esac
EOF
        cat > "$stub_dir/getcap" <<'EOF'
#!/bin/sh
[ -z "${STUB_NO_GETCAP:-}" ] || exit 127
[ -z "${STUB_GETCAP:-}" ] || echo "$1 $STUB_GETCAP"
EOF
        cat > "$stub_dir/sysctl" <<'EOF'
#!/bin/sh
echo "${STUB_SYSCTL_UNPRIVILEGED_PORT_START:-1024}"
EOF
        : > "$stub_dir/rootlesskit"
        chmod +x "$stub_dir"/*
    fi
    [ -z "${STUB_NO_ROOTLESSKIT:-}" ] || export ROOTLESSKIT=

    local tmp
    tmp=$(mktemp -p "$BATS_TEST_TMPDIR")
    for var in "$@"; do
        [ -n "$var" ] && printf '%s\n' "$var" >> "$tmp"
    done

    PATH="$stub_dir:$PATH" ENV_OVERRIDE="$tmp" CLOUD_DIR="${CLOUD_DIR_OVERRIDE:-$REPO_ROOT/../cloud}" \
        "$REPO_ROOT/bin/docker-compose" compose 2>&1
}

# Skip the current test when the cloud/ sibling repo is not checked out.
require_cloud() {
    [ -d "$REPO_ROOT/../cloud" ] || skip "cloud/ not present, skipping cloud-dependent scenario"
}

# Create an empty cloud/ stub directory inside BATS_TEST_TMPDIR for tests
# that need cloud/ "present" but want deterministic content (avoiding
# dependency on the dev's actual cloud/.env or cloud/.env.override).
# Sets CLOUD_DIR_OVERRIDE so subsequent capture_with calls use the stub.
make_cloud_stub() {
    CLOUD_DIR_OVERRIDE="$BATS_TEST_TMPDIR/cloud-stub"
    mkdir -p "$CLOUD_DIR_OVERRIDE"
    export CLOUD_DIR_OVERRIDE
}
