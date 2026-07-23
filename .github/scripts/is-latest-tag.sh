#!/usr/bin/env bash
#
# Exits 0 when the given tag is the highest stable semver among all v* tags
# in the repository. Exits 1 otherwise. Pre-release tags (containing a hyphen
# after the version, e.g. v1.0.0-rc.1) are excluded from the candidate set
# AND are never considered "latest" themselves.
#
# Usage: is-latest-tag.sh <tag>
# Env:   GIT_DIR — override to test against a different repository.

set -euo pipefail

if [[ $# -lt 1 || -z "${1:-}" ]]; then
  echo "Usage: is-latest-tag.sh <tag>" >&2
  exit 2
fi

tag="$1"

if [[ "$tag" == *-* ]]; then
  exit 1
fi

stable_tags=$(git tag -l 'v*' | grep -v -- '-' || true)

if [[ -z "$stable_tags" ]]; then
  exit 0
fi

highest=$(echo "$stable_tags" | sort -V | tail -n1)

if [[ "$tag" == "$highest" ]]; then
  exit 0
fi

exit 1
