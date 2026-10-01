-- 000021_add_photo_thumbnails.up.sql
--
-- Clave de storage de la miniatura de cada foto. Nullable a propósito: las
-- fotos subidas antes de esta migración no tienen miniatura y el endpoint
-- que sirve la foto cae a la original cuando está vacía.
ALTER TABLE profile_photos ADD COLUMN thumb_storage_key TEXT;
