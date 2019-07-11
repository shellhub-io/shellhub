FROM node:11.9.0-alpine as base

WORKDIR /app

COPY package*.json ./

RUN npm install

FROM base as development

WORKDIR /src

COPY --from=base /app/node_modules /node_modules

ADD entrypoint-dev.sh /entrypoint.sh

CMD ["/entrypoint.sh"]

FROM base as production

RUN npm install -g http-server python2

COPY . .

COPY --from=base /app/node_modules ./node_modules

RUN npm run build

CMD [ "http-server", "-c-1", "-P", "http://localhost:8080?", "dist" ]
