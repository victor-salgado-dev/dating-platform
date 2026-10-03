# Cambios - entrega 4

Extrae encima de `dating-platform\` (ruta con prefijo `backend\`).

## backend/internal/profiles
- `interest.go`  Se eliminan `MinHobbyIntensity` y `MaxHobbyIntensity` ("compatibilidad temporal con search").
  Comprobado con `Select-String` sobre todo `backend`: solo existian en este fichero y nadie las usa.
  Ademas el fichero termina ahora con salto de linea (gofmt).

Sin cambios de comportamiento. Verificado: el modulo de prueba compila, `go vet` pasa (con y sin tag `integration`).
