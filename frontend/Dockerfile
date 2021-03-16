FROM node:alpine AS builder
WORKDIR /app
COPY package.json .
COPY package-lock.json .
RUN npm ci
COPY . .
RUN npm test

FROM builder AS dev
CMD npm run serve -- --port ${PORT}

FROM builder AS dist
ARG VUE_APP_API_VERSION
RUN npm run build

FROM scratch AS prod
COPY --from=dist /app/dist .
