#!/bin/sh

bundle () {
  openapi bundle community-openapi.yaml > server/www/community-openapi.yaml &
  openapi bundle cloud-openapi.yaml > server/www/cloud-openapi.yaml &
}

watch () {
  echo "Watching for changes in /spec"
  while inotifywait -q -r -e close_write "spec/"
  do
    bundle
  done
}

bundle
watch &

(cd server && node index.js)
