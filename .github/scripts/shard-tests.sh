#!/usr/bin/env bash

set -euo pipefail

if [[ $# -lt 2 || $# -gt 3 ]]; then
  echo "Usage: shard-tests.sh <shard-count> <shard> [first-shard-regex] < test-names" >&2
  exit 2
fi

count=$1
shard=$2
first=${3:-}

if ! [[ "$count" =~ ^[1-9][0-9]*$ && "$shard" =~ ^[1-9][0-9]*$ ]] || ((shard > count)); then
  echo "shard-tests.sh: the shard must be between 1 and the shard count, got $shard of $count" >&2
  exit 2
fi

names=$(LC_ALL=C sort -u)

if [[ -z "$names" ]]; then
  echo "shard-tests.sh: no test names on stdin" >&2
  exit 1
fi

if invalid=$(grep -vE '^[A-Za-z_][A-Za-z0-9_]*$' <<<"$names"); then
  echo "shard-tests.sh: not a Go test name: $invalid" >&2
  exit 1
fi

shard_of() {
  local crc
  if [[ -n "$first" && "$1" =~ $first ]]; then
    echo 1
    return
  fi
  read -r crc _ < <(printf '%s' "$1" | cksum)
  echo $((crc % count + 1))
}

selected=$(
  while read -r name; do
    if (($(shard_of "$name") == shard)); then
      echo "$name"
    fi
  done <<<"$names" | paste -sd'|' -
)

if [[ -z "$selected" ]]; then
  echo "shard-tests.sh: shard $shard of $count holds no tests" >&2
  exit 1
fi

echo "^(${selected})\$"
