# syntax=docker/dockerfile:1.7
# ── Build stage ───────────────────────────────────────────────
FROM node:20-alpine AS builder
WORKDIR /app

COPY package.json package-lock.json ./
# --mount=type=cache персистит npm-кэш между сборками.
# При clean rebuild пакеты берутся из локального кэша Docker, а не скачиваются заново.
RUN --mount=type=cache,target=/root/.npm \
    npm ci \
    --registry=https://registry.npmmirror.com \
    --no-audit --no-fund \
    --fetch-retries=5 --fetch-retry-maxtimeout=120000

COPY vite.config.js index.html ./
COPY src ./src

RUN npm run build

# ── Runtime stage ─────────────────────────────────────────────
FROM nginx:1.27-alpine
RUN rm -rf /usr/share/nginx/html/*
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
