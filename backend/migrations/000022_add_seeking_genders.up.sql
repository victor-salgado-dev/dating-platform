-- 000022_add_seeking_genders.up.sql
--
-- Género(s) que busca el usuario. Obligatorio, vive en `profiles` junto a
-- `gender` (no en profile_partner_preferences). Reutiliza los valores que
-- ya existen en profiles.gender (female, male, non_binary, other): no se
-- añade ningún valor nuevo. El CHECK de profiles_gender_check no cambia.

ALTER TABLE profiles ADD COLUMN seeking_genders TEXT[];

-- Perfiles existentes: sin restricción hasta que cada usuario lo edite.
UPDATE profiles SET seeking_genders = ARRAY['female', 'male', 'non_binary', 'other'];

ALTER TABLE profiles ALTER COLUMN seeking_genders SET NOT NULL;

ALTER TABLE profiles ADD CONSTRAINT profiles_seeking_genders_check CHECK (
    cardinality(seeking_genders) BETWEEN 1 AND 4
    AND seeking_genders <@ ARRAY['female', 'male', 'non_binary', 'other']::TEXT[]
);

-- Acelera "el género del perfil está en lo que busca X" si luego se hace mutuo.
CREATE INDEX profiles_seeking_genders_idx ON profiles USING GIN (seeking_genders);
