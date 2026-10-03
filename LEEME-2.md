# Cambios - entrega 2

Extrae encima de `dating-platform\` (las rutas llevan el prefijo `backend\`). Solo incluye lo que
cambia respecto a la entrega anterior. No hay nada que borrar.

## backend/internal/profiles
- `imageproc.go`  (MODIFICADO respecto a tu original)  Procesado de fotos: se reduce ANTES de aplanar la
  transparencia y se aplana en el sitio. Una PNG transparente de 16 MP pasa de 245 MB a 115 MB de pico.
  Ademas la lista blanca JPEG/PNG/WebP se aplica aqui (imaging tambien decodifica GIF, BMP y TIFF).
- `imageproc_test.go`  (nuevo)
- `list_item.go`  (MODIFICADO respecto a tu original)  Scan{Base,}ListItem comparten una sola implementacion;
  `ViewerFlagsSQL` valida los alias y entra en panico si no son identificadores simples (o chocan con fl/ff/rl/rf).
- `list_item_test.go`  (nuevo)
- `id_resolver.go`  La cache expulsa UNA entrada al llenarse, en vez de vaciarse entera.
- `id_resolver_test.go`  (nuevo)
- `handler.go`  Solo se retira la nota de "RECONSTRUCCION" (ya validada en funcionamiento).

## backend/internal/auth   (seguridad)
- `service.go`
  - Login: con un email inexistente se hace una comparacion bcrypt ficticia. Antes respondia en ~0,5 us frente
    a ~67 ms con una contrasena incorrecta, y esa diferencia delata que emails estan registrados.
  - RequestPasswordReset: crear el token y enviar el correo pasa a segundo plano; antes la respuesta esperaba
    al SMTP solo si la cuenta existia. Si el proceso se apaga justo entonces el correo puede perderse.
- `service_enumeration_test.go`  (nuevo)  Asume que `email.Sender` tiene un unico metodo `Send(ctx, Message)`
  y que `users.Repository.GetByEmail` devuelve `(*users.User, error)`.

## Despues de extraer
    go build ./... && go vet ./... && go test ./...

## Verificado / no verificado
Verificado (con la libreria real `disintegration/imaging` y `x/crypto`, y stubs de lo demas): profiles 49 tests y
auth 13 tests, con `-race`. Los tests nuevos de memoria y de tiempos FALLAN contra el codigo anterior.
NO verificado: contra tu `users`/`email`/`apperr` reales, ni los tests de integracion (necesitan Postgres/Redis).
