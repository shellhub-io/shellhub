REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

unset EXTRA_COMPOSE_FILE DOCKER_HOST SHELLHUB_DEV_DOCKER_SOCKET ROOTLESSKIT COMPOSE_FILE COMPOSE_ENV_FILES

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

require_cloud() {
    [ -d "$REPO_ROOT/../cloud" ] || skip "cloud/ not present, skipping cloud-dependent scenario"
}

make_cloud_stub() {
    CLOUD_DIR_OVERRIDE="$BATS_TEST_TMPDIR/cloud-stub"
    mkdir -p "$CLOUD_DIR_OVERRIDE"
    export CLOUD_DIR_OVERRIDE
}
