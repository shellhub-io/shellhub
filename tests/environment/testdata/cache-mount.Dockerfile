ARG ALPINE_VERSION

FROM alpine:${ALPINE_VERSION}

RUN --mount=type=cache,id=shellhub-e2e-agent-test,target=/cache,sharing=shared \
    touch /cache/probe
