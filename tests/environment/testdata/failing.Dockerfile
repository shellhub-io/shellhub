ARG ALPINE_VERSION

FROM alpine:${ALPINE_VERSION}

RUN echo agent build fails && exit 3
