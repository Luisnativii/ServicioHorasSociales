FROM node:22.23.4-bookworm-slim AS dev
ARG APP_VERSION=0.0.1
ENV VITE_APP_VERSION=${APP_VERSION}
WORKDIR /workspace/apps/web
RUN chown node:node /workspace/apps/web
USER node
COPY --chown=node:node apps/web/package.json apps/web/package-lock.json ./
RUN npm ci --no-fund --no-audit
COPY --chown=node:node apps/web/ ./
EXPOSE 8080
CMD ["npm", "run", "dev"]

FROM dev AS build
RUN npm run build

FROM nginxinc/nginx-unprivileged:1.30.3-alpine AS runtime
ARG APP_VERSION=0.0.1
LABEL org.opencontainers.image.version=${APP_VERSION}
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf
COPY --from=build /workspace/apps/web/dist /usr/share/nginx/html
USER 101:101
EXPOSE 8080
