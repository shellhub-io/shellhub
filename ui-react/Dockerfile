FROM node:24.13.0-alpine3.22 AS base

ARG NPM_CONFIG_REGISTRY

RUN apk add --update build-base python3 curl git

WORKDIR /app

# Copy workspace package files for npm install
COPY ui-react/package.json ui-react/package-lock.json ./
COPY ui-react/packages/design-system/package.json ./packages/design-system/
COPY ui-react/apps/console/package.json ./apps/console/

RUN npm install

FROM base AS development

ARG NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_REGISTRY ${NPM_CONFIG_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui-react/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

# Bundle OpenAPI spec from YAML sources
FROM node:24.13.0-alpine3.22 AS spec-bundler
RUN npm install -g @redocly/cli@1.0.0-beta.100
WORKDIR /spec
COPY openapi/spec ./spec
COPY openapi/redocly.yaml .
RUN redocly bundle spec/openapi.yaml -o openapi.json

FROM base AS builder

ARG NPM_CONFIG_REGISTRY

WORKDIR /app

COPY ui-react/. .

COPY --from=base /app/node_modules ./node_modules
COPY --from=spec-bundler /spec/openapi.json ./apps/console/openapi.json

RUN OPENAPI_SPEC_PATH=./openapi.json npm run generate -w @shellhub/console
RUN npm run build

FROM nginx:1.29.4-alpine AS production

RUN apk add curl

RUN rm /etc/nginx/conf.d/default.conf
COPY ui-react/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/apps/console/dist /usr/share/nginx/html/ui

COPY ui-react/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
