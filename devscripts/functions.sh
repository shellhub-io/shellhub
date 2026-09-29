#!/bin/sh

. "$(dirname "$0")/utils"

get_api_token() {
    local TOKEN
    TOKEN=`http --ignore-stdin post "$SHELLHUB_URL/api/login" username="$1" password="$2" | jq -r .token`
    echo $TOKEN
}
