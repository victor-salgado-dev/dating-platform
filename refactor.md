# Refactor de unificación (Paso A) y verificación

Registro de los cambios realizados durante esta conversación sobre la base
de código de `dating-platform`. El objetivo era reducir duplicación y dejar
una única fuente de verdad para piezas que estaban copiadas en varios
dominios, **sin cambiar el contrato HTTP ni el esquema de base de datos**.

Este documento recoge: qué se cambió, por qué, cómo se verificó y qué queda
pendiente.

---

## Paso A — Unificación de paginación, errores y ficha base

**Commit:** `9798e12` — `refactor: unifica paginación, errores y ficha base de listados`

Se detectaron cuatro duplicaciones repartidas por los paquetes de dominio:

1. `totalPages(total, pageSize)` reimplementado en varios repositorios, con
   dos semánticas distintas en el borde (`total == 0` vs `total <= 0`).
2. El struct `ValidationError` declarado de forma idéntica en ~8 paquetes.
3. La ficha resumida de perfil (`ListItem`) copiada en `favorites`, `likes`,
   `visits` y `blocking`, idéntica salvo el nombre del timestamp.
4. El bucle de escaneo de fila (`for rows.Next() { rows.Scan(...) }`)
   duplicado en cada listado.

### Cambios por archivo

- `backend/internal/pagination/pagination.go`
  - Nuevo helper `TotalPages(total, pageSize int) int`, única fuente de
    verdad. Devuelve `0` si `total <= 0` o `pageSize <= 0` (evita división
    por cero).
  - Se mantienen las constantes `DefaultPageSize` y `MaxPageSize`.

- `backend/internal/apperr/errors.go`
  - Se añade el tipo `ValidationError` (único) y el helper
    `InvalidField(field, message string) error`.
  - `Error()` produce el mensaje `campo "<field>" inválido: <message>`.

- `backend/internal/profiles/list_item.go`
  - Nuevo struct `BaseListItem` con los campos comunes a todas las fichas:
    `ProfileID, DisplayName, Age, Gender, CountryCode, Region, HasPhoto`.
  - Nuevo helper `ScanBaseListItem(row pgx.Row, extras ...any)`, que escanea
    las 7 columnas base **en este orden exacto**:
    `id, display_name, birth_date, gender, country_code, region, has_photo`
    y calcula `Age` a partir de `birth_date`. Los campos extra (timestamp
    del listado, `COUNT(*) OVER()`, etc.) se pasan como punteros y se
    escanean a continuación, en el orden en que aparezcan en el `SELECT`.

- `backend/internal/profiles/errors.go`
  - `ValidationError` pasa a ser un **alias** de `apperr.ValidationError`
    (`type ValidationError = apperr.ValidationError`), de modo que
    `errors.As` y los handlers existentes siguen funcionando sin cambios.
  - `invalidField` delega en `apperr.InvalidField`.

- `backend/internal/favorites/types.go`
  - `ListItem` embebe `profiles.BaseListItem` y añade
    `RelationshipGoal` y `FavoritedAt`.
  - Las constantes `DefaultPageSize` / `MaxPageSize` apuntan a
    `pagination`.

- `backend/internal/favorites/postgres_repository.go`
  - `List` y `ListReceived` usan `profiles.ScanBaseListItem` y
    `pagination.TotalPages`. Se unifica el orden de columnas del `SELECT`.

- `backend/internal/likes/types.go`
  - `ListItem` y `MatchItem` embeben `profiles.BaseListItem`.
  - Constantes apuntan a `pagination`.

- `backend/internal/likes/postgres_repository.go`
  - `listLikes` y `ListMatches` usan `profiles.ScanBaseListItem` y
    `pagination.TotalPages`. Se unifica el orden de columnas.

- `backend/internal/visits/types.go`
  - `ListItem` embebe `profiles.BaseListItem`; constantes apuntan a
    `pagination`.

- `backend/internal/visits/postgres_repository.go`
  - `listVisits` usa `profiles.ScanBaseListItem` y `pagination.TotalPages`.

- `backend/internal/blocking/types.go`
  - `ListItem` embebe `profiles.BaseListItem`; constantes apuntan a
    `pagination`.

- `backend/internal/blocking/postgres_repository.go`
  - `List` usa `profiles.ScanBaseListItem` y `pagination.TotalPages`.
    **Nota:** su `SELECT` ahora incluye `has_photo`. Si el handler lo
    serializa, la respuesta de `/blocks` gana ese campo (aditivo).

### Avisos / riesgos de este paso

- El mensaje de `ValidationError.Error()` pierde el prefijo del dominio;
  donde el handler lo escribe tal cual, el texto que ve el usuario en un 400
  cambia.
- La unificación de `ValidationError` quedó **a medias**: solo `profiles`
  migró. `admin`, `contact`, `messaging`, `reports` y `search` mantienen
  todavía su copia.
- El compilador **no** detecta un desajuste en el orden de columnas del
  `SELECT` respecto a `ScanBaseListItem`. Por eso se añadió el test de
  integración descrito más abajo.

---

## Higiene — Normalización de finales de línea

**Commit:** `6680e2b` — `chore: forzar finales de línea LF con .gitattributes`

Durante la verificación, `gofmt -l .` marcaba 31 archivos. La causa era
CRLF **cometido** en el repositorio (con `core.autocrlf=false`, el blob
conserva los `\r`). Esto produce falsos positivos de `gofmt` en entornos
Linux (contenedor/CI).

- Se añade `.gitattributes` con `* text=auto eol=lf`.
- Se normalizaron los finales de línea del repositorio
  (`git add --renormalize .`).

---

## Verificación

### Paso 1 — Compilación y estáticos

