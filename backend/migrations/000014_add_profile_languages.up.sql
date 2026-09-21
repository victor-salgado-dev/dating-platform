-- 000014_add_profile_languages.up.sql
--
-- Sustituye profiles.languages (TEXT[] libre, sin nivel) por una tabla
-- 1:N con nivel, mismo patrón que profile_hobbies/profile_personality_
-- answers: separar "qué" de "cuánto". La lista de códigos permitidos es
-- la misma que ya validaba el array (migración 000012), así que no se
-- pierde ninguna restricción, solo se le añade el nivel.
--
-- No hace falta una tabla de catálogo aparte para los idiomas: a
-- diferencia de los hobbies o las afirmaciones de personalidad, esta
-- lista es cerrada y estable (códigos ISO), así que un CHECK inline es
-- más simple y no necesita JOIN para validar.

CREATE TABLE profile_languages (
    profile_id     UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    language_code  TEXT NOT NULL CHECK (language_code IN (
        'en', 'tl', 'ceb', 'hy', 'ar', 'es', 'ja', 'af', 'sq', 'am', 'syr', 'az',
        'id', 'ms', 'be', 'bn', 'ber', 'bg', 'my', 'zh_yue', 'zh_cmn', 'cr', 'hr',
        'cs', 'da', 'nl', 'ti', 'et', 'fa', 'fi', 'fr', 'ka', 'de', 'el', 'gu',
        'ha', 'he', 'hi', 'hu', 'is', 'ilo', 'iu', 'it', 'kk', 'km', 'ky', 'lo',
        'lv', 'lt', 'mk', 'mg', 'ml', 'dv', 'mt', 'mr', 'mn', 'ne', 'no', 'ps',
        'pcm', 'pl', 'pt', 'qu', 'ro', 'ru', 'sr', 'sd', 'si', 'sk', 'sl', 'so',
        'sw', 'sv', 'ta', 'te', 'th', 'bo', 'to', 'tr', 'tk', 'uga', 'uk', 'ur',
        'uz', 'vi', 'cy', 'other'
    )),
    level      SMALLINT CHECK (level BETWEEN 1 AND 5),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (profile_id, language_code)
);

CREATE INDEX profile_languages_profile_id_idx ON profile_languages (profile_id);

CREATE TRIGGER profile_languages_set_updated_at
    BEFORE UPDATE ON profile_languages
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

ALTER TABLE profiles DROP COLUMN languages;
