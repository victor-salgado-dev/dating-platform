#!/bin/sh
# Se ejecuta en cada arranque del contenedor (no solo en el build), porque
# ./backend:/app se monta como volumen y tapa cualquier go.sum generado
# durante `docker build`. Sin este paso, el binario intentaba resolver
# módulos por red en cada arranque y un fallo de red dejaba el contenedor
# en bucle de reinicio.
set -e

echo "==> Sincronizando dependencias (go mod tidy)..."
go mod tidy

echo "==> Arrancando backend..."
exec go run ./cmd/api
