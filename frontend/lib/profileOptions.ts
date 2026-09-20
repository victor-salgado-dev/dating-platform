// Listas de opciones que reflejan los valores permitidos por los CHECK
// de la base de datos (migraciones 000012 y 000013 del backend). Si se
// añade o quita un valor allí, hay que reflejarlo aquí también — no hay
// forma de derivarlo automáticamente porque esas listas viven en SQL.

export type Option = { value: string; label: string };

// --- Básico -------------------------------------------------------------

export const HAS_CHILDREN_OPTIONS: Option[] = [
  { value: 'yes', label: 'Sí' },
  { value: 'no', label: 'No' },
  { value: 'prefer_not_to_say', label: 'Prefiero no decirlo' },
];

export const WANTS_CHILDREN_OPTIONS: Option[] = [
  { value: 'yes', label: 'Sí' },
  { value: 'no', label: 'No' },
  { value: 'not_sure', label: 'No lo sé todavía' },
];

export const RELATIONSHIP_GOAL_OPTIONS: Option[] = [
  { value: 'casual', label: 'Algo casual' },
  { value: 'long_term', label: 'Relación estable' },
  { value: 'friendship', label: 'Amistad' },
  { value: 'marriage', label: 'Matrimonio' },
  { value: 'not_sure', label: 'No lo sé todavía' },
];

// --- Físico y apariencia --------------------------------------------------

export const BODY_TYPE_OPTIONS: Option[] = [
  { value: 'petite', label: 'Menuda' },
  { value: 'slim', label: 'Delgada' },
  { value: 'athletic', label: 'Atlética' },
  { value: 'average', label: 'Media' },
  { value: 'few_extra_pounds', label: 'Algunos kilos de más' },
  { value: 'full_figured', label: 'Con curvas' },
  { value: 'large_and_lovely', label: 'Grande' },
];

export const ETHNICITY_OPTIONS: Option[] = [
  { value: 'arab', label: 'Árabe' },
  { value: 'asian', label: 'Asiática' },
  { value: 'black', label: 'Negra' },
  { value: 'caucasian', label: 'Caucásica' },
  { value: 'hispanic', label: 'Hispana' },
  { value: 'indian', label: 'India' },
  { value: 'mixed', label: 'Mestiza' },
  { value: 'pacific_islander', label: 'Isleña del Pacífico' },
  { value: 'other', label: 'Otra' },
];

export const APPEARANCE_RATING_OPTIONS: Option[] = [
  { value: 'below_average', label: 'Por debajo de la media' },
  { value: 'average', label: 'Media' },
  { value: 'attractive', label: 'Atractiva' },
  { value: 'very_attractive', label: 'Muy atractiva' },
];

export const HAIR_COLOR_OPTIONS: Option[] = [
  { value: 'bald', label: 'Calvo/a' },
  { value: 'black', label: 'Negro' },
  { value: 'blonde', label: 'Rubio' },
  { value: 'brown', label: 'Castaño' },
  { value: 'grey', label: 'Canoso' },
  { value: 'light_brown', label: 'Castaño claro' },
  { value: 'red', label: 'Pelirrojo' },
  { value: 'changes_frequently', label: 'Cambia con frecuencia' },
  { value: 'other', label: 'Otro' },
];

export const EYE_COLOR_OPTIONS: Option[] = [
  { value: 'black', label: 'Negros' },
  { value: 'blue', label: 'Azules' },
  { value: 'brown', label: 'Marrones' },
  { value: 'green', label: 'Verdes' },
  { value: 'grey', label: 'Grises' },
  { value: 'hazel', label: 'Avellana' },
  { value: 'other', label: 'Otros' },
];

export const BODY_ART_OPTIONS: Option[] = [
  { value: 'branding', label: 'Branding' },
  { value: 'earrings', label: 'Pendientes' },
  { value: 'piercing', label: 'Piercing' },
  { value: 'tattoo', label: 'Tatuajes' },
  { value: 'other', label: 'Otro' },
  { value: 'none', label: 'Ninguno' },
];

// --- Estilo de vida y familia ---------------------------------------------

export const SMOKING_HABIT_OPTIONS: Option[] = [
  { value: 'yes', label: 'Sí' },
  { value: 'no', label: 'No' },
  { value: 'occasionally', label: 'A veces' },
];

export const DRINKING_HABIT_OPTIONS: Option[] = SMOKING_HABIT_OPTIONS;

export const RELOCATION_WILLINGNESS_OPTIONS: Option[] = [
  { value: 'within_country', label: 'Dentro de mi país' },
  { value: 'another_country', label: 'A otro país' },
  { value: 'not_willing', label: 'No dispuesto/a' },
  { value: 'not_sure', label: 'No lo sé todavía' },
];

export const MARITAL_STATUS_OPTIONS: Option[] = [
  { value: 'single', label: 'Soltero/a' },
  { value: 'separated', label: 'Separado/a' },
  { value: 'widowed', label: 'Viudo/a' },
  { value: 'divorced', label: 'Divorciado/a' },
  { value: 'other', label: 'Otro' },
];

export const OCCUPATION_OPTIONS: Option[] = [
  { value: 'administrative', label: 'Administrativo' },
  { value: 'advertising', label: 'Publicidad' },
  { value: 'artistic', label: 'Artístico' },
  { value: 'construction', label: 'Construcción' },
  { value: 'domestic_helper', label: 'Empleo doméstico' },
  { value: 'education', label: 'Educación' },
  { value: 'entertainment', label: 'Entretenimiento' },
  { value: 'executive', label: 'Directivo' },
  { value: 'farming', label: 'Agricultura' },
  { value: 'finance', label: 'Finanzas' },
  { value: 'fire_law_enforcement', label: 'Bomberos/Policía' },
  { value: 'hair_dresser', label: 'Peluquería' },
  { value: 'it_communications', label: 'IT/Comunicaciones' },
  { value: 'laborer', label: 'Obrero' },
  { value: 'legal', label: 'Legal' },
  { value: 'medical', label: 'Médico' },
  { value: 'military', label: 'Militar' },
  { value: 'nanny', label: 'Niñera' },
  { value: 'none', label: 'Ninguna' },
  { value: 'non_profit', label: 'ONG' },
  { value: 'political', label: 'Política' },
  { value: 'retail', label: 'Comercio' },
  { value: 'retired', label: 'Jubilado/a' },
  { value: 'sales', label: 'Ventas' },
  { value: 'self_employed', label: 'Autónomo/a' },
  { value: 'sports', label: 'Deporte' },
  { value: 'student', label: 'Estudiante' },
  { value: 'technical', label: 'Técnico' },
  { value: 'transportation', label: 'Transporte' },
  { value: 'travel', label: 'Turismo' },
  { value: 'unemployed', label: 'Desempleado/a' },
  { value: 'other', label: 'Otra' },
];

export const EMPLOYMENT_STATUS_OPTIONS: Option[] = [
  { value: 'student', label: 'Estudiante' },
  { value: 'part_time', label: 'Media jornada' },
  { value: 'full_time', label: 'Jornada completa' },
  { value: 'homemaker', label: 'Labores del hogar' },
  { value: 'retired', label: 'Jubilado/a' },
  { value: 'not_employed', label: 'Sin empleo' },
  { value: 'other', label: 'Otro' },
];

export const INCOME_LEVEL_OPTIONS: Option[] = [
  { value: 'low', label: 'Bajo' },
  { value: 'medium', label: 'Medio' },
  { value: 'high', label: 'Alto' },
  { value: 'very_high', label: 'Muy alto' },
  { value: 'prefer_not_to_say', label: 'Prefiero no decirlo' },
];

export const LIVING_SITUATION_OPTIONS: Option[] = [
  { value: 'live_alone', label: 'Vivo solo/a' },
  { value: 'live_with_friends', label: 'Con amigos' },
  { value: 'live_with_family', label: 'Con familia' },
  { value: 'live_with_kids', label: 'Con mis hijos' },
  { value: 'live_with_spouse', label: 'Con mi pareja' },
  { value: 'other', label: 'Otra' },
];

// --- Fondo, cultura y valores ----------------------------------------------

export const EDUCATION_LEVEL_OPTIONS: Option[] = [
  { value: 'high_school', label: 'Secundaria' },
  { value: 'associates', label: 'Grado medio' },
  { value: 'bachelors', label: 'Grado / Licenciatura' },
  { value: 'masters', label: 'Máster' },
  { value: 'phd', label: 'Doctorado' },
  { value: 'other', label: 'Otro' },
];

export const ENGLISH_ABILITY_OPTIONS: Option[] = [
  { value: 'none', label: 'Ninguno' },
  { value: 'basic', label: 'Básico' },
  { value: 'intermediate', label: 'Intermedio' },
  { value: 'fluent', label: 'Fluido' },
  { value: 'native', label: 'Nativo' },
];

export const RELIGION_OPTIONS: Option[] = [
  { value: 'bahai', label: 'Bahá\u2019í' },
  { value: 'buddhist', label: 'Budista' },
  { value: 'catholic', label: 'Católica' },
  { value: 'christian_other', label: 'Cristiana (otra)' },
  { value: 'protestant', label: 'Protestante' },
  { value: 'hindu', label: 'Hindú' },
  { value: 'islam', label: 'Islam' },
  { value: 'jainism', label: 'Jainismo' },
  { value: 'jewish', label: 'Judía' },
  { value: 'parsi', label: 'Parsi' },
  { value: 'shintoism', label: 'Sintoísmo' },
  { value: 'sikhism', label: 'Sijismo' },
  { value: 'taoism', label: 'Taoísmo' },
  { value: 'other', label: 'Otra' },
  { value: 'none', label: 'Ninguna' },
];

export const RELIGIOUS_VALUES_OPTIONS: Option[] = [
  { value: 'not_religious', label: 'No religioso/a' },
  { value: 'religious', label: 'Religioso/a' },
  { value: 'very_religious', label: 'Muy religioso/a' },
];

export const STAR_SIGN_OPTIONS: Option[] = [
  { value: 'aquarius', label: 'Acuario' },
  { value: 'aries', label: 'Aries' },
  { value: 'cancer', label: 'Cáncer' },
  { value: 'capricorn', label: 'Capricornio' },
  { value: 'gemini', label: 'Géminis' },
  { value: 'leo', label: 'Leo' },
  { value: 'libra', label: 'Libra' },
  { value: 'pisces', label: 'Piscis' },
  { value: 'sagittarius', label: 'Sagitario' },
  { value: 'scorpio', label: 'Escorpio' },
  { value: 'taurus', label: 'Tauro' },
  { value: 'virgo', label: 'Virgo' },
];

// --- Über mich / estilo de vida (000013) ------------------------------

export const FUTURE_VISION_OPTIONS: Option[] = [
  { value: 'balance_family_career', label: 'Combinar familia y trabajo' },
  { value: 'focus_family_household', label: 'Centrarme en familia y hogar' },
  { value: 'beauty_and_partner_time', label: 'Cuidarme y estar con mi pareja' },
  { value: 'part_time_work', label: 'Trabajar a tiempo parcial' },
  { value: 'support_partner_career', label: 'Apoyar la carrera de mi pareja' },
  { value: 'new_education', label: 'Hacer una nueva formación' },
];

export const SPORTS_OPTIONS: Option[] = [
  { value: 'fitness', label: 'Fitness' },
  { value: 'motorsport', label: 'Motor' },
  { value: 'strength_training', label: 'Musculación' },
  { value: 'ball_sports', label: 'Deportes de balón' },
  { value: 'water_sports', label: 'Deportes acuáticos' },
  { value: 'jogging', label: 'Correr' },
  { value: 'winter_sports', label: 'Deportes de invierno' },
  { value: 'cycling', label: 'Ciclismo' },
  { value: 'athletics', label: 'Atletismo' },
  { value: 'climbing', label: 'Escalada' },
  { value: 'horse_riding', label: 'Equitación' },
  { value: 'hiking', label: 'Senderismo' },
  { value: 'other', label: 'Otro' },
];

export const LIKES_PETS_OPTIONS: Option[] = [
  { value: 'yes', label: 'Sí' },
  { value: 'neutral', label: 'Me da igual' },
  { value: 'no', label: 'No' },
];

export const PETS_OWNED_OPTIONS: Option[] = [
  { value: 'none', label: 'Ninguna' },
  { value: 'cat', label: 'Gato' },
  { value: 'dog', label: 'Perro' },
  { value: 'horse', label: 'Caballo' },
  { value: 'other', label: 'Otra' },
];

export const FAVORITE_SEASON_OPTIONS: Option[] = [
  { value: 'spring', label: 'Primavera' },
  { value: 'summer', label: 'Verano' },
  { value: 'autumn', label: 'Otoño' },
  { value: 'winter', label: 'Invierno' },
];

export const IDEAL_VACATION_STYLE_OPTIONS: Option[] = [
  { value: 'small_charming_hotel', label: 'Hotel pequeño y con encanto' },
  { value: 'luxury_hotel', label: 'Hotel de lujo' },
  { value: 'cruise_ship', label: 'Crucero' },
  { value: 'club_hotel', label: 'Hotel club' },
  { value: 'rental_apartment', label: 'Apartamento de alquiler' },
  { value: 'countryside_house', label: 'Casa rural' },
  { value: 'camping_rv', label: 'Camping / autocaravana' },
  { value: 'staying_home', label: 'Quedarme en casa' },
  { value: 'at_friends', label: 'En casa de amigos' },
];

export const VACATION_ACTIVITIES_OPTIONS: Option[] = [
  { value: 'cafes_shopping_nightlife', label: 'Cafés, compras y vida nocturna' },
  { value: 'mix_relaxation_activities', label: 'Mezcla de relax y actividades' },
  { value: 'lazing_and_relaxing', label: 'Descansar sin más' },
  { value: 'beach_holiday', label: 'Playa' },
  { value: 'sightseeing_cities', label: 'Visitar ciudades' },
  { value: 'lots_of_sports', label: 'Mucho deporte' },
];

// --- Preferencias de pareja ---------------------------------------------

export const DESIRED_TRAITS_OPTIONS: Option[] = [
  { value: 'humorous', label: 'Con sentido del humor' },
  { value: 'self_confident', label: 'Seguro/a de sí mismo/a' },
  { value: 'loving', label: 'Cariñoso/a' },
  { value: 'kind_hearted', label: 'De buen corazón' },
  { value: 'intelligent', label: 'Inteligente' },
  { value: 'faithful', label: 'Fiel' },
  { value: 'honest', label: 'Honesto/a' },
  { value: 'ambitious', label: 'Ambicioso/a' },
  { value: 'family_oriented', label: 'Orientado/a a la familia' },
  { value: 'adventurous', label: 'Aventurero/a' },
  { value: 'romantic', label: 'Romántico/a' },
  { value: 'patient', label: 'Paciente' },
  { value: 'easy_going', label: 'De trato fácil' },
  { value: 'financially_stable', label: 'Estable económicamente' },
  { value: 'other', label: 'Otro' },
];

export const PARTNER_MAY_HAVE_CHILDREN_OPTIONS: Option[] = [
  { value: 'yes', label: 'Sí' },
  { value: 'no', label: 'No' },
  { value: 'doesnt_matter', label: 'Me da igual' },
];

export const PARTNER_RELIGION_PREFERENCE_OPTIONS: Option[] = [
  ...RELIGION_OPTIONS,
  { value: 'doesnt_matter', label: 'Me da igual' },
];

export const FIRST_MEETING_PREFERENCE_OPTIONS: Option[] = [
  { value: 'doesnt_matter', label: 'Me da igual' },
  { value: 'public_place', label: 'Lugar público' },
  { value: 'my_city', label: 'En mi ciudad' },
  { value: 'their_city', label: 'En su ciudad' },
  { value: 'video_call_first', label: 'Videollamada primero' },
];

export const DESIRED_LIVING_PLACE_OPTIONS: Option[] = [
  { value: 'big_city', label: 'Gran ciudad' },
  { value: 'medium_city', label: 'Ciudad media' },
  { value: 'small_town', label: 'Pueblo' },
  { value: 'countryside', label: 'Campo' },
  { value: 'abroad', label: 'Extranjero' },
];

// Las 10 barras de "qué es importante en una relación". Es una lista
// fija y pequeña (decisión tomada en la Fase 1: no vive en un catálogo
// de BD), así que aquí también va como constante, no como fetch.
export const PARTNER_IMPORTANCE_FIELDS: { key: string; label: string }[] = [
  { key: 'importance_shared_thoughts', label: 'Compartir pensamientos' },
  { key: 'importance_shared_hobbies', label: 'Aficiones en común' },
  { key: 'importance_intimacy', label: 'Intimidad' },
  { key: 'importance_romantic_love', label: 'Amor romántico' },
  { key: 'importance_financial_security', label: 'Seguridad económica' },
  { key: 'importance_fun', label: 'Diversión' },
  { key: 'importance_shared_friends', label: 'Amigos en común' },
  { key: 'importance_shared_humor', label: 'Humor en común' },
  { key: 'importance_personal_space', label: 'Espacio personal' },
  { key: 'importance_independence', label: 'Independencia' },
];

// Las 5 etiquetas de rasgos de personalidad (Big Five). A diferencia de
// las afirmaciones dentro de cada rasgo (que sí vienen del catálogo del
// backend, porque esas sí crecen), los 5 rasgos son fijos.
export const PERSONALITY_TRAIT_LABELS: Record<string, string> = {
  extraversion: 'Extraversión',
  emotional_stability: 'Estabilidad emocional',
  conscientiousness: 'Meticulosidad',
  agreeableness: 'Amabilidad',
  openness: 'Apertura a experiencias',
};

export const HOBBY_CATEGORY_LABELS: Record<string, string> = {
  at_home: 'En casa',
  social: 'Social',
  creative: 'Creatividad',
  mobility: 'Movilidad y actividad',
  nature: 'Naturaleza',
  further_ed: 'Formación',
  going_out: 'Salir',
};
