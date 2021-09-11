FROM node:12.16.0-alpine as base

ARG NPM_CONFIG_REGISTRY

RUN apk add --update build-base python

WORKDIR /app

COPY ui/package*.json ./

RUN npm install

FROM base as development

ARG NPM_CONFIG_REGISTRY
ENV NPM_CONFIG_REGISTRY ${NPM_CONFIG_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

FROM base as builder

ARG NPM_CONFIG_REGISTRY

WORKDIR /app

COPY ui/. .

COPY --from=base /app/node_modules ./node_modules

RUN npm run build

FROM nginx:1.21.3-alpine as production

RUN rm /etc/nginx/conf.d/default.conf
COPY ui/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/dist /usr/share/nginx/html

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
