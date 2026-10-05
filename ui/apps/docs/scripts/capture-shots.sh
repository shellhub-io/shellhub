#!/usr/bin/env bash

set -euo pipefail

docs_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
manifest="${docs_dir}/.astro/shots.json"
out_dir="${docs_dir}/public/img/shots"

demo_dir="${SHELLHUB_DEMO_DIR:-${docs_dir}/../../../../shellhub-demo}"

if [ ! -x "${demo_dir}/stage" ]; then
    echo "no shellhub-demo checkout at ${demo_dir}" >&2
    echo "-> clone it beside shellhub, or set SHELLHUB_DEMO_DIR" >&2
    exit 1
fi

if [ ! -f "$manifest" ]; then
    echo "no shot list at ${manifest}" >&2
    echo "-> npm run build -w @shellhub/docs first" >&2
    exit 1
fi

exec "${demo_dir}/stage" capture --manifest "$manifest" --out "$out_dir" "$@"
