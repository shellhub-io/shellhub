#!/bin/bash

set -x

IMAGE_NAME=$1
ARCHS=(amd64 arm32v6 arm64v8)
DOCKER_ARCHS=(amd64 arm arm64)

for i in ${!ARCHS[@]}; do
    docker push $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]}
    docker manifest create -a $IMAGE_NAME:${CIRCLE_SHA1} $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]}
    docker manifest annotate $IMAGE_NAME:${CIRCLE_SHA1} $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]} --os linux --arch ${DOCKER_ARCHS[i]}
done

docker manifest push $IMAGE_NAME:${CIRCLE_SHA1}

for i in ${!ARCHS[@]}; do
    docker manifest create -a $IMAGE_NAME:latest $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]}
    docker manifest annotate $IMAGE_NAME:latest $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]} --os linux --arch ${DOCKER_ARCHS[i]}
done

docker manifest push $IMAGE_NAME:latest

if [ -n "$CIRCLE_TAG" ]; then
    for i in ${!ARCHS[@]}; do
        docker push $IMAGE_NAME:${CIRCLE_TAG}-${ARCHS[i]}
        docker manifest create -a $IMAGE_NAME:${CIRCLE_TAG} $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]}
        docker manifest annotate $IMAGE_NAME:${CIRCLE_TAG} $IMAGE_NAME:${CIRCLE_SHA1}-${ARCHS[i]} --os linux --arch ${DOCKER_ARCHS[i]}
    done

    docker manifest push $IMAGE_NAME:${CIRCLE_TAG}
fi
