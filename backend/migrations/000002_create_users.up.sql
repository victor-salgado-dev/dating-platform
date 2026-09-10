-- 000002_create_users.up.sql
--
-- Crea la tabla `users`, que representa la CUENTA (no el perfil público).
-- Los datos de perfil (display name, edad, país, bio, etc.) vivirán en una
-- tabla `profiles` separada a partir de la Fase 4, enlazada 1:1 con users.
--
-- También crea una función `set_updated_at()` reutilizable: la usará esta
-- tabla y, previsiblemente, la mayoría de tablas futuras que necesiten
-- mantener `updated_at` sin depender de que el código de aplicación
-- recuerde actualizarlo en cada UPDATE.

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE users (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email              TEXT NOT NULL,
    password_hash      TEXT NOT NULL,
    status             TEXT NOT NULL DEFAULT 'active',
    email_verified_at  TIMESTAMPTZ,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at         TIMESTAMPTZ,

    CONSTRAINT users_status_check CHECK (status IN ('active', 'suspended', 'deleted'))
);

-- Unicidad de email solo entre cuentas no eliminadas: permite que un email
-- se vuelva a registrar después de una eliminación de cuenta sin dejar
-- restricciones permanentes ligadas a una fila borrada lógicamente.
CREATE UNIQUE INDEX users_email_unique_active_idx
    ON users (lower(email))
    WHERE deleted_at IS NULL;

-- Búsquedas administrativas por estado (Fase 10 - moderación).
CREATE INDEX users_status_idx ON users (status) WHERE deleted_at IS NULL;

CREATE TRIGGER users_set_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
