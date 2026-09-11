-- 000003_create_profiles.up.sql
--
-- Crea la tabla `profiles`: datos de perfil público/estructurado,
-- separados de `users` (cuenta). Relación 1:1 con users.
--
-- REGLA DE LOS DATOS FALTANTES (sección 8 del documento de proyecto):
-- los campos que el usuario no ha rellenado deben quedar en NULL, nunca
-- con un valor por defecto que simule una respuesta. Por eso casi todos
-- los campos opcionales no tienen DEFAULT y aceptan NULL explícitamente:
-- la Fase 5 (búsqueda) debe poder distinguir "no lo sé" de "no/false".

CREATE TABLE profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Campos obligatorios para poder usar la plataforma.
    display_name        TEXT NOT NULL,
    birth_date          DATE NOT NULL,
    gender              TEXT NOT NULL,
    country_code        TEXT NOT NULL,

    -- Campos opcionales: NULL = el usuario no lo ha indicado.
    region              TEXT,
    languages           TEXT[],
    relationship_goal   TEXT,
    has_children        BOOLEAN,
    wants_children      BOOLEAN,
    bio                 TEXT,
    interests           TEXT[],

    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT profiles_gender_check
        CHECK (gender IN ('female', 'male', 'non_binary', 'other')),

    CONSTRAINT profiles_relationship_goal_check
        CHECK (relationship_goal IS NULL OR relationship_goal IN
            ('casual', 'long_term', 'friendship', 'marriage', 'not_sure')),

    CONSTRAINT profiles_country_code_check
        CHECK (country_code ~ '^[A-Z]{2}$'),

    -- V1 es solo para mayores de 18 años (sección 1).
    CONSTRAINT profiles_birth_date_adult_check
        CHECK (birth_date <= (CURRENT_DATE - INTERVAL '18 years'))
);

CREATE INDEX profiles_country_code_idx ON profiles (country_code);
CREATE INDEX profiles_gender_idx ON profiles (gender);
CREATE INDEX profiles_birth_date_idx ON profiles (birth_date);

CREATE TRIGGER profiles_set_updated_at
    BEFORE UPDATE ON profiles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
