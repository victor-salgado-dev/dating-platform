-- 000023_add_profile_interests_key_idx.up.sql
--
-- Los filtros de búsqueda por interés (EXISTS sobre profile_interests con
-- interest_key = ...) solo podían empezar por el perfil: recorrer perfiles y
-- comprobar en cada uno si tiene el interés. Con este índice Postgres puede
-- empezar por el interés (unos cientos de filas) y llegar a los perfiles desde
-- ahí. INCLUDE (level) cubre también los filtros por rango de nivel.
--
-- En producción con la tabla ya grande, crea este índice a mano con
-- CREATE INDEX CONCURRENTLY (no puede ir dentro de una migración transaccional).

CREATE INDEX IF NOT EXISTS profile_interests_key_profile_idx
    ON profile_interests (interest_key, profile_id) INCLUDE (level);

-- Redundante: la clave primaria (profile_id, interest_key) ya sirve para buscar
-- por profile_id, y un índice menos abarata cada escritura.
DROP INDEX IF EXISTS profile_interests_profile_id_idx;

ANALYZE profile_interests;
