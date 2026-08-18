#!/usr/bin/env bash
# Refresh the screenshots the docs use, by handing shellhub-demo the manifest the last build
# wrote and the directory the pages read from.
#
# The demo deliberately knows neither path: it photographs whatever list it is given and
# writes wherever it is told, which is what keeps a demo-environment builder free of any
# knowledge of the documentation that consumes it. That boundary constrains the demo, not
# the caller - and this side knows both paths perfectly well, so it fills them in.

set -euo pipefail

docs_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd -P)"
manifest="${docs_dir}/.astro/shots.json"
out_dir="${docs_dir}/public/img/shots"

# Assumed to sit beside the shellhub checkout, which is how the two are normally cloned.
# Overridable because "normally" is not "always", and a wrong guess should be a sentence
# rather than a stack trace.
demo_dir="${SHELLHUB_DEMO_DIR:-${docs_dir}/../../../../shellhub-demo}"

if [ ! -x "${demo_dir}/stage" ]; then
    echo "no shellhub-demo checkout at ${demo_dir}" >&2
    echo "-> clone it beside shellhub, or set SHELLHUB_DEMO_DIR" >&2
    exit 1
fi

# Written at astro:build:done, so a stale manifest means a stale shot list - and a shot
# removed from a page would otherwise keep being photographed forever.
if [ ! -f "$manifest" ]; then
    echo "no shot list at ${manifest}" >&2
    echo "-> npm run build -w @shellhub/docs first" >&2
    exit 1
fi

exec "${demo_dir}/stage" capture --manifest "$manifest" --out "$out_dir" "$@"
