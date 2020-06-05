#!/bin/sh

get_api_token() {
    local TOKEN
    TOKEN=`http --ignore-stdin post http://localhost/api/login username="$1" password="$2" | jq -r .token`
    echo $TOKEN
}
