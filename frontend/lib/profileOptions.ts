// Listas de valores permitidos por los CHECK de la base de datos
// (migraciones 000012, 000013 y 000014 del backend).
//
// Cada array contiene SOLO los valores técnicos que viajan al backend;
// las etiquetas visibles viven en el namespace `options` de los
// diccionarios de i18n (es.ts / en.ts), con un bucket por categoría cuyo
// nombre coincide con el nombre del array en camelCase. Se resuelven con
// los helpers tOption / tOptionList de lib/i18n/options.ts.
//
// Si se añade o quita un valor en la base de datos, hay que reflejarlo
// aquí Y en ambos diccionarios.

// --- Básico ---------------------------------------------------------------

export const HAS_CHILDREN_OPTIONS = ['yes', 'no', 'prefer_not_to_say'];
export const WANTS_CHILDREN_OPTIONS = ['yes', 'no', 'not_sure'];
export const RELATIONSHIP_GOAL_OPTIONS = ['casual', 'long_term', 'friendship', 'marriage', 'not_sure'];

// --- Físico y apariencia --------------------------------------------------

export const BODY_TYPE_OPTIONS = [
  'petite',
  'slim',
  'athletic',
  'average',
  'few_extra_pounds',
  'full_figured',
  'large_and_lovely',
];

export const ETHNICITY_OPTIONS = [
  'arab',
  'asian',
  'black',
  'caucasian',
  'hispanic',
  'indian',
  'mixed',
  'pacific_islander',
  'other',
];

export const APPEARANCE_RATING_OPTIONS = [
  'below_average',
  'average',
  'attractive',
  'very_attractive',
];

export const HAIR_COLOR_OPTIONS = [
  'bald',
  'black',
  'blonde',
  'brown',
  'grey',
  'light_brown',
  'red',
  'changes_frequently',
  'other',
];

export const EYE_COLOR_OPTIONS = ['black', 'blue', 'brown', 'green', 'grey', 'hazel', 'other'];

export const BODY_ART_OPTIONS = ['branding', 'earrings', 'piercing', 'tattoo', 'other', 'none'];

// --- Estilo de vida y familia ---------------------------------------------

export const SMOKING_HABIT_OPTIONS = ['yes', 'no', 'occasionally'];
export const DRINKING_HABIT_OPTIONS = SMOKING_HABIT_OPTIONS;

export const RELOCATION_WILLINGNESS_OPTIONS = [
  'within_country',
  'another_country',
  'not_willing',
  'not_sure',
];

export const MARITAL_STATUS_OPTIONS = ['single', 'separated', 'widowed', 'divorced', 'other'];

export const OCCUPATION_OPTIONS = [
  'administrative',
  'advertising',
  'artistic',
  'construction',
  'domestic_helper',
  'education',
  'entertainment',
  'executive',
  'farming',
  'finance',
  'fire_law_enforcement',
  'hair_dresser',
  'it_communications',
  'laborer',
  'legal',
  'medical',
  'military',
  'nanny',
  'none',
  'non_profit',
  'political',
  'retail',
  'retired',
  'sales',
  'self_employed',
  'sports',
  'student',
  'technical',
  'transportation',
  'travel',
  'unemployed',
  'other',
];

export const EMPLOYMENT_STATUS_OPTIONS = [
  'student',
  'part_time',
  'full_time',
  'homemaker',
  'retired',
  'not_employed',
  'other',
];

export const INCOME_LEVEL_OPTIONS = ['low', 'medium', 'high', 'very_high', 'prefer_not_to_say'];

export const LIVING_SITUATION_OPTIONS = [
  'live_alone',
  'live_with_friends',
  'live_with_family',
  'live_with_kids',
  'live_with_spouse',
  'other',
];

// --- Fondo, cultura y valores ---------------------------------------------

export const EDUCATION_LEVEL_OPTIONS = [
  'high_school',
  'associates',
  'bachelors',
  'masters',
  'phd',
  'other',
];

export const ENGLISH_ABILITY_OPTIONS = ['none', 'basic', 'intermediate', 'fluent', 'native'];

export const RELIGION_OPTIONS = [
  'bahai',
  'buddhist',
  'catholic',
  'christian_other',
  'protestant',
  'hindu',
  'islam',
  'jainism',
  'jewish',
  'parsi',
  'shintoism',
  'sikhism',
  'taoism',
  'other',
  'none',
];

export const RELIGIOUS_VALUES_OPTIONS = ['not_religious', 'religious', 'very_religious'];

export const STAR_SIGN_OPTIONS = [
  'aquarius',
  'aries',
  'cancer',
  'capricorn',
  'gemini',
  'leo',
  'libra',
  'pisces',
  'sagittarius',
  'scorpio',
  'taurus',
  'virgo',
];

// --- Über mich / estilo de vida -------------------------------------------

export const FUTURE_VISION_OPTIONS = [
  'balance_family_career',
  'focus_family_household',
  'beauty_and_partner_time',
  'part_time_work',
  'support_partner_career',
  'new_education',
];

export const SPORTS_OPTIONS = [
  'fitness',
  'motorsport',
  'strength_training',
  'ball_sports',
  'water_sports',
  'jogging',
  'winter_sports',
  'cycling',
  'athletics',
  'climbing',
  'horse_riding',
  'hiking',
  'other',
];

export const LIKES_PETS_OPTIONS = ['yes', 'neutral', 'no'];

export const PETS_OWNED_OPTIONS = ['none', 'cat', 'dog', 'horse', 'other'];

export const FAVORITE_SEASON_OPTIONS = ['spring', 'summer', 'autumn', 'winter'];

export const IDEAL_VACATION_STYLE_OPTIONS = [
  'small_charming_hotel',
  'luxury_hotel',
  'cruise_ship',
  'club_hotel',
  'rental_apartment',
  'countryside_house',
  'camping_rv',
  'staying_home',
  'at_friends',
];

export const VACATION_ACTIVITIES_OPTIONS = [
  'cafes_shopping_nightlife',
  'mix_relaxation_activities',
  'lazing_and_relaxing',
  'beach_holiday',
  'sightseeing_cities',
  'lots_of_sports',
];

// --- Preferencias de pareja -----------------------------------------------

export const DESIRED_TRAITS_OPTIONS = [
  'humorous',
  'self_confident',
  'loving',
  'kind_hearted',
  'intelligent',
  'faithful',
  'honest',
  'ambitious',
  'family_oriented',
  'adventurous',
  'romantic',
  'patient',
  'easy_going',
  'financially_stable',
  'other',
];

export const PARTNER_MAY_HAVE_CHILDREN_OPTIONS = ['yes', 'no', 'doesnt_matter'];

export const PARTNER_RELIGION_PREFERENCE_OPTIONS = [...RELIGION_OPTIONS, 'doesnt_matter'];

export const FIRST_MEETING_PREFERENCE_OPTIONS = [
  'doesnt_matter',
  'public_place',
  'my_city',
  'their_city',
  'video_call_first',
];

export const DESIRED_LIVING_PLACE_OPTIONS = [
  'big_city',
  'medium_city',
  'small_town',
  'countryside',
  'abroad',
];

// Las 10 barras de "qué es importante en una relación". Lista fija; los
// campos del partnerForm se indexan por `key`, por eso mantenemos el shape.
export const PARTNER_IMPORTANCE_FIELDS: { key: string }[] = [
  { key: 'importance_shared_thoughts' },
  { key: 'importance_shared_hobbies' },
  { key: 'importance_intimacy' },
  { key: 'importance_romantic_love' },
  { key: 'importance_financial_security' },
  { key: 'importance_fun' },
  { key: 'importance_shared_friends' },
  { key: 'importance_shared_humor' },
  { key: 'importance_personal_space' },
  { key: 'importance_independence' },
];

// Los 5 rasgos del Big Five. No se iteran aquí, pero se listan para dejar
// constancia de los valores admitidos por el backend.
export const PERSONALITY_TRAIT_KEYS = [
  'extraversion',
  'emotional_stability',
  'conscientiousness',
  'agreeableness',
  'openness',
];

// Categorías de interés del catálogo del backend.
export const INTEREST_CATEGORY_KEYS = [
  'sport_activity',
  'creativity_manual',
  'culture_intellectual',
  'leisure_entertainment',
  'lifestyle_other',
  'art_creativity',
  'diy_crafts',
  'music',
  'music_genres',
  'gaming_geek',
  'sports_specific',
  'motor',
  'nature_animals',
  'travel',
  'gastronomy',
  'film_entertainment',
  'books',
];

// Códigos ISO de `profile_languages.language_code`.
export const LANGUAGE_OPTIONS = [
  'en',
  'tl',
  'ceb',
  'hy',
  'ar',
  'es',
  'ja',
  'af',
  'sq',
  'am',
  'syr',
  'az',
  'id',
  'ms',
  'be',
  'bn',
  'ber',
  'bg',
  'my',
  'zh_yue',
  'zh_cmn',
  'cr',
  'hr',
  'cs',
  'da',
  'nl',
  'ti',
  'et',
  'fa',
  'fi',
  'fr',
  'ka',
  'de',
  'el',
  'gu',
  'ha',
  'he',
  'hi',
  'hu',
  'is',
  'ilo',
  'iu',
  'it',
  'kk',
  'km',
  'ky',
  'lo',
  'lv',
  'lt',
  'mk',
  'mg',
  'ml',
  'dv',
  'mt',
  'mr',
  'mn',
  'ne',
  'no',
  'ps',
  'pcm',
  'pl',
  'pt',
  'qu',
  'ro',
  'ru',
  'sr',
  'sd',
  'si',
  'sk',
  'sl',
  'so',
  'sw',
  'sv',
  'ta',
  'te',
  'th',
  'bo',
  'to',
  'tr',
  'tk',
  'uga',
  'uk',
  'ur',
  'uz',
  'vi',
  'cy',
  'other',
];
