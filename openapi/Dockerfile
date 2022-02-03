FROM node:16.9.1-alpine3.11 as openapi

RUN npm install --global @redocly/openapi-cli

RUN apk add inotify-tools

RUN mkdir /openapi

COPY ./openapi/server/package*.json /openapi/server/

RUN cd /openapi/server && npm install

COPY ./openapi/. /openapi

WORKDIR /openapi

ENTRYPOINT ["/openapi/entrypoint.sh"]
