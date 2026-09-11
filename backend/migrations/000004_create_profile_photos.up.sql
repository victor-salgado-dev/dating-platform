-- 000004_create_profile_photos.up.sql
--
-- Fotos de perfil. `storage_key` es una clave opaca resuelta por la
-- abstracción de storage (internal/storage): en desarrollo apunta a un
-- fichero local, en producción podrá apuntar a un objeto S3-compatible,
-- sin que esta tabla ni el resto del dominio tengan que cambiar.

CREATE TABLE profile_photos (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    profile_id   UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    storage_key  TEXT NOT NULL,
    content_type TEXT NOT NULL,
    position     INTEGER NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Acelera "dame las fotos de este perfil en orden".
CREATE INDEX profile_photos_profile_id_position_idx
    ON profile_photos (profile_id, position);
