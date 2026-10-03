# Cambios - entrega 3 (paquete search)

Extrae encima de `dating-platform\` (las rutas llevan el prefijo `backend\`). Solo incluye lo que cambia
respecto a las entregas anteriores. No hay nada que borrar.

## backend/internal/search
- `query.go` (nuevo)  Constructor de la consulta SQL, puro y sin base de datos. Sustituye a los ~45 `if` casi
  identicos de `Search`. El SQL generado es IDENTICO al anterior salvo lo indicado abajo.
- `postgres_repository.go`  `Search` solo ejecuta y escanea (381 -> 92 lineas).
- `service.go`  `buildParams` dividido en funciones por seccion, con el mismo orden de validacion.
- `types.go`  `MinInterestLevel/MaxInterestLevel` salen de `profiles` (antes duplicados "por suposicion");
  nuevas constantes `MaxPage` y `MaxInterestFilters`.
- `handler.go`  URL de foto desde `profiles.PublicPhotoURL`; campo nuevo `photo_thumb_url`.
- `integration_test.go`  El literal `profiles.Profile{WantsChildren: ...}` ya no compilaba desde la entrega 1
  (`ProfileDetails`); corregido.
- Tests nuevos: `query_test.go`, `service_limits_test.go`, `handler_test.go`.

## backend/internal/profiles (los minimos que search necesita)
- `postgres_repository.go`  Nueva `VisibleSQL(viewerParam)`: la regla de visibilidad, exportada para que search
  use la MISMA definicion que el resto (antes tenia su propia copia con el `OR`).
- `photo_url.go`  Nueva `PublicPhotoURL`. 
- `repo_sql_test.go`, `photo_url_test.go`  Tests de lo anterior.

## Cambios de comportamiento (todos intencionados)
1. Visibilidad: `NOT (A OR B)` pasa a dos `NOT EXISTS` (equivalente; permite usar el indice en cada sentido).
2. Edad: los limites se calculan sin `AddDate`. Solo difiere cuando "hoy" es 29 de febrero.
3. `page` > 10000 o mas de 25 filtros de intereses -> 400 (antes: OFFSET desbordado/500, o una subconsulta por interes).
4. `nationality` en blanco ya no filtra (antes filtraba por la cadena vacia y no devolvia nada).
5. Con varios limites de interes invalidos se informa siempre del mismo (antes, de uno al azar). Un rasgo de
   personalidad con espacios en la clave conserva sus limites.
6. `photo_thumb_url` es un campo NUEVO (aditivo); `photo_url` no cambia. El cliente puede usar la miniatura en las tarjetas.

## Despues de extraer
    go build ./... && go vet ./... && go test ./...
    go test -tags=integration ./internal/search/ ./internal/profiles/     (necesita Postgres)

## Verificado / no verificado
Verificado: 37 tests en search y 52 en profiles, con -race. Ademas, la consulta nueva se comparo con la ORIGINAL
sobre 30.000 combinaciones aleatorias de filtros (SQL, argumentos y recuento identicos), y buildParams con 60.000
peticiones aleatorias (mismos resultados y mensajes de error). Esos comparadores no se entregan.
NO verificado: el SQL contra Postgres real ni su rendimiento (no hay base de datos en mi entorno).
