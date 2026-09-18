-- 1. Modificar campos booleanos existentes que ahora requieren más de dos opciones.
ALTER TABLE profiles DROP COLUMN has_children;
ALTER TABLE profiles DROP COLUMN wants_children;

ALTER TABLE profiles 
    ADD COLUMN has_children TEXT CHECK (has_children IN ('yes', 'no', 'prefer_not_to_say')),
    ADD COLUMN wants_children TEXT CHECK (wants_children IN ('yes', 'no', 'not_sure'));

-- 2. Restringir el campo "languages" (que ya existía) para evitar texto libre.
-- Vaciamos los datos antiguos porque no cumplen el nuevo formato estricto de códigos ISO.
UPDATE profiles SET languages = NULL;

ALTER TABLE profiles ADD CONSTRAINT profiles_languages_check CHECK (
    languages <@ ARRAY[
        'en', 'tl', 'ceb', 'hy', 'ar', 'es', 'ja', 'af', 'sq', 'am', 'syr', 'az', 
        'id', 'ms', 'be', 'bn', 'ber', 'bg', 'my', 'zh_yue', 'zh_cmn', 'cr', 'hr', 
        'cs', 'da', 'nl', 'ti', 'et', 'fa', 'fi', 'fr', 'ka', 'de', 'el', 'gu', 
        'ha', 'he', 'hi', 'hu', 'is', 'ilo', 'iu', 'it', 'kk', 'km', 'ky', 'lo', 
        'lv', 'lt', 'mk', 'mg', 'ml', 'dv', 'mt', 'mr', 'mn', 'ne', 'no', 'ps', 
        'pcm', 'pl', 'pt', 'qu', 'ro', 'ru', 'sr', 'sd', 'si', 'sk', 'sl', 'so', 
        'sw', 'sv', 'ta', 'te', 'th', 'bo', 'to', 'tr', 'tk', 'uga', 'uk', 'ur', 
        'uz', 'vi', 'cy', 'other'
    ]::TEXT[]
);

-- 3. Físico y Apariencia
ALTER TABLE profiles
    ADD COLUMN height INTEGER CHECK (height > 50 AND height < 300), -- En centímetros
    ADD COLUMN weight INTEGER CHECK (weight > 20 AND weight < 400), -- En kilogramos
    ADD COLUMN body_type TEXT CHECK (body_type IN ('petite', 'slim', 'athletic', 'average', 'few_extra_pounds', 'full_figured', 'large_and_lovely')),
    ADD COLUMN ethnicity TEXT CHECK (ethnicity IN ('arab', 'asian', 'black', 'caucasian', 'hispanic', 'indian', 'mixed', 'pacific_islander', 'other')),
    ADD COLUMN appearance_rating TEXT CHECK (appearance_rating IN ('below_average', 'average', 'attractive', 'very_attractive')),
    ADD COLUMN hair_color TEXT CHECK (hair_color IN ('bald', 'black', 'blonde', 'brown', 'grey', 'light_brown', 'red', 'changes_frequently', 'other')),
    ADD COLUMN eye_color TEXT CHECK (eye_color IN ('black', 'blue', 'brown', 'green', 'grey', 'hazel', 'other')),
    -- Arrays múltiples
    ADD COLUMN body_art TEXT[] CHECK (body_art <@ ARRAY['branding', 'earrings', 'piercing', 'tattoo', 'other', 'none']::TEXT[]);

-- 4. Estilo de vida y Familia
ALTER TABLE profiles
    ADD COLUMN smoking_habit TEXT CHECK (smoking_habit IN ('yes', 'no', 'occasionally')),
    ADD COLUMN drinking_habit TEXT CHECK (drinking_habit IN ('yes', 'no', 'occasionally')),
    -- Arrays múltiples
    ADD COLUMN relocation_willingness TEXT[] CHECK (relocation_willingness <@ ARRAY['within_country', 'another_country', 'not_willing', 'not_sure']::TEXT[]),
    
    ADD COLUMN marital_status TEXT CHECK (marital_status IN ('single', 'separated', 'widowed', 'divorced', 'other')),
    ADD COLUMN children_count INTEGER CHECK (children_count >= 0 AND children_count <= 25),
    ADD COLUMN youngest_child_age INTEGER CHECK (youngest_child_age >= 0 AND youngest_child_age <= 100),
    ADD COLUMN oldest_child_age INTEGER CHECK (oldest_child_age >= 0 AND oldest_child_age <= 100),
    
    ADD COLUMN occupation TEXT CHECK (occupation IN (
        'administrative', 'advertising', 'artistic', 'construction', 'domestic_helper', 
        'education', 'entertainment', 'executive', 'farming', 'finance', 
        'fire_law_enforcement', 'hair_dresser', 'it_communications', 'laborer', 'legal', 
        'medical', 'military', 'nanny', 'none', 'non_profit', 'political', 'retail', 
        'retired', 'sales', 'self_employed', 'sports', 'student', 'technical', 
        'transportation', 'travel', 'unemployed', 'other'
    )),
    ADD COLUMN employment_status TEXT CHECK (employment_status IN ('student', 'part_time', 'full_time', 'homemaker', 'retired', 'not_employed', 'other')),
    ADD COLUMN income_level TEXT CHECK (income_level IN ('low', 'medium', 'high', 'very_high', 'prefer_not_to_say')),
    ADD COLUMN living_situation TEXT CHECK (living_situation IN ('live_alone', 'live_with_friends', 'live_with_family', 'live_with_kids', 'live_with_spouse', 'other'));

-- 5. Fondo, Cultura y Valores
ALTER TABLE profiles
    -- Usamos regex para forzar el código ISO de 2 letras, igual que hiciste con country_code
    ADD COLUMN nationality TEXT CHECK (nationality ~ '^[A-Z]{2}$'), 
    
    ADD COLUMN education_level TEXT CHECK (education_level IN ('high_school', 'associates', 'bachelors', 'masters', 'phd', 'other')),
    ADD COLUMN english_ability TEXT CHECK (english_ability IN ('none', 'basic', 'intermediate', 'fluent', 'native')),
    ADD COLUMN religion TEXT CHECK (religion IN ('bahai', 'buddhist', 'catholic', 'christian_other', 'protestant', 'hindu', 'islam', 'jainism', 'jewish', 'parsi', 'shintoism', 'sikhism', 'taoism', 'other', 'none')),
    ADD COLUMN religious_values TEXT CHECK (religious_values IN ('not_religious', 'religious', 'very_religious')),
    ADD COLUMN star_sign TEXT CHECK (star_sign IN ('aquarius', 'aries', 'cancer', 'capricorn', 'gemini', 'leo', 'libra', 'pisces', 'sagittarius', 'scorpio', 'taurus', 'virgo'));