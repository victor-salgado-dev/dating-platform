-- 000021_add_photo_thumbnails.down.sql
ALTER TABLE profile_photos DROP COLUMN IF EXISTS thumb_storage_key;
