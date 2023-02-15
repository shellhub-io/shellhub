FROM node:16.13.1-alpine3.13 as base

ARG NPM_CONFIG_REGISTRY

RUN apk add --update build-base python3

WORKDIR /app

COPY new-ui/package*.json ./

RUN npm install

FROM base as development

ARG NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_REGISTRY ${NPM_CONFIG_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY new-ui/scripts /scripts

RUN apk add openjdk11-jre

RUN npm install -g @openapitools/openapi-generator-cli

RUN openapi-generator-cli version-manager set 6.0.0

CMD ["/scripts/entrypoint-dev.sh"]

FROM base as builder

ARG NPM_CONFIG_REGISTRY

WORKDIR /app

COPY new-ui/. .

COPY --from=base /app/node_modules ./node_modules

RUN npm run build

FROM nginx:1.23.0-alpine as production

RUN rm /etc/nginx/conf.d/default.conf
COPY new-ui/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/dist /usr/share/nginx/html

COPY new-ui/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
