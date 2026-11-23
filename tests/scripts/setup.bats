#!/usr/bin/env bats

REPO_ROOT="$(cd "$(dirname "${BATS_TEST_FILENAME}")/../.." && pwd -P)"
SETUP="$REPO_ROOT/bin/setup"

setup() {
    STUBS="$BATS_TEST_TMPDIR/bin"
    LOG="$BATS_TEST_TMPDIR/log"
    mkdir -p "$STUBS" "$LOG"

    cat > "$STUBS/docker" <<'EOF'
#!/bin/sh
[ "$1 $2" = "compose ps" ] || exit 0
case " $* " in
    *" --status running "*) ;;
    *) printf '%s\n' server ui; exit 0 ;;
esac
case " $* " in
    *" --services "*) printf '%s\n' ${STUB_RUNNING-server} ;;
esac
EOF
    cat > "$STUBS/curl" <<'EOF'
#!/bin/sh
for arg in "$@"; do last=$arg; done
printf '%s\n' "$last" >> "$STUB_LOG/curl"
printf '%s' "${STUB_INFO-}"
EOF
    cat > "$STUBS/xdg-open" <<'EOF'
#!/bin/sh
printf '%s\n' "$1" >> "$STUB_LOG/opened"
EOF
    chmod +x "$STUBS"/*

    for tool in /bin/* /usr/bin/* /usr/local/bin/*; do
        name=${tool##*/}
        case "$name" in
            open|x-www-browser|xdg-open|docker|curl) continue ;;
        esac
        [ -e "$STUBS/$name" ] || ln -s "$tool" "$STUBS/$name"
    done

    OVERRIDE="$BATS_TEST_TMPDIR/env.override"
    : > "$OVERRIDE"
    mkdir -p "$BATS_TEST_TMPDIR/cloud"
}

with_env() {
    printf '%s\n' "$@" >> "$OVERRIDE"
}

run_setup() {
    run env PATH="$STUBS" STUB_LOG="$LOG" ENV_OVERRIDE="$OVERRIDE" CLOUD_DIR="$BATS_TEST_TMPDIR/cloud" \
        "$@" "$SETUP"
}

without_browser() {
    rm "$STUBS/xdg-open"
}

opened_url() {
    for _ in $(seq 50); do
        [ -s "$LOG/opened" ] && break
        sleep 0.1
    done
    cat "$LOG/opened"
}

assert_checked_and_opened() {
    [ "$status" -eq 0 ] || { echo "exit $status: $output" >&2; return 1; }
    [ "$(cat "$LOG/curl")" = "$1/info" ] || { echo "curl got: $(cat "$LOG/curl")" >&2; return 1; }
    [ "$(opened_url)" = "$1/setup" ] || { echo "opened: $(opened_url)" >&2; return 1; }
}

@test "the defaults check and open localhost on port 80" {
    run_setup
    assert_checked_and_opened "http://localhost"
}

@test "a domain replaces localhost" {
    with_env SHELLHUB_DOMAIN=shellhub.example.com
    run_setup
    assert_checked_and_opened "http://shellhub.example.com"
}

@test "a custom HTTP port is part of the URL" {
    with_env SHELLHUB_DOMAIN=shellhub.example.com SHELLHUB_HTTP_PORT=8080
    run_setup
    assert_checked_and_opened "http://shellhub.example.com:8080"
}

@test "automatic HTTPS on the default port uses https without a port" {
    with_env SHELLHUB_DOMAIN=shellhub.example.com SHELLHUB_AUTO_SSL=true SHELLHUB_HTTP_PORT=8080
    run_setup
    assert_checked_and_opened "https://shellhub.example.com"
}

@test "automatic HTTPS on a custom port uses the HTTPS port" {
    with_env SHELLHUB_DOMAIN=shellhub.example.com SHELLHUB_AUTO_SSL=true SHELLHUB_HTTPS_PORT=8443
    run_setup
    assert_checked_and_opened "https://shellhub.example.com:8443"
}

@test "the public URL wins over the domain and ports" {
    with_env SHELLHUB_DOMAIN=ignored.example.com SHELLHUB_HTTP_PORT=8080 \
        SHELLHUB_PUBLIC_URL=https://proxy.example.com/
    run_setup
    assert_checked_and_opened "https://proxy.example.com"
}

@test "an enterprise edition reads the cloud env file under the override" {
    printf '%s\n' SHELLHUB_DOMAIN=cloud.example.com SHELLHUB_HTTP_PORT=8080 > "$BATS_TEST_TMPDIR/cloud/.env"
    with_env SHELLHUB_EDITION=enterprise SHELLHUB_HTTP_PORT=9090
    run_setup
    assert_checked_and_opened "http://cloud.example.com:9090"
}

@test "the community edition ignores the cloud env file" {
    printf '%s\n' SHELLHUB_DOMAIN=cloud.example.com > "$BATS_TEST_TMPDIR/cloud/.env"
    run_setup
    assert_checked_and_opened "http://localhost"
}

@test "a completed setup exits without opening anything" {
    with_env SHELLHUB_DOMAIN=shellhub.example.com SHELLHUB_HTTP_PORT=8080
    run_setup STUB_INFO='{"setup":true}'
    [ "$status" -eq 0 ]
    [[ "$output" == *"already been completed"* ]]
    [ "$(cat "$LOG/curl")" = "http://shellhub.example.com:8080/info" ]
    [ ! -e "$LOG/opened" ]
}

@test "a stopped stack fails before checking the instance" {
    run_setup STUB_RUNNING=
    [ "$status" -ne 0 ]
    [[ "$output" == *"not running"* ]]
    [ ! -e "$LOG/curl" ]
}

@test "a stack running without the server fails before checking the instance" {
    run_setup STUB_RUNNING=ui
    [ "$status" -ne 0 ]
    [[ "$output" == *"not running"* ]]
    [ ! -e "$LOG/curl" ]
}

@test "an invalid edition fails before checking the instance" {
    with_env SHELLHUB_EDITION=ee
    run_setup
    [ "$status" -ne 0 ]
    [[ "$output" == *"invalid SHELLHUB_EDITION 'ee'"* ]]
    [ ! -e "$LOG/curl" ]
}

@test "without a browser the URL is printed" {
    without_browser
    with_env SHELLHUB_DOMAIN=shellhub.example.com
    run_setup
    [ "$status" -eq 0 ]
    [[ "$output" == *"Open this URL in your browser to complete setup: http://shellhub.example.com/setup"* ]]
}
