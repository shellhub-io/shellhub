ARG DOCKERHUB_MIRROR=
ARG ALPINE_VERSION

FROM ${DOCKERHUB_MIRROR}alpine:${ALPINE_VERSION}

RUN --mount=type=cache,id=shellhub-e2e-agent-test,target=/cache,sharing=shared \
    touch /cache/probe
