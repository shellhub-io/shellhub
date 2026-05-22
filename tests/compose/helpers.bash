# Helpers for tests/compose/*.bats
#
# Strategy: stub `docker` in PATH so the wrapper's `exec docker compose`
# captures the resulting COMPOSE_FILE and COMPOSE_ENV_FILES values instead of
# actually invoking Compose. Lets us assert on the wrapper's decisions
# (which overlays it chains, which env files it loads, when it aborts)
# without starting containers or even needing a working docker daemon.

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

# Run the wrapper with the given override variables written to a tmpfile,
# while replacing `docker` in PATH with a stub that just echoes the env
# the wrapper exported.
# Usage: capture_with VAR1=value VAR2=value ...
capture_with() {
    local stub_dir="$BATS_TEST_TMPDIR/stub"
    if [ ! -x "$stub_dir/docker" ]; then
        mkdir -p "$stub_dir"
        cat > "$stub_dir/docker" <<'EOF'
#!/bin/sh
echo "COMPOSE_FILE=$COMPOSE_FILE"
echo "COMPOSE_ENV_FILES=$COMPOSE_ENV_FILES"
echo "COMPOSE_PROFILES=$COMPOSE_PROFILES"
EOF
        chmod +x "$stub_dir/docker"
    fi

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
