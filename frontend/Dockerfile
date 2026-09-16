# Imagen de PRODUCCIÓN del frontend. Usa la salida "standalone" de
# Next.js (ya configurada en next.config.js desde la Fase 1): la imagen
# final solo lleva el servidor de Node mínimo necesario, no todo
# node_modules ni el código fuente.

FROM node:20-alpine AS deps
WORKDIR /app
COPY package.json ./
RUN npm install

FROM node:20-alpine AS build
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ENV NEXT_TELEMETRY_DISABLED=1
RUN npm run build

# --- Imagen final: solo la salida standalone ---
FROM node:20-alpine AS runner
WORKDIR /app

RUN addgroup -S app && adduser -S app -G app
ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=build /app/public ./public
COPY --from=build --chown=app:app /app/.next/standalone ./
COPY --from=build --chown=app:app /app/.next/static ./.next/static

USER app
EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD node -e "fetch('http://localhost:3000/').then(r=>process.exit(r.ok?0:1)).catch(()=>process.exit(1))"

CMD ["node", "server.js"]
