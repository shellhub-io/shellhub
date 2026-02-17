#!/bin/sh

SCRIPTS_DIR="$(dirname "$(readlink -f "$0")")"

"$SCRIPTS_DIR/gen-config.sh" /usr/share/nginx/html/ui/config.json

exec nginx -g "daemon off;"
