-- 000013_add_lifestyle_hobbies_personality_partner.up.sql
--
-- Fase 1 de la ampliación del perfil (secciones "Über mich", "Hobbies",
-- "Persönlichkeit" y "Partner" del modelo de referencia).
--
-- Decisión de diseño (acordada antes de escribir esto):
--   - "Über mich": grupos select/multi-select de cardinalidad baja y
--     estable -> columnas planas en `profiles` con CHECK, igual que en
--     la migración 000012. No necesitan catálogo.
--   - "Hobbies" y "Personalidad": catálogos de ítems que van a crecer
--     con el tiempo (nuevos hobbies, nuevas afirmaciones) -> tablas de
--     catálogo + tabla de respuestas, para que añadir un ítem sea un
--     INSERT y no una migración.
--   - "Partner": rangos y listas -> columnas en una tabla 1:1 propia
--     (no en `profiles`, para no seguir engordando esa tabla con un
--     dominio distinto: "lo que busco" vs "quién soy"). Las ~10
--     "importancias" son una lista fija y pequeña -> columnas, no
--     catálogo (decisión explícita del usuario).
--   - Hobbies con feedback negativo explícito ("no me gusta"): no basta
--     con "marcado + intensidad". Se modela con `liked BOOLEAN NOT NULL`
--     + `intensity` (solo tiene sentido si liked = true). Así "sin
--     fila" = no contestado, "liked=false" = no me gusta explícito,
--     "liked=true + intensity" = me gusta con un grado.

-- =====================================================================
-- 1. "Über mich": nuevos campos de estilo de vida en `profiles`
-- =====================================================================

ALTER TABLE profiles
    ADD COLUMN future_vision TEXT[] CHECK (future_vision <@ ARRAY[
        'balance_family_career', 'focus_family_household', 'beauty_and_partner_time',
        'part_time_work', 'support_partner_career', 'new_education'
    ]::TEXT[]),

    ADD COLUMN sports TEXT[] CHECK (sports <@ ARRAY[
        'fitness', 'motorsport', 'strength_training', 'ball_sports', 'water_sports',
        'jogging', 'winter_sports', 'cycling', 'athletics', 'climbing',
        'horse_riding', 'hiking', 'other'
    ]::TEXT[]),

    ADD COLUMN likes_pets TEXT CHECK (likes_pets IN ('yes', 'neutral', 'no')),

    ADD COLUMN pets_owned TEXT[] CHECK (pets_owned <@ ARRAY[
        'none', 'cat', 'dog', 'horse', 'other'
    ]::TEXT[]),

    ADD COLUMN favorite_season TEXT CHECK (favorite_season IN ('spring', 'summer', 'autumn', 'winter')),

    ADD COLUMN ideal_vacation_style TEXT[] CHECK (ideal_vacation_style <@ ARRAY[
        'small_charming_hotel', 'luxury_hotel', 'cruise_ship', 'club_hotel',
        'rental_apartment', 'countryside_house', 'camping_rv', 'staying_home', 'at_friends'
    ]::TEXT[]),

    ADD COLUMN vacation_activities TEXT[] CHECK (vacation_activities <@ ARRAY[
        'cafes_shopping_nightlife', 'mix_relaxation_activities', 'lazing_and_relaxing',
        'beach_holiday', 'sightseeing_cities', 'lots_of_sports'
    ]::TEXT[]),

    -- Cita/frase de presentación libre (la que aparece en cursiva bajo "Über mich").
    ADD COLUMN profile_quote TEXT CHECK (profile_quote IS NULL OR char_length(profile_quote) <= 500),

    -- "¿Qué es lo que siempre he querido hacer? ¿Cuál es mi sueño más loco?"
    ADD COLUMN dream_wish TEXT CHECK (dream_wish IS NULL OR char_length(dream_wish) <= 500);

-- smoking_habit y drinking_habit ya existen desde la 000012 con el mismo
-- rango de valores (yes/no/occasionally) que "Rauche ich? / Trinke ich
-- Alkohol?", así que se reutilizan tal cual.

-- =====================================================================
-- 2. Hobbies: catálogo + respuestas por perfil
-- =====================================================================

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

    -- Si no le gusta, no tiene sentido guardar un grado de intensidad.
    CONSTRAINT profile_hobbies_intensity_requires_liked
        CHECK (liked = true OR intensity IS NULL)
);

CREATE INDEX profile_hobbies_profile_id_idx ON profile_hobbies (profile_id);

CREATE TRIGGER profile_hobbies_set_updated_at
    BEFORE UPDATE ON profile_hobbies
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- =====================================================================
-- 3. Personalidad: catálogo de afirmaciones + respuestas por perfil
-- =====================================================================

CREATE TABLE personality_statements (
    key         TEXT PRIMARY KEY,
    trait_key   TEXT NOT NULL,
    label       TEXT NOT NULL,
    sort_order  SMALLINT NOT NULL DEFAULT 0,

    CONSTRAINT personality_statements_trait_check CHECK (trait_key IN (
        'extraversion', 'emotional_stability', 'conscientiousness',
        'agreeableness', 'openness'
    ))
);

INSERT INTO personality_statements (key, trait_key, label, sort_order) VALUES
    ('extra_reserved_calm',         'extraversion',        'Ich bin eher zurückhaltend und ruhig.',                        1),
    ('extra_funny_laughs',          'extraversion',        'In Gesellschaft bin ich lustig und lache viel.',               2),
    ('extra_center_of_party',       'extraversion',        'Ich mag es auf einer Party im Mittelpunkt zu stehen.',         3),
    ('extra_enjoys_alone_time',     'extraversion',        'Ich bin auch sehr gerne allein.',                              4),

    ('emo_sensitive_vulnerable',    'emotional_stability', 'Ich bin sehr sensibel und verletzlich.',                       1),
    ('emo_moody',                   'emotional_stability', 'Ich bin manchmal launisch.',                                   2),
    ('emo_self_confident',          'emotional_stability', 'Meine Freunde sagen, dass ich eine selbstbewusste Frau bin.',  3),
    ('emo_hard_to_rattle',          'emotional_stability', 'Ich bin so schnell durch nichts aus der Fassung zu bringen.',  4),

    ('cons_chaotic',                'conscientiousness',   'Ich bin ein eher chaotischer Mensch.',                         1),
    ('cons_no_planning',            'conscientiousness',   'Am liebsten lebe ich in den Tag hinein und plane nichts.',     2),
    ('cons_goal_oriented',          'conscientiousness',   'Ich bin zielstrebig und gebe nicht so schnell auf, wenn ich mir etwas vorgenommen habe.', 3),
    ('cons_very_tidy',              'conscientiousness',   'Ich bin ein sehr ordentlicher Mensch.',                        4),

    ('agree_distrustful_at_first',  'agreeableness',       'Neuen Menschen gegenüber bin ich zunächst misstrauisch.',      1),
    ('agree_helpful_caring',        'agreeableness',       'Ich bin sehr hilfsbereit und sorge mich um andere Menschen.',  2),
    ('agree_hard_to_get_along',     'agreeableness',       'Mit manchen Menschen komme ich einfach nicht klar.',           3),
    ('agree_believes_in_good',      'agreeableness',       'Ich glaube an das Gute im Menschen.',                          4),

    ('open_original_new_ideas',     'openness',             'Ich bin originell und habe oft neue Ideen.',                   1),
    ('open_cautious_with_new',      'openness',             'Neuem gegenüber bin ich eher vorsichtig.',                     2),
    ('open_interested_arts',        'openness',             'Ich interessiere mich sehr für Kunst, Musik und Kultur.',      3),
    ('open_traditions_matter',      'openness',             'Traditionen und alte Werte sind mir sehr wichtig.',            4);

CREATE TABLE profile_personality_answers (
    profile_id      UUID NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    statement_key   TEXT NOT NULL REFERENCES personality_statements(key),
    score           SMALLINT NOT NULL CHECK (score BETWEEN 1 AND 5),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (profile_id, statement_key)
);

CREATE INDEX profile_personality_answers_profile_id_idx ON profile_personality_answers (profile_id);

CREATE TRIGGER profile_personality_answers_set_updated_at
    BEFORE UPDATE ON profile_personality_answers
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();

-- El "Gesamt" por rasgo (la barra agregada) se calcula al vuelo, no se
-- almacena: así nunca puede desincronizarse de las respuestas
-- individuales. Si en el futuro pesa demasiado en las consultas de
-- búsqueda, se puede materializar sin tocar esta vista.
CREATE VIEW profile_personality_trait_scores AS
    SELECT
        pa.profile_id,
        ps.trait_key,
        AVG(pa.score)::NUMERIC(3, 2) AS avg_score,
        COUNT(*) AS answered_count
    FROM profile_personality_answers pa
    JOIN personality_statements ps ON ps.key = pa.statement_key
    GROUP BY pa.profile_id, ps.trait_key;

-- =====================================================================
-- 4. Preferencias de pareja ("Partner"): tabla 1:1 propia
-- =====================================================================

CREATE TABLE profile_partner_preferences (
    profile_id UUID PRIMARY KEY REFERENCES profiles(id) ON DELETE CASCADE,

    age_min    SMALLINT CHECK (age_min >= 18 AND age_min <= 120),
    age_max    SMALLINT CHECK (age_max >= 18 AND age_max <= 120),
    height_min SMALLINT CHECK (height_min > 50 AND height_min < 300),
    height_max SMALLINT CHECK (height_max > 50 AND height_max < 300),

    desired_traits TEXT[] CHECK (desired_traits <@ ARRAY[
        'humorous', 'self_confident', 'loving', 'kind_hearted', 'intelligent',
        'faithful', 'honest', 'ambitious', 'family_oriented', 'adventurous',
        'romantic', 'patient', 'easy_going', 'financially_stable', 'other'
    ]::TEXT[]),

    partner_may_have_children TEXT CHECK (partner_may_have_children IN ('yes', 'no', 'doesnt_matter')),

    -- Reutiliza el mismo dominio de valores que profiles.religion, más 'doesnt_matter'.
    partner_religion_preference TEXT CHECK (partner_religion_preference IN (
        'bahai', 'buddhist', 'catholic', 'christian_other', 'protestant', 'hindu',
        'islam', 'jainism', 'jewish', 'parsi', 'shintoism', 'sikhism', 'taoism',
        'other', 'none', 'doesnt_matter'
    )),

    about_partner_text TEXT CHECK (about_partner_text IS NULL OR char_length(about_partner_text) <= 1000),

    first_meeting_preference TEXT CHECK (first_meeting_preference IN (
        'doesnt_matter', 'public_place', 'my_city', 'their_city', 'video_call_first'
    )),

    desired_living_place TEXT[] CHECK (desired_living_place <@ ARRAY[
        'big_city', 'medium_city', 'small_town', 'countryside', 'abroad'
    ]::TEXT[]),

    -- "Qué aspectos son importantes en una relación" (las barras de la
    -- pestaña Partner). Lista fija y pequeña -> columnas, no catálogo.
    importance_shared_thoughts     SMALLINT CHECK (importance_shared_thoughts BETWEEN 1 AND 5),
    importance_shared_hobbies      SMALLINT CHECK (importance_shared_hobbies BETWEEN 1 AND 5),
    importance_intimacy            SMALLINT CHECK (importance_intimacy BETWEEN 1 AND 5),
    importance_romantic_love       SMALLINT CHECK (importance_romantic_love BETWEEN 1 AND 5),
    importance_financial_security  SMALLINT CHECK (importance_financial_security BETWEEN 1 AND 5),
    importance_fun                 SMALLINT CHECK (importance_fun BETWEEN 1 AND 5),
    importance_shared_friends      SMALLINT CHECK (importance_shared_friends BETWEEN 1 AND 5),
    importance_shared_humor        SMALLINT CHECK (importance_shared_humor BETWEEN 1 AND 5),
    importance_personal_space      SMALLINT CHECK (importance_personal_space BETWEEN 1 AND 5),
    importance_independence        SMALLINT CHECK (importance_independence BETWEEN 1 AND 5),

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT profile_partner_preferences_age_range_check
        CHECK (age_min IS NULL OR age_max IS NULL OR age_min <= age_max),
    CONSTRAINT profile_partner_preferences_height_range_check
        CHECK (height_min IS NULL OR height_max IS NULL OR height_min <= height_max)
);

CREATE TRIGGER profile_partner_preferences_set_updated_at
    BEFORE UPDATE ON profile_partner_preferences
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
