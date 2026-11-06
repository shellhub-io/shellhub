#!/bin/sh
set -eu

OUTPUT_DIR="${1:?usage: fetch.sh OUTPUT_DIR}"
BASE_URL="${DBIP_DOWNLOAD_URL:-https://download.db-ip.com/free}"

mkdir -p "$OUTPUT_DIR"

year=$(date -u +%Y)
month=$(date -u +%m)
current="$year-$month"
if [ "$month" = "01" ]; then
    previous="$((year - 1))-12"
else
    previous=$(printf '%s-%02d' "$year" "$((${month#0} - 1))")
fi

for month in "$current" "$previous"; do
    name="dbip-city-lite-$month.mmdb"
    if curl -fsSL -o "$OUTPUT_DIR/$name.gz" "$BASE_URL/$name.gz" && gunzip "$OUTPUT_DIR/$name.gz"; then
        echo "$OUTPUT_DIR/$name"
        exit 0
    fi
    rm -f "$OUTPUT_DIR/$name" "$OUTPUT_DIR/$name.gz"
done

echo "🚫 ERROR: neither $current nor $previous of DB-IP City Lite is published at $BASE_URL" >&2
exit 1
