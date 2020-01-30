#!/bin/sh

set -x

IMAGE_NAME=$1

docker tag $IMAGE_NAME:$CIRCLE_SHA1 $IMAGE_NAME:latest
docker push $IMAGE_NAME:$CIRCLE_SHA1
docker push $IMAGE_NAME:latest

if [ -n "$CIRCLE_TAG" ]; then
    docker tag $IMAGE_NAME:$CIRCLE_SHA1 $IMAGE_NAME:$CIRCLE_TAG
    docker push $IMAGE_NAME:$CIRCLE_TAG
fi
