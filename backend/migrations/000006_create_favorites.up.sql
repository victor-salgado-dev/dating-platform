-- 000006_create_favorites.up.sql
--
-- Favoritos: relación "a mí me interesa este perfil". Se referencia por
-- profile_id (no user_id) porque toda la API pública (Fases 5 y 6) ya
-- identifica a las personas por su profile_id y nunca expone user_id;
-- mantenerlo así evita tener que filtrar un dato nuevo al cliente.

CREATE TABLE favorites (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    favorite_profile_id UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (user_id, favorite_profile_id)
);

-- Acelera "dame mis favoritos, más recientes primero" (el listado).
CREATE INDEX favorites_user_id_created_at_idx ON favorites (user_id, created_at DESC);
