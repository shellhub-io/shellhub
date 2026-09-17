#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
name=${E2E_STACK_NAME:-default}
edition=${E2E_EDITION:-community}

stack() {
  (cd "$REPO_ROOT/tests" && go run ./cmd/stack "$@")
}

case "${1:-test}" in
  up)
    shift
    stack up --edition "$edition" --name "$name" "$@" | grep '^export '
    ;;
  down)
    stack down --name "$name"
    ;;
  test)
    shift
    [ -n "${CI:-}" ] && trap 'stack down --name "$name"' EXIT
    env=$(stack up --edition "$edition" --name "$name" "$@") || exit $?
    eval "$(echo "$env" | grep '^export ')"
    cd "$REPO_ROOT/ui/apps/console"
    status=0
    npx playwright test || status=$?
    if [ -z "${CI:-}" ]; then
      echo "stack '$name' ($E2E_EDITION) kept at $E2E_BASE_URL; drop it with: npm run e2e:down -w @shellhub/console" >&2
    fi
    exit "$status"
    ;;
  *)
    echo "Usage: $0 {up|down|test} [stack flags]" >&2
    exit 1
    ;;
esac
