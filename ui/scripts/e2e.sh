#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd -P)"
name_args=()
[ -n "${E2E_STACK_NAME:-}" ] && name_args=(--name "$E2E_STACK_NAME")
edition=${E2E_EDITION:-community}
runner=${E2E_RUNNER:-docker}

case $runner in
  host | docker) ;;
  *)
    echo "E2E_RUNNER must be host or docker, not '$runner'" >&2
    exit 1
    ;;
esac

compose_run() {
  "$REPO_ROOT/bin/docker-compose" run --rm -T "$@"
}

require_tools_services() {
  local services
  services=$("$REPO_ROOT/bin/docker-compose" --profile tools config --services) || exit 1
  grep -qx test <<<"$services" && return
  echo "the docker runner needs the test and e2e services from docker-compose.dev.yml: set" \
    "SHELLHUB_ENV=development in .env.override, or run with E2E_RUNNER=host" >&2
  exit 1
}

stack() {
  if [ "$runner" = host ]; then
    (cd "$REPO_ROOT/tests" && go run ./cmd/stack "$@")
    return
  fi

  local run_args=(-w "$REPO_ROOT/tests")
  local var
  for var in STRIPE_SECRET_KEY STRIPE_PRICE_ID SHELLHUB_STRIPE_PUBLISHABLE_KEY; do
    run_args+=(-e "$var")
  done
  compose_run "${run_args[@]}" test go run ./cmd/stack "$@"
}

stack_bin="$REPO_ROOT/tests/.stack/bin/stack"

build_stack() {
  if [ "$runner" = host ]; then
    (cd "$REPO_ROOT/tests" && CGO_ENABLED=0 go build -o "$stack_bin" ./cmd/stack)
    return
  fi

  compose_run -w "$REPO_ROOT/tests" -e CGO_ENABLED=0 test go build -o "$stack_bin" ./cmd/stack
}

playwright() {
  if [ "$runner" = host ]; then
    (cd "$REPO_ROOT/ui/apps/console" && npx playwright test "$@")
    return
  fi

  compose_run \
    -e E2E_COMPOSE_PROJECT -e E2E_BASE_URL -e E2E_EDITION -e E2E_ADMIN_USER -e E2E_ADMIN_PASSWORD -e E2E_ADMIN_NAMESPACE \
    -e E2E_AGENT_IMAGE -e E2E_ACCOUNT_LOCKOUT_BASE_MS -e E2E_MAILPIT_URL \
    -e E2E_STACK_BIN -e E2E_STACK_NAME -e E2E_STACK_DIR \
    -e E2E_EXPIRED_LICENSE -e E2E_LICENSE_ISSUER_KEY \
    -e CI \
    e2e npx playwright test "$@"
}

if [ "$runner" = docker ]; then
  require_tools_services
fi

case "${1:-test}" in
  up)
    shift
    build_stack >&2
    stack up --edition "$edition" ${name_args[@]+"${name_args[@]}"} "$@" | grep '^export '
    echo "export E2E_STACK_BIN=$stack_bin"
    ;;
  down)
    stack down ${name_args[@]+"${name_args[@]}"}
    ;;
  test)
    shift
    playwright_args=()
    for arg; do
      case $arg in
        --edition=*) edition=${arg#*=} ;;
        *) playwright_args+=("$arg") ;;
      esac
    done
    if [ -n "${CI:-}" ] && [ -z "${E2E_KEEP_STACK:-}" ]; then
      trap 'stack down ${name_args[@]+"${name_args[@]}"}' EXIT
    fi
    build_stack >&2
    env=$(stack up --edition "$edition" ${name_args[@]+"${name_args[@]}"}) || exit $?
    eval "$(echo "$env" | grep '^export ')"
    export E2E_STACK_BIN=$stack_bin
    status=0
    playwright ${playwright_args[@]+"${playwright_args[@]}"} || status=$?
    if [ -z "${CI:-}" ]; then
      echo "stack '$E2E_COMPOSE_PROJECT' ($E2E_EDITION) kept at $E2E_BASE_URL; drop it with: ui/scripts/e2e.sh down" >&2
    fi
    exit "$status"
    ;;
  *)
    echo "Usage: $0 {up [stack flags]|down|test [--edition=<edition>] [playwright args]}" >&2
    exit 1
    ;;
esac
