ARG NODE_VERSION
ARG ALPINE_VERSION
ARG SWS_VERSION
ARG REDOCLY_VERSION

FROM ghcr.io/static-web-server/static-web-server:${SWS_VERSION} AS sws

FROM node:${NODE_VERSION} AS base

RUN apk add --update build-base python3 curl git gcompat

COPY ui/scripts/install-vale.sh /tmp/install-vale
RUN /tmp/install-vale /usr/local/bin && rm /tmp/install-vale

WORKDIR /app

COPY ui/package.json ui/package-lock.json ./
COPY ui/packages/design-system/package.json ./packages/design-system/
COPY ui/apps/console/package.json ./apps/console/
COPY ui/apps/website/package.json ./apps/website/
COPY ui/apps/docs/package.json ./apps/docs/

RUN npm install

FROM base AS development

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

FROM node:${NODE_VERSION} AS spec-bundler
ARG REDOCLY_VERSION
RUN npm install -g @redocly/cli@${REDOCLY_VERSION:?}
WORKDIR /spec
COPY openapi/spec ./spec
COPY openapi/plugins ./plugins
COPY openapi/redocly.yaml .
RUN redocly bundle spec/openapi.yaml -o openapi.json

FROM base AS builder

WORKDIR /app

COPY ui/. .

COPY --from=base /app/node_modules ./node_modules
COPY --from=spec-bundler /spec/openapi.json ./apps/console/openapi.json

RUN OPENAPI_SPEC_PATH=./openapi.json npm run generate -w @shellhub/console
RUN npm run build -w @shellhub/console

RUN apk add --no-cache brotli gzip && \
    find apps/console/dist -type f -size +1024c \
      \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.json' \
         -o -name '*.svg' -o -name '*.ico' \) \
      ! -name 'config.json' \
      -exec brotli -q 11 -k {} \; \
      -exec gzip -9 -k {} \;

FROM alpine:${ALPINE_VERSION} AS production

RUN apk upgrade --no-cache libssl3 libcrypto3

COPY --from=sws /static-web-server /usr/local/bin/static-web-server

RUN addgroup -g 1000 -S shellhub && adduser -u 1000 -S -G shellhub shellhub

COPY ui/sws.toml /etc/sws.toml
COPY ui/scripts /scripts

COPY --from=builder --chown=1000:1000 /app/apps/console/dist /var/www

USER 1000:1000

EXPOSE 8080

CMD ["/scripts/entrypoint.sh"]
