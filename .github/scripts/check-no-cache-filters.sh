#!/usr/bin/env bash

set -euo pipefail

root=${1:-.}
status=0

for workflow in "$root"/.github/workflows/*.yml; do
  rows=$(yq -r '.jobs[].steps[]? | select(.with."no-cache-filters") | [.with.file // ((.with.context // ".") + "/Dockerfile"), (.with."no-cache-filters" | sub("\n", ","))] | @tsv' "$workflow")
  while IFS=$'\t' read -r file filters; do
    [[ -n "$file" ]] || continue
    dockerfile="$root/$file"
    if [[ ! -f "$dockerfile" ]]; then
      echo "check-no-cache-filters.sh: ${workflow#"$root"/} builds $file, which does not exist" >&2
      status=1
      continue
    fi
    IFS=',' read -ra stages <<<"$filters"
    for stage in "${stages[@]}"; do
      stage=${stage//[[:space:]]/}
      [[ -n "$stage" ]] || continue
      if ! grep -qiE "^FROM[[:space:]].*[[:space:]]AS[[:space:]]+${stage}[[:space:]]*$" "$dockerfile"; then
        echo "check-no-cache-filters.sh: ${workflow#"$root"/} skips the cache for stage $stage, which $file does not define" >&2
        status=1
      fi
    done
  done <<<"$rows"
done

exit "$status"
