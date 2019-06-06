FROM node:11.9.0-alpine as base

WORKDIR /app

COPY package*.json ./

RUN npm install

FROM base as development

WORKDIR /src

COPY --from=base /app/node_modules ./node_modules

CMD ["yarn", "serve"]

FROM base as production

RUN npm install -g http-server python2

COPY . .

RUN npm run build

CMD [ "http-server", "-c-1", "dist" ]
