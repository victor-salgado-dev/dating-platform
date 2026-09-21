-- 000016_replace_hobbies_with_interests.down.sql
--
-- Nota: como en cualquier down de este tipo, se pierden los datos de
-- los perfiles (qué interés tenía marcado cada uno con qué nivel); se
-- restaura la estructura, no el contenido.

DROP TRIGGER IF EXISTS profile_interests_set_updated_at ON profile_interests;
DROP TABLE IF EXISTS profile_interests;
DROP TABLE IF EXISTS interests;

ALTER TABLE profiles ADD COLUMN interests TEXT[];

CREATE TABLE hobby_definitions (
    key         TEXT PRIMARY KEY,
    category    TEXT NOT NULL,
    label       TEXT NOT NULL,
    sort_order  SMALLINT NOT NULL DEFAULT 0
);

INSERT INTO hobby_definitions (key, category, label, sort_order) VALUES
    ('handicrafts',        'at_home',       'Handarbeit',                       1),
    ('cooking_baking',     'at_home',       'Kochen / Backen',                  2),
    ('housework',          'at_home',       'Hausarbeit',                       3),

    ('meeting_friends',    'social',        'Freunde treffen',                  1),
    ('family_gatherings',  'social',        'Familienfeiern etc.',              2),
    ('meeting_new_people', 'social',        'Neue Leute kennenlernen',          3),

    ('photography_drawing','creative',      'Fotografieren, Zeichnen etc.',     1),
    ('fashion_cosmetics',  'creative',      'Mode / Kosmetik',                  2),
    ('interior_decoration','creative',      'Dekoration / Inneneinrichtung',    3),

    ('sport_fitness',      'mobility',      'Sport & Fitness',                  1),
    ('traveling',          'mobility',      'Reisen',                          2),
    ('nature_walks',       'mobility',      'Ausflüge in die Natur / Spazieren gehen', 3),

    ('flowers_plants',     'nature',        'Blumen / Zimmerpflanzen',          1),
    ('gardening',          'nature',        'Garten',                          2),
    ('animals',            'nature',        'Tiere',                           3),

    ('reading',            'further_ed',    'Lesen',                           1),
    ('science',            'further_ed',    'Wissenschaft',                    2),
    ('computers_internet', 'further_ed',    'Computer / Internet',             3),

    ('clubs_nightlife',    'going_out',     'Clubs, Nightlife etc.',           1),
    ('shopping_dining',    'going_out',     'Shopping, Essen gehen',           2),
    ('arts_culture',       'going_out',     'Kunst + Kultur (Kino, Theater, Oper...)', 3);

CREATE TABLE profile_hobbies (
    profile_id  UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    hobby_key   TEXT NOT NULL REFERENCES hobby_definitions(key),
    liked       BOOLEAN NOT NULL,
    intensity   SMALLINT CHECK (intensity BETWEEN 1 AND 5),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (profile_id, hobby_key),
    CONSTRAINT profile_hobbies_intensity_requires_liked
        CHECK (liked = true OR intensity IS NULL)
);

CREATE INDEX profile_hobbies_profile_id_idx ON profile_hobbies (profile_id);

CREATE TRIGGER profile_hobbies_set_updated_at
    BEFORE UPDATE ON profile_hobbies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
