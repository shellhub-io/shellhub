FROM node:16.9.1-alpine3.11 as openapi

RUN npm install --global @redocly/openapi-cli

WORKDIR /app

RUN mkdir server

COPY openapi/server/package*.json ./server/

RUN cd server && npm install

COPY openapi/. ./

ENTRYPOINT ["/app/entrypoint.sh"]
