ALTER TABLE profiles DROP CONSTRAINT IF EXISTS profiles_languages_check;

ALTER TABLE profiles
    DROP COLUMN IF EXISTS height,
    DROP COLUMN IF EXISTS weight,
    DROP COLUMN IF EXISTS body_type,
    DROP COLUMN IF EXISTS ethnicity,
    DROP COLUMN IF EXISTS appearance_rating,
    DROP COLUMN IF EXISTS hair_color,
    DROP COLUMN IF EXISTS eye_color,
    DROP COLUMN IF EXISTS body_art,
    DROP COLUMN IF EXISTS smoking_habit,
    DROP COLUMN IF EXISTS drinking_habit,
    DROP COLUMN IF EXISTS relocation_willingness,
    DROP COLUMN IF EXISTS marital_status,
    DROP COLUMN IF EXISTS children_count,
    DROP COLUMN IF EXISTS youngest_child_age,
    DROP COLUMN IF EXISTS oldest_child_age,
    DROP COLUMN IF EXISTS occupation,
    DROP COLUMN IF EXISTS employment_status,
    DROP COLUMN IF EXISTS income_level,
    DROP COLUMN IF EXISTS living_situation,
    DROP COLUMN IF EXISTS nationality,
    DROP COLUMN IF EXISTS education_level,
    DROP COLUMN IF EXISTS english_ability,
    DROP COLUMN IF EXISTS religion,
    DROP COLUMN IF EXISTS religious_values,
    DROP COLUMN IF EXISTS star_sign;

-- Revertir has_children y wants_children a boolean (Nota: se perderán los datos "not_sure")
ALTER TABLE profiles DROP COLUMN IF EXISTS has_children;
ALTER TABLE profiles DROP COLUMN IF EXISTS wants_children;

ALTER TABLE profiles 
    ADD COLUMN has_children BOOLEAN,
    ADD COLUMN wants_children BOOLEAN;