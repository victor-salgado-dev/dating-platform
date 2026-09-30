# Fase 2: backend ligero + flags por ficha + header en una petición

Estructura del zip: `backend/` (Go) y `frontend/` (Next). Copia cada archivo a su ruta equivalente.
Las rutas de `cmd/api/main.go` y `internal/server/routes.go` son las que asumí; ajústalas si difieren.

## 1. Pasos MANUALES (obligatorios)

1. **Migración**: aplica `backend/migrations/000020_perf_indexes_and_popularity.up.sql`.
2. **`internal/activity/types.go`** (no lo tenía): añade a `Item` estos cuatro campos, si no, no compila:
   ```go
   Liked            bool
   Favorited        bool
   ReceivedLike     bool
   ReceivedFavorite bool
   ```
3. **Frontend**: `lib/api.ts` sustituye al tuyo (incluye el parche anterior y los tipos nuevos:
   `LikesResponse`, `FavoritesResponse`, `VisitsResponse`, `MatchesResponse`, `ActivityResponse`, `MeResponse`...).
4. `lib/useProfileInteractions.ts` YA NO lo usa nada de este zip. No lo borro porque puede haber pantallas
   que no vi (matches, etc.). Haz `grep -r useProfileInteractions src app` y migra lo que salga (quitar las
   props initialLiked/initialFavorited/receivedLike/receivedFavorite/onToggle*, ProfileCard ya lee los flags
   de la ficha); cuando no quede ningún import, borra el archivo.

## 2. Qué cambia en el backend

- **search**: los 4 endpoints devuelven el mismo shape completo (con `relationship_goals` y los flags
  `liked`, `favorited`, `received_like`, `received_favorite`). Un solo helper reemplaza a 3 handlers.
  Popularidad = vista materializada `profile_popularity` (últimos 30 días), refrescada cada 5 min
  (`search.StartPopularityRefresher`, con advisory lock). Página fuera de rango ya devuelve `total` correcto.
  `online-now` funciona: `last_active_at` ahora se escribe (ver auth).
- **favorites / likes / visits / activity**: cada ficha trae sus 4 flags (EXISTS por índice único);
  `profiles.IDResolver` sustituye a `GetByUserID` + `GetPublicByID` (60 columnas) por una consulta corta
  y caché en memoria del profile_id. `POST /likes/{id}` devuelve `200 {"matched": bool}` (antes 204).
  visits: `page_size` acotado (antes se podía pedir todo).
- **auth**: `RequireAuth` pasa de {Redis GET+TTL + fila completa de `users`} a **un** viaje a Redis (script Lua)
  que valida sesión, cuenta bloqueada y decide si tocar `last_active_at` (como mucho 1 vez/2 min/usuario).
  Suspender/eliminar una cuenta la bloquea al instante con una marca en Redis (`admin.NewService` recibe el
  `sessionStore`; `main.go` rellena las marcas al arrancar para no reabrir sesiones antiguas).
  `GET /auth/me` devuelve además `photo_url`, `profile_completion`, `profile_id`, `has_profile`.
- **BUG corregido (visitas)**: `profiles.VisitRecorder` recibe IDs de PERFIL pero `main.go` le pasaba
  `visits.Service` (que espera un user_id): el registro desde `/full` fallaba siempre. Ahora se le pasa
  `visitsRepo`. Para no contar visitas fantasma al precargar en Quick Match: `GET /profiles/{id}/full?visit=0`
  no registra visita (`profiles.SkipVisitFromQuery` en routes.go + 1 línea en `profiles/service.go`).
- La migración recrea el trigger `users_set_updated_at` para que tocar `last_active_at` no cambie `updated_at`.

## 3. Qué cambia en el frontend

- `ProfileCard` lee los flags de la ficha; Home, Discover, Likes/Visits/Favorites, Activity y Quick Match
  ya no descargan 4 listas de 100 al arrancar.
- `InteractionListSection` pierde su caché propia (usaba marcas caducadas); vale la de `apiFetch`
  (60 s, se vacía en cada mutación).
- `useFullProfile` usa la caché de `apiFetch` (la propia nunca se invalidaba) y ya no se queda con el
  estado viejo de like/favorito. Quick Match precarga con `visit=0` y registra la visita real al mostrar la carta.
- `app/account-nav.tsx`: 1 petición (`/auth/me`) en vez de 3; el % lo calcula el servidor.
- Perfil público y Quick Match usan la respuesta de `POST /likes` (`matched`) en vez de pedir `/matches/{id}`.

## 4. Qué he probado y qué no

Probado: el backend compila (con stubs de pgx/redis y de los paquetes que no tengo); el frontend pasa `tsc`;
las migraciones 1-19 + la 20 se aplican en un PostgreSQL 16 real, y todas las consultas nuevas (search recent/
popular/online, favoritos ×3, likes, matches, visitas, actividad, ResolveTarget con bloqueo, resumen de cuenta,
trigger) devuelven lo esperado con datos de prueba; el script Lua se probó en un Redis real (6 escenarios).

NO probado: el arranque real de la app, `main.go`/`routes.go` completos (los revisé a mano), ni carga.
Los tests Go existentes no se han tocado y algunos dejarán de compilar: fakes de `users.Repository`
(falta `TouchLastActive`), y llamadas a `favorites/likes/visits/activity.NewService` y `admin.NewService`
(firmas nuevas). Los dejaste aparte; aviso para cuando los retomes.

## 5. Sobre el zip anterior (home-tabs-refactor)

Sigue vigente: redirects de `/new-members`, `/online-now`, `/popular`, `app/page.module.css` y los diccionarios.
De ese zip, este sustituye a `app/page.tsx`, `lib/api.ts` y `lib/useProfileList.ts` (ya sin el mapeo de la
respuesta reducida: el backend devuelve el mismo shape en los cuatro endpoints).
