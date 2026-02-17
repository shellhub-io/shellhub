#!/bin/sh

SCRIPTS_DIR=$(dirname $(readlink -f $0))

$SCRIPTS_DIR/env.sh SHELLHUB > /usr/share/nginx/html/ui/env.js
$SCRIPTS_DIR/env.sh SHELLHUB > /usr/share/nginx/html/website/env.js

exec nginx-debug -g "daemon off;"
