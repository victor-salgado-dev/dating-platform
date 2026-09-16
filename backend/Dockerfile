# Imagen de PRODUCCIÓN del backend. A diferencia de Dockerfile.dev:
# - Compila un binario en el build, no usa `go run` ni monta el código
#   como volumen (la imagen final es autocontenida).
# - No incluye herramientas de desarrollo ni el código fuente de Go.
# - Corre como usuario sin privilegios.

FROM golang:1.22-alpine AS build

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .

# CGO_ENABLED=0: binario estático, la imagen final no necesita libc.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
    -o /bin/promote-admin ./cmd/promote-admin

# --- Imagen final: solo el binario, sin toolchain de Go ---
FROM alpine:3.20

RUN apk add --no-cache ca-certificates curl && \
    addgroup -S app && adduser -S app -G app

COPY --from=build /bin/api /usr/local/bin/api
COPY --from=build /bin/promote-admin /usr/local/bin/promote-admin

USER app
EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -fsS http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/usr/local/bin/api"]
