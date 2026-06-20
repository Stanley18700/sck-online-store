# syntax=docker/dockerfile:1.7
# Optimized store-web Dockerfile.
# Key wins vs. previous version:
#   1. Multi-stage build: deps layer is reusable, build layer uses it,
#      final runtime image only carries the standalone Next.js bundle.
#   2. `npm ci` instead of `npm install` (~30% faster, deterministic).
#   3. BuildKit cache mount for /root/.npm keeps the package cache
#      warm across builds.
#   4. Final image size drops from ~1.4GB -> ~150-200MB by using
#      `output: 'standalone'` in next.config.js.

ARG STORE_SERVICE_URL

# ---------- Stage 1: install dependencies ----------
FROM node:21-alpine AS deps
WORKDIR /app
# Only copy the npm lockfile; package-lock.json is the source of truth.
# yarn.lock is intentionally not used here.
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci --no-audit --no-fund --prefer-offline

# ---------- Stage 2: build ----------
FROM node:21-alpine AS builder
WORKDIR /app
ARG STORE_SERVICE_URL
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# Configure runtime URL if provided as build arg
RUN if [ -n "$STORE_SERVICE_URL" ]; then \
      echo "STORE_SERVICE_URL=${STORE_SERVICE_URL}" > .env; \
    fi

# next.config.js has `output: 'standalone'` so this produces a
# self-contained server in .next/standalone
RUN --mount=type=cache,target=/root/.npm \
    --mount=type=cache,target=/app/.next/cache \
    npm run build

# ---------- Stage 3: slim runtime ----------
FROM node:21-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000

# Non-root user for the runtime
RUN addgroup --system --gid 1001 nodejs && \
    adduser  --system --uid 1001 nextjs

# Copy only what's needed at runtime:
#   - standalone server bundle (has node_modules traced by Next)
#   - static assets
#   - public directory
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static
COPY --from=builder --chown=nextjs:nodejs /app/public ./public

USER nextjs
EXPOSE 3000

# Standalone output uses server.js as its entrypoint
CMD ["node", "server.js"]