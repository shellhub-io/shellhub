#!/bin/sh

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

CONFIG=/var/www/config.json

"$SCRIPTS_DIR/gen-config.sh" "$CONFIG" || exit 1

# A sibling would outrank the file just written for every client accepting it.
rm -f "$CONFIG.br" "$CONFIG.gz"

exec static-web-server -w /etc/sws.toml
