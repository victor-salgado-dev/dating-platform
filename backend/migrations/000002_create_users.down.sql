-- 000002_create_users.down.sql

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
DROP TABLE IF EXISTS users;

-- Las migraciones se revierten siempre en orden inverso (la más reciente
-- primero), así que al llegar aquí ninguna tabla posterior que dependiera
-- de esta función sigue existiendo: es seguro eliminarla.
DROP FUNCTION IF EXISTS set_updated_at();
