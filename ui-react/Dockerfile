FROM node:24.13.0-alpine3.22 AS base

ARG NPM_CONFIG_REGISTRY

RUN apk add --update build-base python3 curl git

WORKDIR /app

# Copy workspace package files for npm install
COPY ui-react/package.json ui-react/package-lock.json ./
COPY ui-react/packages/design-system/package.json ./packages/design-system/
COPY ui-react/apps/admin/package.json ./apps/admin/

RUN npm install

FROM base AS development

ARG NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_REGISTRY ${NPM_CONFIG_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui-react/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

FROM base AS builder

ARG NPM_CONFIG_REGISTRY

WORKDIR /app

COPY ui-react/. .

COPY --from=base /app/node_modules ./node_modules

RUN npm run build

FROM nginx:1.29.4-alpine AS production

RUN apk add curl

RUN rm /etc/nginx/conf.d/default.conf
COPY ui-react/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/apps/admin/dist /usr/share/nginx/html/ui

COPY ui-react/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
