-- 000022_add_seeking_genders.down.sql
DROP INDEX IF EXISTS profiles_seeking_genders_idx;
ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_seeking_genders_check;
ALTER TABLE profiles DROP COLUMN IF EXISTS seeking_genders;
