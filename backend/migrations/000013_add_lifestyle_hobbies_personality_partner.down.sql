-- 000013_add_lifestyle_hobbies_personality_partner.down.sql

DROP TRIGGER IF EXISTS profile_partner_preferences_set_updated_at ON profile_partner_preferences;
DROP TABLE IF EXISTS profile_partner_preferences;

DROP VIEW IF EXISTS profile_personality_trait_scores;

DROP TRIGGER IF EXISTS profile_personality_answers_set_updated_at ON profile_personality_answers;
DROP TABLE IF EXISTS profile_personality_answers;
DROP TABLE IF EXISTS personality_statements;

DROP TRIGGER IF EXISTS profile_hobbies_set_updated_at ON profile_hobbies;
DROP TABLE IF EXISTS profile_hobbies;
DROP TABLE IF EXISTS hobby_definitions;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS future_vision,
    DROP COLUMN IF EXISTS sports,
    DROP COLUMN IF EXISTS likes_pets,
    DROP COLUMN IF EXISTS pets_owned,
    DROP COLUMN IF EXISTS favorite_season,
    DROP COLUMN IF EXISTS ideal_vacation_style,
    DROP COLUMN IF EXISTS vacation_activities,
    DROP COLUMN IF EXISTS profile_quote,
    DROP COLUMN IF EXISTS dream_wish;
