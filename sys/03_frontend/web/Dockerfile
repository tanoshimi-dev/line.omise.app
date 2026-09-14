# syntax=docker/dockerfile:1

FROM node:22-alpine AS deps
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci

FROM node:22-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
# Public environment variables are compiled into the browser bundle by
# `next build`; a container runtime env_file cannot change these links.
ARG NEXT_PUBLIC_API_URL=http://localhost:8080
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build

FROM node:22-alpine AS run
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -S nodejs && adduser -S nextjs -G nodejs

# `output: 'standalone'` (next.config.ts) traces only the files needed at
# runtime, so node_modules doesn't need to be copied/installed here.
COPY --from=build /app/public ./public
COPY --from=build --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=build --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs
EXPOSE 3000
ENV PORT=3000
# Docker sets HOSTNAME to the container ID. Next's standalone server would
# bind only to the IP that name resolves to, which breaks access from the
# other interface when the container joins both internal and Traefik networks.
ENV HOSTNAME=0.0.0.0
CMD ["node", "server.js"]
