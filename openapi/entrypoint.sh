#!/bin/sh

openapi preview-docs community-openapi.yaml --port 8081 &
openapi preview-docs cloud-openapi.yaml --port 8082 &

(cd server && node index.js)
