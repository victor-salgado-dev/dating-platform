-- seed_fake_users.sql
--
-- Rellena, para TODOS los perfiles, lo que SÍ vive en la base de datos:
--   1. profiles.seeking_genders        (género que buscan)
--   2. profile_partner_preferences     (edad, altura, rasgos, importancias...)
--
-- Sobrescribe lo que hubiera. Úsalo solo con usuarios falsos.
-- Para limitarlo a unas cuentas, añade en los dos bloques:
--   JOIN users u ON u.id = p.user_id WHERE u.email LIKE '%@datingdemo.com'
--
-- Los FILTROS DE BÚSQUEDA (pantalla Search) NO están aquí: viven en el
-- localStorage del navegador de cada usuario, no en la base de datos.

BEGIN;

-- ---------------------------------------------------------------------
-- 1. Género buscado (obligatorio en el perfil)
--    ~80 % busca el género contrario, ~12 % el mismo, ~8 % cualquiera.
-- ---------------------------------------------------------------------
UPDATE profiles p
SET seeking_genders = CASE
    WHEN r.r < 0.80 THEN
        CASE p.gender
            WHEN 'female' THEN ARRAY['male']
            WHEN 'male'   THEN ARRAY['female']
            ELSE ARRAY['female', 'male']
        END
    WHEN r.r < 0.92 THEN
        CASE p.gender
            WHEN 'female' THEN ARRAY['female']
            WHEN 'male'   THEN ARRAY['male']
            ELSE ARRAY['non_binary', 'other']
        END
    ELSE ARRAY['female', 'male', 'non_binary', 'other']
END
FROM (SELECT id, random() AS r FROM profiles) r
WHERE r.id = p.id;

-- ---------------------------------------------------------------------
-- 2. Preferencias de pareja
--    ~20 % sin rango de edad (NULL = "no lo ha dicho"), para ejercitar
--    el caso en que no hay preferencia y los listados no filtran por edad.
-- ---------------------------------------------------------------------
INSERT INTO profile_partner_preferences (
    profile_id,
    age_min, age_max, height_min, height_max,
    desired_traits,
    partner_may_have_children, partner_religion_preference,
    about_partner_text, first_meeting_preference, desired_living_place,
    importance_shared_thoughts, importance_shared_hobbies, importance_intimacy,
    importance_romantic_love, importance_financial_security, importance_fun,
    importance_shared_friends, importance_shared_humor,
    importance_personal_space, importance_independence,
    updated_at
)
SELECT
    p.id,
    CASE WHEN x.no_age THEN NULL
         ELSE GREATEST(18, x.my_age - (3 + (random() * 3)::int)) END,
    CASE WHEN x.no_age THEN NULL
         ELSE LEAST(90, x.my_age + (5 + (random() * 6)::int)) END,
    150 + (random() * 15)::int,
    180 + (random() * 25)::int,
    ARRAY[
        (ARRAY['humorous', 'kind_hearted', 'honest', 'loving'])[floor(random() * 4) + 1],
        (ARRAY['intelligent', 'adventurous', 'easy_going', 'romantic'])[floor(random() * 4) + 1],
        (ARRAY['family_oriented', 'financially_stable', 'self_confident', 'faithful'])[floor(random() * 4) + 1]
    ]::text[],
    (ARRAY['yes', 'no', 'doesnt_matter'])[floor(random() * 3) + 1],
    (ARRAY['doesnt_matter', 'none', 'catholic', 'christian_other'])[floor(random() * 4) + 1],
    'Busco una persona sincera, con sentido del humor y ganas de compartir buenos momentos.',
    (ARRAY['public_place', 'doesnt_matter', 'video_call_first'])[floor(random() * 3) + 1],
    ARRAY['big_city', 'medium_city']::text[],
    3 + (random() * 2)::int,
    2 + (random() * 3)::int,
    3 + (random() * 2)::int,
    3 + (random() * 2)::int,
    2 + (random() * 3)::int,
    4 + (random() * 1)::int,
    2 + (random() * 2)::int,
    4 + (random() * 1)::int,
    3 + (random() * 2)::int,
    3 + (random() * 2)::int,
    now()
FROM profiles p
CROSS JOIN LATERAL (
    SELECT EXTRACT(YEAR FROM age(current_date, p.birth_date))::int AS my_age,
           random() < 0.20 AS no_age
) x
ON CONFLICT (profile_id) DO UPDATE SET
    age_min = EXCLUDED.age_min,
    age_max = EXCLUDED.age_max,
    height_min = EXCLUDED.height_min,
    height_max = EXCLUDED.height_max,
    desired_traits = EXCLUDED.desired_traits,
    partner_may_have_children = EXCLUDED.partner_may_have_children,
    partner_religion_preference = EXCLUDED.partner_religion_preference,
    about_partner_text = EXCLUDED.about_partner_text,
    first_meeting_preference = EXCLUDED.first_meeting_preference,
    desired_living_place = EXCLUDED.desired_living_place,
    importance_shared_thoughts = EXCLUDED.importance_shared_thoughts,
    importance_shared_hobbies = EXCLUDED.importance_shared_hobbies,
    importance_intimacy = EXCLUDED.importance_intimacy,
    importance_romantic_love = EXCLUDED.importance_romantic_love,
    importance_financial_security = EXCLUDED.importance_financial_security,
    importance_fun = EXCLUDED.importance_fun,
    importance_shared_friends = EXCLUDED.importance_shared_friends,
    importance_shared_humor = EXCLUDED.importance_shared_humor,
    importance_personal_space = EXCLUDED.importance_personal_space,
    importance_independence = EXCLUDED.importance_independence,
    updated_at = now();

COMMIT;

-- La carga/actualización masiva cambia la distribución que usa el planner.
-- Refrescar estadísticas evita que las búsquedas planifiquen como si estas
-- tablas siguieran vacías, especialmente justo después de preparar los datos.
ANALYZE;

-- ---------------------------------------------------------------------
-- Comprobación rápida
-- ---------------------------------------------------------------------
SELECT gender, seeking_genders, count(*) AS perfiles
FROM profiles
GROUP BY gender, seeking_genders
ORDER BY gender, perfiles DESC;

SELECT count(*) FILTER (WHERE age_min IS NULL) AS sin_rango_edad,
       count(*) FILTER (WHERE age_min IS NOT NULL) AS con_rango_edad
FROM profile_partner_preferences;
