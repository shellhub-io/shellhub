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
ENV NPM_CONFIG_REGISTRY ${NPM_CONFIG_REGISTRY}

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

FROM nginx:1.31.3-alpine AS production

RUN apk add curl

RUN rm /etc/nginx/conf.d/default.conf
COPY ui/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/apps/console/dist /usr/share/nginx/html/ui

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
