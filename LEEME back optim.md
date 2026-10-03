# Cambios en el backend (bloques 1-4)

Extrae este zip **encima de la carpeta del proyecto** (`dating-platform\`): las rutas
ya llevan el prefijo `backend\`. Haz un commit o una rama antes, por si quieres comparar.

## 1. Ficheros a BORRAR a mano (un zip no puede borrar)

- `backend\internal\profiles\visit_skip.go`   (sustituido por `FullPublicOptions`)
- `backend\internal\profiles\service_fotos_go.txt`   (ya estaba integrado en service.go; si aún existe)

## 2. Qué incluye

**backend/cmd/api/main.go** - `SetInteractionDeps` sin `blockingService`; nuevo `SetIDResolver`.
**backend/internal/server/routes.go** - `/profiles/{profileID}/full` sin el middleware `SkipVisitFromQuery`.

**backend/internal/profiles/** (nuevos marcados con *)
- Bloque 1: `postgres_repository.go`, `repository.go`, `service.go`, `handler.go`
  (limite de fotos atomico, limpieza con contexto desacoplado, 404 en foto publica, logs de 500, `updated_at`).
- Bloque 2: `id_resolver.go`, `catalog_cache.go`*, `service.go`, `postgres_repository.go`, `repository.go`
  (IDResolver en el service, IsVisible, GetPublicPhoto, regla de visibilidad unica, cache de catalogo).
- Bloque 3: `field.go`*, `handler_util.go`*, `patch.go`, `profile.go`, `partner_preferences.go`
  (Field[T], ProfileDetails, helpers del handler, SQL dinamico por reflexion).
- Bloque 4: `photo_url.go`*, `account_summary.go`, `handler.go`, `service.go`
  (URLs de foto en un solo sitio, avatar con miniatura, criterio de nationality, FullPublicOptions).
- Tests: `errors_test.go`*, `catalog_cache_test.go`*, `patch_test.go`*, `service_profile_test.go`*,
  `repo_sql_test.go`*, `photo_url_test.go`*, `visit_test.go`*, y con tag `integration`:
  `photo_limit_integration_test.go`*, `visibility_integration_test.go`*, `postgres_repository_test.go` (adaptado).

## 3. Cambios de API que pueden romper OTROS paquetes o tests

- `Repository`: `AddPhoto(..., maxPhotos)`, y nuevos `IsVisible` y `GetPublicPhoto` -> actualiza mocks.
- `Service.CreateProfile(ctx, userID, *Profile)`; `CreateProfileInput` ya no existe.
- `ProfilePatch` / `PartnerPreferencesPatch` usan `Field[T]` (`FieldOf(...)` para construirlos); no hay campos `...Set`.
- `Profile` embebe `ProfileDetails`: un literal con campos opcionales se escribe `Profile{ProfileDetails: ProfileDetails{...}}`.
- `Service.SetInteractionDeps(favs, likes, visits)`: sin `BlockChecker` (la interfaz se elimina).
- `Service.GetFullPublicProfile(ctx, viewer, profile, FullPublicOptions)`.
- `profiles.SkipVisitFromQuery` ya no existe.

## 4. Despues de extraer

    go build ./...
    go vet ./...
    go test ./...
    go test -tags=integration ./...      (necesita Postgres y Redis)

## 5. Que esta verificado y que no

Verificado: el paquete `profiles` compila, pasa `go vet` (con y sin tag `integration`) y sus 31 tests
(con `-race`), usando stubs para pgx y para auth/apperr/httpx/storage.
NO verificado: `main.go` y `routes.go` (dependen de todo el proyecto), los tests de integracion
(no hay Postgres), ni el comportamiento con tu `apperr`/`httpx` reales.
Los ficheros parten de las versiones que me pasaste: si has editado alguno desde entonces, compara antes de sobrescribir.
