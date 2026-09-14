#!/bin/sh
set -eu

destination=${1:?destination is required}
version=3.17.0

case "$(uname -m)" in
    x86_64) archive="vale_${version}_Linux_64-bit.tar.gz"; sha="a903f1f60c3293fac643e0137f599a462881cc691ee19d6120dcfc786f1be86d" ;;
    aarch64|arm64) archive="vale_${version}_Linux_arm64.tar.gz"; sha="c7da52f10d25fb97e14370b2f77ac5ebdbd23cf0abc156659463cfa785282692" ;;
    *) echo "unsupported architecture: $(uname -m)" >&2; exit 1 ;;
esac

temporary_directory=$(mktemp -d)
trap 'rm -rf "$temporary_directory"' EXIT

curl -fsSL "https://github.com/vale-cli/vale/releases/download/v${version}/${archive}" -o "$temporary_directory/$archive"
echo "$sha  $temporary_directory/$archive" | sha256sum -c -
mkdir -p "$destination"
tar -xzf "$temporary_directory/$archive" -C "$destination" vale
