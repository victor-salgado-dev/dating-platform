-- 000023_add_profile_interests_key_idx.down.sql
CREATE INDEX IF NOT EXISTS profile_interests_profile_id_idx ON profile_interests (profile_id);
DROP INDEX IF EXISTS profile_interests_key_profile_idx;
