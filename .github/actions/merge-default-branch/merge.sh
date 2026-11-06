#!/usr/bin/env bash

set -euo pipefail

if [[ $# -ne 2 || -z "$1" || -z "$2" ]]; then
  echo "Usage: merge.sh <checked-out-ref> <default-branch>" >&2
  exit 2
fi

ref="$1"
default="$2"

if [[ "$ref" == "$default" ]]; then
  exit 0
fi

fetch=(git fetch --quiet --no-tags)
if [[ "$(git rev-parse --is-shallow-repository)" == true ]]; then
  fetch+=(--unshallow)
fi
"${fetch[@]}" origin "+refs/heads/$default:refs/remotes/origin/$default"

if ! git -c user.name="github-actions[bot]" \
  -c user.email="41898282+github-actions[bot]@users.noreply.github.com" \
  merge --quiet --no-ff --no-edit "origin/$default" >/dev/null; then
  git merge --abort
  echo "::error::$ref does not merge cleanly with $default; rebase it onto $default" >&2
  exit 1
fi
