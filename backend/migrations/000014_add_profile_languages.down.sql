-- 000014_add_profile_languages.down.sql
--
-- Nota: no se recuperan los datos. Al pasar de un array plano a filas
-- con nivel no hay forma de reconstruir el array original tal cual
-- era, así que profiles.languages vuelve vacío para todos.

ALTER TABLE profiles ADD COLUMN languages TEXT[];

DROP TRIGGER IF EXISTS profile_languages_set_updated_at ON profile_languages;
DROP TABLE IF EXISTS profile_languages;
