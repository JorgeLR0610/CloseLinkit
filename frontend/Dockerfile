ARG NODE_VERSION=24.18-alpine3.23
ARG NGINX_VERSION=1.30.5-alpine3.24

# Build stage
FROM node:${NODE_VERSION} AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginxinc/nginx-unprivileged:${NGINX_VERSION}
COPY default.conf.template /etc/nginx/templates/default.conf.template
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 8080
