ARG DOCKER_VERSION
ARG PLAYWRIGHT_VERSION

FROM docker:${DOCKER_VERSION}-cli AS docker

FROM mcr.microsoft.com/playwright:v${PLAYWRIGHT_VERSION}-noble

COPY --from=docker /usr/local/bin/docker /usr/local/bin/docker
COPY --from=docker /usr/local/libexec/docker/cli-plugins/docker-compose /usr/local/libexec/docker/cli-plugins/docker-compose
