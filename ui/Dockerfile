FROM node:24.16.0-alpine3.22 AS base

ARG NPM_CONFIG_REGISTRY

RUN apk add --update build-base python3 curl git

WORKDIR /app

# Copy workspace package files for npm install
COPY ui/package.json ui/package-lock.json ./
COPY ui/packages/design-system/package.json ./packages/design-system/
COPY ui/apps/console/package.json ./apps/console/
COPY ui/apps/website/package.json ./apps/website/
COPY ui/apps/docs/package.json ./apps/docs/

RUN npm install

FROM base AS development

ARG NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_REGISTRY=${NPM_CONFIG_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

# Bundle OpenAPI spec from YAML sources
FROM node:24.16.0-alpine3.22 AS spec-bundler
RUN npm install -g @redocly/cli@2.31.5
WORKDIR /spec
COPY openapi/spec ./spec
COPY openapi/plugins ./plugins
COPY openapi/redocly.yaml .
RUN redocly bundle spec/openapi.yaml -o openapi.json

FROM base AS builder

ARG NPM_CONFIG_REGISTRY

WORKDIR /app

COPY ui/. .

COPY --from=base /app/node_modules ./node_modules
COPY --from=spec-bundler /spec/openapi.json ./apps/console/openapi.json

RUN OPENAPI_SPEC_PATH=./openapi.json npm run generate -w @shellhub/console
RUN npm run build -w @shellhub/console

# Caddy has no brotli encoder, so a sibling on disk is the only way the console
# is ever served as brotli. config.json is excluded because the entrypoint
# rewrites it at boot and a stale sibling would outrank the rewrite.
RUN apk add --no-cache brotli gzip && \
    find apps/console/dist -type f -size +1024c \
      \( -name '*.js' -o -name '*.css' -o -name '*.html' -o -name '*.json' \
         -o -name '*.svg' -o -name '*.ico' \) \
      ! -name 'config.json' \
      -exec brotli -q 11 -k {} \; \
      -exec gzip -9 -k {} \;

FROM alpine:3.24.1 AS production

# CVE-2026-14456: remove when alpine base image includes the fix
RUN apk upgrade --no-cache libssl3 libcrypto3

# Pinned exactly, not to a major: the header precedence ui/sws.toml relies on is
# undocumented, so a release that changed it would break caching silently.
ARG SWS_VERSION=2.44.0
ARG TARGETARCH

# The release CDN refuses or drops the connection often enough to fail a noticeable
# share of builds, and busybox wget retries nothing on its own, so retry here.
RUN set -eux; \
    case "$TARGETARCH" in \
      amd64) target=x86_64-unknown-linux-musl; \
             sha=804bc0c31c78385ac04e9a36f3c2aa3d3170eb77d66807c2a1660c56b2026bb1 ;; \
      arm64) target=aarch64-unknown-linux-musl; \
             sha=e9d6bd0b05dea441028eb4e690249ab585d711b311646bf9f420fcea01449670 ;; \
      *) echo "unsupported architecture: $TARGETARCH" >&2; exit 1 ;; \
    esac; \
    url="https://github.com/static-web-server/static-web-server/releases/download/v${SWS_VERSION}/static-web-server-v${SWS_VERSION}-${target}.tar.gz"; \
    downloaded=; \
    for attempt in 1 2 3 4 5; do \
      if wget -T 20 -qO /tmp/sws.tar.gz "$url"; then downloaded=1; break; fi; \
      echo "static-web-server download failed (attempt ${attempt}/5)" >&2; \
      sleep $(( attempt * 3 )); \
    done; \
    [ -n "$downloaded" ]; \
    echo "${sha}  /tmp/sws.tar.gz" | sha256sum -c -; \
    tar -xzf /tmp/sws.tar.gz -C /tmp; \
    mv "/tmp/static-web-server-v${SWS_VERSION}-${target}/static-web-server" /usr/local/bin/; \
    rm -rf /tmp/sws.tar.gz "/tmp/static-web-server-v${SWS_VERSION}-${target}"

RUN addgroup -g 1000 -S shellhub && adduser -u 1000 -S -G shellhub shellhub

COPY ui/sws.toml /etc/sws.toml
COPY ui/scripts /scripts

# Owned, not just readable: the entrypoint writes config.json in here.
COPY --from=builder --chown=1000:1000 /app/apps/console/dist /var/www

USER 1000:1000

EXPOSE 8080

CMD ["/scripts/entrypoint.sh"]
