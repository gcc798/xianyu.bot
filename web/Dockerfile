FROM node:20-slim AS web-build
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable && corepack prepare pnpm@10.33.2 --activate
WORKDIR /app/web
COPY web/package.json ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile=false
COPY web/ ./
RUN pnpm run build

FROM node:20-slim AS agent-build
WORKDIR /app/agent-web
COPY agent-web/package.json agent-web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --prefer-offline --no-audit --no-fund
COPY agent-web/ ./
RUN npm run build

FROM nginx:alpine
LABEL MAINTAINER="bypanghu@163.com"
COPY web/.docker-compose/nginx/conf.d/my.conf /etc/nginx/conf.d/my.conf
COPY --from=web-build /app/web/dist /usr/share/nginx/html
COPY --from=agent-build /app/agent-web/dist /usr/share/nginx/html/agent
RUN ls -al /usr/share/nginx/html /usr/share/nginx/html/agent
