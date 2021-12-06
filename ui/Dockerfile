FROM node:15.14.0-alpine3.13 as base

ARG YARN_REGISTRY

RUN apk add --update build-base python3

WORKDIR /app

COPY ui/package.json ui/yarn.lock ./

RUN yarn install

FROM base as development

ARG YARN_REGISTRY
ENV YARN_REGISTRY ${YARN_REGISTRY}

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint-dev.sh"]

FROM base as builder

ARG YARN_REGISTRY

WORKDIR /app

COPY ui/. .

COPY --from=base /app/node_modules ./node_modules

RUN yarn run build

FROM nginx:1.21.4-alpine as production

RUN rm /etc/nginx/conf.d/default.conf
COPY ui/nginx.conf /etc/nginx/conf.d

COPY --from=builder /app/dist /usr/share/nginx/html

COPY ui/scripts /scripts

CMD ["/scripts/entrypoint.sh"]
