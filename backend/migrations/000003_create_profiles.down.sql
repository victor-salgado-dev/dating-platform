-- 000003_create_profiles.down.sql

DROP TRIGGER IF EXISTS profiles_set_updated_at ON profiles;
DROP TABLE IF EXISTS profiles;
