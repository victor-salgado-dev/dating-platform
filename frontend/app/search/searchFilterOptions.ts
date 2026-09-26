// Copia local de las listas de valores permitidos por los CHECK de la base
// de datos, para poder construir el formulario de /search. Si ya tienes
// esto centralizado (p.ej. en lib/profileOptions.ts), borra este archivo e
// importa desde ahí en app/search/page.tsx en su lugar — está escrito así
// para no arriesgarme a pisar un archivo real con una ruta adivinada.

export const GENDER_OPTIONS = ['female', 'male', 'non_binary', 'other'];
export const HAS_CHILDREN_OPTIONS = ['yes', 'no', 'prefer_not_to_say'];
export const WANTS_CHILDREN_OPTIONS = ['yes', 'no', 'not_sure'];
export const RELATIONSHIP_GOAL_OPTIONS = ['casual', 'long_term', 'friendship', 'marriage', 'not_sure'];

export const BODY_TYPE_OPTIONS = [
  'petite', 'slim', 'athletic', 'average', 'few_extra_pounds', 'full_figured', 'large_and_lovely',
];
export const ETHNICITY_OPTIONS = [
  'arab', 'asian', 'black', 'caucasian', 'hispanic', 'indian', 'mixed', 'pacific_islander', 'other',
];
export const APPEARANCE_RATING_OPTIONS = ['below_average', 'average', 'attractive', 'very_attractive'];
export const HAIR_COLOR_OPTIONS = [
  'bald', 'black', 'blonde', 'brown', 'grey', 'light_brown', 'red', 'changes_frequently', 'other',
];
export const EYE_COLOR_OPTIONS = ['black', 'blue', 'brown', 'green', 'grey', 'hazel', 'other'];
export const BODY_ART_OPTIONS = ['branding', 'earrings', 'piercing', 'tattoo', 'other', 'none'];

export const SMOKING_HABIT_OPTIONS = ['yes', 'no', 'occasionally'];
export const DRINKING_HABIT_OPTIONS = SMOKING_HABIT_OPTIONS;
export const RELOCATION_WILLINGNESS_OPTIONS = ['within_country', 'another_country', 'not_willing', 'not_sure'];
export const MARITAL_STATUS_OPTIONS = ['single', 'separated', 'widowed', 'divorced', 'other'];
export const OCCUPATION_OPTIONS = [
  'administrative', 'advertising', 'artistic', 'construction', 'domestic_helper', 'education',
  'entertainment', 'executive', 'farming', 'finance', 'fire_law_enforcement', 'hair_dresser',
  'it_communications', 'laborer', 'legal', 'medical', 'military', 'nanny', 'none', 'non_profit',
  'political', 'retail', 'retired', 'sales', 'self_employed', 'sports', 'student', 'technical',
  'transportation', 'travel', 'unemployed', 'other',
];
export const EMPLOYMENT_STATUS_OPTIONS = [
  'student', 'part_time', 'full_time', 'homemaker', 'retired', 'not_employed', 'other',
];
export const INCOME_LEVEL_OPTIONS = ['low', 'medium', 'high', 'very_high', 'prefer_not_to_say'];
export const LIVING_SITUATION_OPTIONS = [
  'live_alone', 'live_with_friends', 'live_with_family', 'live_with_kids', 'live_with_spouse', 'other',
];

export const EDUCATION_LEVEL_OPTIONS = ['high_school', 'associates', 'bachelors', 'masters', 'phd', 'other'];
export const ENGLISH_ABILITY_OPTIONS = ['none', 'basic', 'intermediate', 'fluent', 'native'];
export const RELIGION_OPTIONS = [
  'bahai', 'buddhist', 'catholic', 'christian_other', 'protestant', 'hindu', 'islam', 'jainism',
  'jewish', 'parsi', 'shintoism', 'sikhism', 'taoism', 'other', 'none',
];
export const RELIGIOUS_VALUES_OPTIONS = ['not_religious', 'religious', 'very_religious'];
export const STAR_SIGN_OPTIONS = [
  'aquarius', 'aries', 'cancer', 'capricorn', 'gemini', 'leo', 'libra', 'pisces', 'sagittarius',
  'scorpio', 'taurus', 'virgo',
];

export const FUTURE_VISION_OPTIONS = [
  'balance_family_career', 'focus_family_household', 'beauty_and_partner_time', 'part_time_work',
  'support_partner_career', 'new_education',
];
export const SPORTS_OPTIONS = [
  'fitness', 'motorsport', 'strength_training', 'ball_sports', 'water_sports', 'jogging',
  'winter_sports', 'cycling', 'athletics', 'climbing', 'horse_riding', 'hiking', 'other',
];
export const LIKES_PETS_OPTIONS = ['yes', 'neutral', 'no'];
export const PETS_OWNED_OPTIONS = ['none', 'cat', 'dog', 'horse', 'other'];
export const FAVORITE_SEASON_OPTIONS = ['spring', 'summer', 'autumn', 'winter'];
export const IDEAL_VACATION_STYLE_OPTIONS = [
  'small_charming_hotel', 'luxury_hotel', 'cruise_ship', 'club_hotel', 'rental_apartment',
  'countryside_house', 'camping_rv', 'staying_home', 'at_friends',
];
export const VACATION_ACTIVITIES_OPTIONS = [
  'cafes_shopping_nightlife', 'mix_relaxation_activities', 'lazing_and_relaxing', 'beach_holiday',
  'sightseeing_cities', 'lots_of_sports',
];

export const PERSONALITY_TRAIT_KEYS = [
  'extraversion', 'emotional_stability', 'conscientiousness', 'agreeableness', 'openness',
];

export const INTEREST_CATEGORY_KEYS = [
  'sport_activity', 'creativity_manual', 'culture_intellectual', 'leisure_entertainment',
  'lifestyle_other', 'art_creativity', 'diy_crafts', 'music', 'music_genres', 'gaming_geek',
  'sports_specific', 'motor', 'nature_animals', 'travel', 'gastronomy', 'film_entertainment', 'books',
];

export const LANGUAGE_OPTIONS = [
  'en', 'tl', 'ceb', 'hy', 'ar', 'es', 'ja', 'af', 'sq', 'am', 'syr', 'az', 'id', 'ms', 'be', 'bn',
  'ber', 'bg', 'my', 'zh_yue', 'zh_cmn', 'cr', 'hr', 'cs', 'da', 'nl', 'ti', 'et', 'fa', 'fi', 'fr',
  'ka', 'de', 'el', 'gu', 'ha', 'he', 'hi', 'hu', 'is', 'ilo', 'iu', 'it', 'kk', 'km', 'ky', 'lo',
  'lv', 'lt', 'mk', 'mg', 'ml', 'dv', 'mt', 'mr', 'mn', 'ne', 'no', 'ps', 'pcm', 'pl', 'pt', 'qu',
  'ro', 'ru', 'sr', 'sd', 'si', 'sk', 'sl', 'so', 'sw', 'sv', 'ta', 'te', 'th', 'bo', 'to', 'tr',
  'tk', 'uga', 'uk', 'ur', 'uz', 'vi', 'cy', 'other',
];
