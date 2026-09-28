-- 000019_add_new_members_activity.sql
--
-- La sección "new members" necesita:
-- 1. Saber cuándo fue la última actividad de un usuario para poder
--    mostrar (o no) el estado "en línea" / "visto recientemente".
--    La tabla `users` ya tiene `created_at`, así que el tiempo de
--    registro ya estaba cubierto; falta `last_active_at`.
-- 2. Un índice que acelere la consulta principal de la sección:
--    usuarios activos (deleted_at IS NULL) ordenados por fecha de
--    registro descendente.

ALTER TABLE users
    ADD COLUMN last_active_at TIMESTAMPTZ;

CREATE INDEX users_created_at_for_new_members_idx
    ON users (created_at DESC)
    WHERE deleted_at IS NULL;
