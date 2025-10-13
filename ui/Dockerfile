FROM node:24.10.0-alpine3.22 AS base

ARG NPM_CONFIG_REGISTRY

RUN apk add --update build-base python3 curl git

WORKDIR /app

COPY ui/package*.json ./

RUN npm install

FROM base AS development

ARG NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_REGISTRY ${NPM_CONFIG_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

FROM base AS builder

ARG NPM_CONFIG_REGISTRY

WORKDIR /app

COPY ui/. .

COPY --from=base /app/node_modules ./node_modules

RUN npm run build

FROM nginx:1.29.2-alpine AS production

RUN apk add curl

RUN rm /etc/nginx/conf.d/default.conf
COPY ui/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/dist /usr/share/nginx/html

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
