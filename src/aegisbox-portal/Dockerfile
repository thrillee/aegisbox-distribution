FROM node:20-alpine AS base

# Install dependencies only when needed
FROM base AS deps
RUN apk add --no-cache libc6-compat curl
RUN corepack enable pnpm && corepack prepare pnpm@10.28.0 --activate
WORKDIR /app

# pnpm is pinned to the version that wrote pnpm-lock.yaml: newer majors fail the install on
# unapproved dependency build scripts. package.json "pnpm" lists the packages allowed to run
# them (esbuild and sharp need theirs; @clerk/shared's is only a notice).
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

# Rebuild the source code only when needed
FROM base AS builder
RUN apk add --no-cache libc6-compat curl
RUN corepack enable pnpm && corepack prepare pnpm@10.28.0 --activate
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_* variables must be set at build time (inlined into client JS by Next.js)
ARG NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY
ENV NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY=${NEXT_PUBLIC_CLERK_PUBLISHABLE_KEY}
ARG NEXT_PUBLIC_CLERK_ADMIN_ORG_ID
ENV NEXT_PUBLIC_CLERK_ADMIN_ORG_ID=${NEXT_PUBLIC_CLERK_ADMIN_ORG_ID}
ARG NEXT_PUBLIC_APP_URL
ENV NEXT_PUBLIC_APP_URL=${NEXT_PUBLIC_APP_URL}

# Use placeholder for DATABASE_URL at build time - actual value comes from runtime env
ENV DATABASE_URL="postgresql://placeholder:placeholder@localhost:5432/placeholder"

RUN pnpm run build

# Production image, copy all the files and run next
FROM base AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
# Runtime environment variables will be provided via docker-compose or env_file
# DATABASE_URL, CLERK_SECRET_KEY, CLERK_ADMIN_ORG_ID are set at runtime

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

COPY --from=builder /app/public ./

RUN mkdir .next
RUN chown nextjs:nodejs .next

COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000

ENV PORT=3000
ENV HOSTNAME="0.0.0.0"

CMD ["node", "server.js"]
