'use client';

import { useEffect, useMemo, useRef, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';

import { apiFetch, InterestDefinition } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
import { sortedCountryOptions } from '@/lib/countryOptions';
import { loadSavedFilters, saveFilters, clearSavedFilters } from '@/lib/searchFilters';
import {
  HAS_CHILDREN_OPTIONS,
  WANTS_CHILDREN_OPTIONS,
  RELATIONSHIP_GOAL_OPTIONS,
  BODY_TYPE_OPTIONS,
  ETHNICITY_OPTIONS,
  APPEARANCE_RATING_OPTIONS,
  HAIR_COLOR_OPTIONS,
  EYE_COLOR_OPTIONS,
  BODY_ART_OPTIONS,
  SMOKING_HABIT_OPTIONS,
  DRINKING_HABIT_OPTIONS,
  RELOCATION_WILLINGNESS_OPTIONS,
  MARITAL_STATUS_OPTIONS,
  OCCUPATION_OPTIONS,
  EMPLOYMENT_STATUS_OPTIONS,
  INCOME_LEVEL_OPTIONS,
  LIVING_SITUATION_OPTIONS,
  EDUCATION_LEVEL_OPTIONS,
  ENGLISH_ABILITY_OPTIONS,
  RELIGION_OPTIONS,
  RELIGIOUS_VALUES_OPTIONS,
  STAR_SIGN_OPTIONS,
  FUTURE_VISION_OPTIONS,
  SPORTS_OPTIONS,
  LIKES_PETS_OPTIONS,
  PETS_OWNED_OPTIONS,
  FAVORITE_SEASON_OPTIONS,
  IDEAL_VACATION_STYLE_OPTIONS,
  VACATION_ACTIVITIES_OPTIONS,
  PERSONALITY_TRAIT_KEYS,
  INTEREST_CATEGORY_KEYS,
  LANGUAGE_OPTIONS,
} from '@/lib/profileOptions';
import styles from './page.module.css';

// No está en profileOptions.ts (viene del enum Gender del backend, no de un
// CHECK con lista fija en ese archivo) — se deja aquí como única excepción.
const GENDER_OPTIONS = ['female', 'male', 'non_binary', 'other'];

// Rangos numéricos del formulario. Todos son <select>, nunca <input>: el
// backend admite rangos más amplios (edad 18-120, altura 50-300cm, peso
// 20-400kg — ver search/types.go), pero para un desplegable seleccionable
// tiene más sentido acotar a un rango realista. Si algún caso real cae
// fuera de esto, se amplía aquí sin tocar nada más.
const MIN_SEARCH_AGE_UI = 18;
const MAX_SEARCH_AGE_UI = 99;
const MIN_HEIGHT_CM = 140;
const MAX_HEIGHT_CM = 210;
const MIN_WEIGHT_KG = 40;
const MAX_WEIGHT_KG = 150;
const MAX_CHILDREN_UI = 10;

// Nivel de interés/rasgo de personalidad: mismo 1-5 que valida el backend
// (search.MinInterestLevel/MaxInterestLevel, profiles.MinPersonalityScore/
// MaxPersonalityScore). La personalidad admite medios puntos (es la media
// de varias respuestas 1-5); el nivel de interés es un entero.
const INTEREST_LEVEL_OPTIONS = ['1', '2', '3', '4', '5'];
const PERSONALITY_SCORE_OPTIONS = ['1', '1.5', '2', '2.5', '3', '3.5', '4', '4.5', '5'];

function integerRange(start: number, end: number): string[] {
  const out: string[] = [];
  for (let n = start; n <= end; n += 1) out.push(String(n));
  return out;
}

const AGE_OPTIONS = integerRange(MIN_SEARCH_AGE_UI, MAX_SEARCH_AGE_UI);
const HEIGHT_OPTIONS = integerRange(MIN_HEIGHT_CM, MAX_HEIGHT_CM);
const WEIGHT_OPTIONS = integerRange(MIN_WEIGHT_KG, MAX_WEIGHT_KG);
const MAX_CHILDREN_OPTIONS = integerRange(0, MAX_CHILDREN_UI);

type Comparator = 'gte' | 'lte' | 'eq' | 'any';

// Fila de comparador reutilizada por Intereses-con-nivel y Personalidad:
// "campo" >= / <= / = "valor" (o "any" = solo pertenencia, sin exigir
// valor — solo tiene sentido para intereses, no para personalidad).
interface ComparatorRow {
  comparator: Comparator | 'none';
  value: string;
}

// Convierte los límites guardados (?interest_x_min / _max, ?trait_x_min / _max)
// en la fila del formulario. Min y max iguales = "=", solo min = ">=", solo
// max = "<=". Un rango con min y max distintos no cabe en un único comparador:
// se conserva el mínimo (">=").
function rowFromBounds(min: string | null, max: string | null): ComparatorRow {
  const lo = (min ?? '').trim();
  const hi = (max ?? '').trim();
  if (lo && hi && lo === hi) return { comparator: 'eq', value: lo };
  if (lo) return { comparator: 'gte', value: lo };
  if (hi) return { comparator: 'lte', value: hi };
  return { comparator: 'none', value: '' };
}

// -----------------------------------------------------------------------------
// Piezas reutilizables del formulario
// -----------------------------------------------------------------------------
function SelectField({
  label,
  value,
  onChange,
  options,
  labels,
  anyLabel,
}: {
  label: string;
  value: string;
  onChange: (v: string) => void;
  options: string[];
  labels: Record<string, string>;
  anyLabel: string;
}) {
  return (
    <div className={styles.field}>
      <label className={styles.label}>{label}</label>
      <select className={styles.select} value={value} onChange={(e) => onChange(e.target.value)}>
        <option value="">{anyLabel}</option>
        {options.map((opt) => (
          <option key={opt} value={opt}>
            {labels?.[opt] ?? opt}
          </option>
        ))}
      </select>
    </div>
  );
}

function ChipGroup({
  label,
  values,
  onToggle,
  options,
  labels,
}: {
  label: string;
  values: string[];
  onToggle: (opt: string) => void;
  options: string[];
  labels: Record<string, string>;
}) {
  return (
    <div className={styles.field}>
      <label className={styles.label}>{label}</label>
      <div className={styles.chipGroup}>
        {options.map((opt) => {
          const active = values.includes(opt);
          return (
            <button
              type="button"
              key={opt}
              className={active ? `${styles.chip} ${styles.chipActive}` : styles.chip}
              onClick={() => onToggle(opt)}
            >
              {labels?.[opt] ?? opt}
            </button>
          );
        })}
      </div>
    </div>
  );
}

// Selector de rango con dos <select> (mín. / máx.) en vez de dos <input
// type="number">: mismo layout que antes (.rangeRow / .rangeSep), pero
// 100% seleccionable, nada de escritura.
function RangeSelect({
  label,
  minValue,
  maxValue,
  onMinChange,
  onMaxChange,
  options,
  anyLabel,
}: {
  label: string;
  minValue: string;
  maxValue: string;
  onMinChange: (v: string) => void;
  onMaxChange: (v: string) => void;
  options: string[];
  anyLabel: string;
}) {
  return (
    <div className={styles.field}>
      <label className={styles.label}>{label}</label>
      <div className={styles.rangeRow}>
        <select className={styles.select} value={minValue} onChange={(e) => onMinChange(e.target.value)}>
          <option value="">{anyLabel}</option>
          {options.map((opt) => (
            <option key={opt} value={opt}>
              {opt}
            </option>
          ))}
        </select>
        <span className={styles.rangeSep}>–</span>
        <select className={styles.select} value={maxValue} onChange={(e) => onMaxChange(e.target.value)}>
          <option value="">{anyLabel}</option>
          {options.map((opt) => (
            <option key={opt} value={opt}>
              {opt}
            </option>
          ))}
        </select>
      </div>
    </div>
  );
}

// Fila de comparador (usada para Intereses-con-nivel y Personalidad):
// "campo" >= / <= / = "valor"
function ComparatorSelect({
  value,
  onChange,
  dictionary,
  includeAny,
}: {
  value: Comparator | 'none';
  onChange: (v: Comparator | 'none') => void;
  dictionary: Dictionary;
  includeAny: boolean;
}) {
  return (
    <select className={styles.select} value={value} onChange={(e) => onChange(e.target.value as Comparator | 'none')}>
      {includeAny && <option value="any">{dictionary.search.comparatorAny}</option>}
      {!includeAny && <option value="none">{dictionary.search.comparatorNone}</option>}
      <option value="gte">{dictionary.search.comparatorGte}</option>
      <option value="lte">{dictionary.search.comparatorLte}</option>
      <option value="eq">{dictionary.search.comparatorEq}</option>
    </select>
  );
}

// -----------------------------------------------------------------------------
// Página
// -----------------------------------------------------------------------------
export default function SearchPage() {
  const router = useRouter();
  const { dictionary, locale } = useI18n();

  // Hydration: locale depende de Intl en el navegador, y sortedCountryOptions
  // usa Intl.DisplayNames. Para que el SSR y el primer render del cliente
  // coincidan, no renderizamos los selects de país/nacionalidad en absoluto
  // hasta que el componente se monta en el cliente.
  const [isClient, setIsClient] = useState(false);

  useEffect(() => {
    setIsClient(true);
  }, []);

  // --- Rangos numéricos (todo <select>, ver constantes arriba) ---
  const [minAge, setMinAge] = useState('');
  const [maxAge, setMaxAge] = useState('');
  const [country, setCountry] = useState('');
  const [nationality, setNationality] = useState('');
  const [minHeight, setMinHeight] = useState('');
  const [maxHeight, setMaxHeight] = useState('');
  const [minWeight, setMinWeight] = useState('');
  const [maxWeight, setMaxWeight] = useState('');
  const [maxChildren, setMaxChildren] = useState('');

  // --- Selects de un solo valor ---
  const [hasChildren, setHasChildren] = useState('');
  const [wantsChildren, setWantsChildren] = useState('');
  const [bodyType, setBodyType] = useState('');
  const [ethnicity, setEthnicity] = useState('');
  const [appearanceRating, setAppearanceRating] = useState('');
  const [hairColor, setHairColor] = useState('');
  const [eyeColor, setEyeColor] = useState('');
  const [smokingHabit, setSmokingHabit] = useState('');
  const [drinkingHabit, setDrinkingHabit] = useState('');
  const [maritalStatus, setMaritalStatus] = useState('');
  const [occupation, setOccupation] = useState('');
  const [employmentStatus, setEmploymentStatus] = useState('');
  const [incomeLevel, setIncomeLevel] = useState('');
  const [livingSituation, setLivingSituation] = useState('');
  const [educationLevel, setEducationLevel] = useState('');
  const [englishAbility, setEnglishAbility] = useState('');
  const [religion, setReligion] = useState('');
  const [religiousValues, setReligiousValues] = useState('');
  const [starSign, setStarSign] = useState('');
  const [likesPets, setLikesPets] = useState('');
  const [favoriteSeason, setFavoriteSeason] = useState('');
  const [sort, setSort] = useState('');

  // --- Multi-selección (chips) ---
  const [genders, setGenders] = useState<string[]>([]);
  const [relationshipGoals, setRelationshipGoals] = useState<string[]>([]);
  const [languages, setLanguages] = useState<string[]>([]);
  const [bodyArt, setBodyArt] = useState<string[]>([]);
  const [relocationWillingness, setRelocationWillingness] = useState<string[]>([]);
  const [futureVision, setFutureVision] = useState<string[]>([]);
  const [sports, setSports] = useState<string[]>([]);
  const [petsOwned, setPetsOwned] = useState<string[]>([]);
  const [idealVacationStyle, setIdealVacationStyle] = useState<string[]>([]);
  const [vacationActivities, setVacationActivities] = useState<string[]>([]);
  const [selectedInterests, setSelectedInterests] = useState<string[]>([]);

  // --- País / nacionalidad: solo se calcula y se renderiza en el cliente. ---
  const countryOptions = useMemo(
    () => (isClient ? sortedCountryOptions(locale) : []),
    [isClient, locale],
  );

  // --- Catálogo de intereses (GET /interests) ---
  const [interests, setInterests] = useState<InterestDefinition[]>([]);
  const [interestsError, setInterestsError] = useState(false);

  // --- Nivel de interés, solo para los del catálogo con has_level=true.
  // Se siembra en cuanto llega el catálogo (una fila "sin filtrar" por
  // cada clave con nivel). El resto de intereses (has_level=false) siguen
  // siendo un simple chip on/off en selectedInterests. ---
  const [interestLevels, setInterestLevels] = useState<Record<string, ComparatorRow>>({});

  useEffect(() => {
    apiFetch<InterestDefinition[]>('/interests')
      .then((data) => {
        setInterests(data);
        setInterestLevels(
          Object.fromEntries(
            data.filter((d) => d.has_level).map((d) => [d.key, { comparator: 'none', value: '' } as ComparatorRow]),
          ),
        );
      })
      .catch(() => setInterestsError(true));
  }, []);

  // --- Personalidad (5 rasgos fijos, cada uno con su comparador) ---
  const [personality, setPersonality] = useState<Record<string, ComparatorRow>>(
    Object.fromEntries(PERSONALITY_TRAIT_KEYS.map((k) => [k, { comparator: 'none', value: '' }])),
  );

  // --- Rellenar el formulario con los filtros guardados --------------------
  // Búsqueda es quien guarda los filtros (al enviar) y los muestra al entrar;
  // Discover y Quick Match los leen del mismo sitio (lib/searchFilters.ts).
  const [savedQuery, setSavedQuery] = useState<string | null>(null);
  useEffect(() => {
    let alive = true;
    loadSavedFilters().then((q) => {
      if (alive) setSavedQuery(q);
    });
    return () => {
      alive = false;
    };
  }, []);

  // Campos simples, listas y personalidad (no dependen del catálogo).
  useEffect(() => {
    if (savedQuery === null) return;
    const p = new URLSearchParams(savedQuery);
    const one = (key: string) => p.get(key) ?? '';
    const many = (key: string) =>
      (p.get(key) ?? '')
        .split(',')
        .map((v) => v.trim())
        .filter(Boolean);

    setMinAge(one('min_age')); setMaxAge(one('max_age'));
    setCountry(one('country')); setNationality(one('nationality'));
    setMinHeight(one('min_height')); setMaxHeight(one('max_height'));
    setMinWeight(one('min_weight')); setMaxWeight(one('max_weight'));
    setMaxChildren(one('max_children'));
    setHasChildren(one('has_children')); setWantsChildren(one('wants_children'));
    setBodyType(one('body_type')); setEthnicity(one('ethnicity'));
    setAppearanceRating(one('appearance_rating'));
    setHairColor(one('hair_color')); setEyeColor(one('eye_color'));
    setSmokingHabit(one('smoking_habit')); setDrinkingHabit(one('drinking_habit'));
    setMaritalStatus(one('marital_status')); setOccupation(one('occupation'));
    setEmploymentStatus(one('employment_status')); setIncomeLevel(one('income_level'));
    setLivingSituation(one('living_situation'));
    setEducationLevel(one('education_level')); setEnglishAbility(one('english_ability'));
    setReligion(one('religion')); setReligiousValues(one('religious_values'));
    setStarSign(one('star_sign'));
    setLikesPets(one('likes_pets')); setFavoriteSeason(one('favorite_season'));
    setSort(one('sort'));

    setGenders(many('gender'));
    setRelationshipGoals(many('relationship_goals'));
    setLanguages(many('language'));
    setBodyArt(many('body_art'));
    setRelocationWillingness(many('relocation_willingness'));
    setFutureVision(many('future_vision'));
    setSports(many('sports'));
    setPetsOwned(many('pets_owned'));
    setIdealVacationStyle(many('ideal_vacation_style'));
    setVacationActivities(many('vacation_activities'));

    setPersonality(
      Object.fromEntries(
        PERSONALITY_TRAIT_KEYS.map((trait) => [trait, rowFromBounds(p.get(`trait_${trait}_min`), p.get(`trait_${trait}_max`))]),
      ),
    );
  }, [savedQuery]);

  // Intereses: necesitan el catálogo (para saber cuáles llevan nivel), así que
  // se rellenan cuando llegan tanto el catálogo como los filtros guardados.
  const interestsHydrated = useRef(false);
  useEffect(() => {
    if (savedQuery === null || interests.length === 0 || interestsHydrated.current) return;
    interestsHydrated.current = true;

    const p = new URLSearchParams(savedQuery);
    const bare = (p.get('interests') ?? '')
      .split(',')
      .map((v) => v.trim())
      .filter(Boolean);
    const leveled = interests.filter((i) => i.has_level);
    const leveledKeys = new Set(leveled.map((i) => i.key));

    setSelectedInterests(bare.filter((key) => !leveledKeys.has(key)));
    setInterestLevels(
      Object.fromEntries(
        leveled.map((i) => [
          i.key,
          bare.includes(i.key)
            ? ({ comparator: 'any', value: '' } as ComparatorRow)
            : rowFromBounds(p.get(`interest_${i.key}_min`), p.get(`interest_${i.key}_max`)),
        ]),
      ),
    );
  }, [savedQuery, interests]);

  function toggleValue(list: string[], setList: (v: string[]) => void, value: string) {
    setList(list.includes(value) ? list.filter((v) => v !== value) : [...list, value]);
  }

  // Al elegir un comparador que necesita valor (gte/lte/eq) y todavía no
  // hay ninguno seleccionado, arrancamos en el punto medio de la escala:
  // así la fila sigue siendo "solo elegir", nunca hace falta escribir.
  function updateComparatorRow(
    setRows: React.Dispatch<React.SetStateAction<Record<string, ComparatorRow>>>,
    key: string,
    patch: Partial<ComparatorRow>,
    midpoint: string,
  ) {
    setRows((prev) => {
      const current = prev[key] ?? { comparator: 'none', value: '' };
      const next: ComparatorRow = { ...current, ...patch };
      if ((next.comparator === 'gte' || next.comparator === 'lte' || next.comparator === 'eq') && next.value === '') {
        next.value = midpoint;
      }
      return { ...prev, [key]: next };
    });
  }

  function updateInterestLevel(key: string, patch: Partial<ComparatorRow>) {
    updateComparatorRow(setInterestLevels, key, patch, '3');
  }

  function updatePersonality(trait: string, patch: Partial<ComparatorRow>) {
    updateComparatorRow(setPersonality, trait, patch, '3');
  }

  function buildParams(): URLSearchParams {
    const params = new URLSearchParams();
    const setIf = (key: string, value: string) => {
      if (value.trim() !== '') params.set(key, value.trim());
    };
    const setMulti = (key: string, values: string[]) => {
      if (values.length > 0) params.set(key, values.join(','));
    };

    setMulti('gender', genders);
    setIf('min_age', minAge);
    setIf('max_age', maxAge);
    setIf('country', country);
    setMulti('language', languages);
    setMulti('relationship_goals', relationshipGoals);
    setIf('has_children', hasChildren);
    setIf('wants_children', wantsChildren);

    setIf('min_height', minHeight);
    setIf('max_height', maxHeight);
    setIf('min_weight', minWeight);
    setIf('max_weight', maxWeight);
    setIf('body_type', bodyType);
    setIf('ethnicity', ethnicity);
    setIf('appearance_rating', appearanceRating);
    setIf('hair_color', hairColor);
    setIf('eye_color', eyeColor);
    setMulti('body_art', bodyArt);

    setIf('smoking_habit', smokingHabit);
    setIf('drinking_habit', drinkingHabit);
    setMulti('relocation_willingness', relocationWillingness);
    setIf('marital_status', maritalStatus);
    setIf('max_children', maxChildren);
    setIf('occupation', occupation);
    setIf('employment_status', employmentStatus);
    setIf('income_level', incomeLevel);
    setIf('living_situation', livingSituation);

    setIf('nationality', nationality);
    setIf('education_level', educationLevel);
    setIf('english_ability', englishAbility);
    setIf('religion', religion);
    setIf('religious_values', religiousValues);
    setIf('star_sign', starSign);

    setMulti('future_vision', futureVision);
    setMulti('sports', sports);
    setIf('likes_pets', likesPets);
    setMulti('pets_owned', petsOwned);
    setIf('favorite_season', favoriteSeason);
    setMulti('ideal_vacation_style', idealVacationStyle);
    setMulti('vacation_activities', vacationActivities);

    // Intereses: los que no llevan nivel (has_level=false) son pertenencia
    // simple, igual que antes. Los que sí llevan nivel se pliegan también
    // en "interests" cuando el comparador es "any" (solo que le guste, sin
    // exigir nivel); si el comparador exige valor (gte/lte/eq) generan
    // ?interest_{key}_min / _max en su lugar, igual que trait_{key}_*.
    const bareInterests = [...selectedInterests];
    Object.entries(interestLevels).forEach(([key, row]) => {
      if (row.comparator === 'any') {
        bareInterests.push(key);
      } else if (row.comparator !== 'none' && row.value.trim() !== '') {
        if (row.comparator === 'gte' || row.comparator === 'eq') params.set(`interest_${key}_min`, row.value.trim());
        if (row.comparator === 'lte' || row.comparator === 'eq') params.set(`interest_${key}_max`, row.value.trim());
      }
    });
    setMulti('interests', bareInterests);

    // Personalidad: mismo patrón, sin la opción "any" (no tiene sentido
    // pedir un rasgo sin ningún límite).
    PERSONALITY_TRAIT_KEYS.forEach((trait) => {
      const row = personality[trait];
      if (!row || row.comparator === 'none' || row.value.trim() === '') return;
      if (row.comparator === 'gte' || row.comparator === 'eq') params.set(`trait_${trait}_min`, row.value.trim());
      if (row.comparator === 'lte' || row.comparator === 'eq') params.set(`trait_${trait}_max`, row.value.trim());
    });

    if (sort) params.set('sort', sort);

    return params;
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const query = buildParams().toString();
    // Se guarda antes de navegar (vacío = borra los filtros guardados), para
    // que Discover y Quick Match ya lo encuentren al cargar.
    await saveFilters(query);
    router.push(`/discover?${query}`);
  }

  function handleReset() {
    void clearSavedFilters();
    setMinAge(''); setMaxAge(''); setCountry(''); setNationality('');
    setMinHeight(''); setMaxHeight(''); setMinWeight(''); setMaxWeight(''); setMaxChildren('');
    setHasChildren(''); setWantsChildren(''); setBodyType(''); setEthnicity(''); setAppearanceRating('');
    setHairColor(''); setEyeColor(''); setSmokingHabit(''); setDrinkingHabit(''); setMaritalStatus('');
    setOccupation(''); setEmploymentStatus(''); setIncomeLevel(''); setLivingSituation('');
    setEducationLevel(''); setEnglishAbility(''); setReligion(''); setReligiousValues(''); setStarSign('');
    setLikesPets(''); setFavoriteSeason(''); setSort('');
    setGenders([]); setRelationshipGoals([]); setLanguages([]); setBodyArt([]);
    setRelocationWillingness([]); setFutureVision([]); setSports([]); setPetsOwned([]);
    setIdealVacationStyle([]); setVacationActivities([]); setSelectedInterests([]);
    setInterestLevels((prev) => Object.fromEntries(Object.keys(prev).map((k) => [k, { comparator: 'none', value: '' }])));
    setPersonality(Object.fromEntries(PERSONALITY_TRAIT_KEYS.map((k) => [k, { comparator: 'none', value: '' }])));
  }

  const interestsByCategory = INTEREST_CATEGORY_KEYS.map((cat) => ({
    key: cat,
    items: interests.filter((i) => i.category === cat),
  })).filter((g) => g.items.length > 0);

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.search.title}</h1>
      <p className={styles.subtitle}>{dictionary.search.subtitle}</p>

      <form onSubmit={handleSubmit}>
        {/* --- Básico --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionBasic}</h2>
          <div className={styles.fieldGrid}>
            <ChipGroup label={dictionary.search.labelGender} values={genders} onToggle={(v) => toggleValue(genders, setGenders, v)} options={GENDER_OPTIONS} labels={dictionary.options.gender} />
            <RangeSelect
              label={dictionary.search.labelAge}
              minValue={minAge}
              maxValue={maxAge}
              onMinChange={setMinAge}
              onMaxChange={setMaxAge}
              options={AGE_OPTIONS}
              anyLabel={dictionary.search.any}
            />
            {isClient ? (
              <div className={styles.field}>
                <label className={styles.label}>{dictionary.search.labelCountry}</label>
                <select className={styles.select} value={country} onChange={(e) => setCountry(e.target.value)}>
                  <option value="">{dictionary.search.any}</option>
                  {countryOptions.map((c) => (
                    <option key={c.code} value={c.code}>
                      {c.label}
                    </option>
                  ))}
                </select>
              </div>
            ) : null}
            <ChipGroup label={dictionary.search.labelRelationshipGoal} values={relationshipGoals} onToggle={(v) => toggleValue(relationshipGoals, setRelationshipGoals, v)} options={RELATIONSHIP_GOAL_OPTIONS} labels={dictionary.options.relationshipGoal} />
            <SelectField label={dictionary.search.labelHasChildren} value={hasChildren} onChange={setHasChildren} options={HAS_CHILDREN_OPTIONS} labels={dictionary.options.hasChildren} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelWantsChildren} value={wantsChildren} onChange={setWantsChildren} options={WANTS_CHILDREN_OPTIONS} labels={dictionary.options.wantsChildren} anyLabel={dictionary.search.any} />
          </div>

          {/* Idiomas */}
          <div style={{ marginTop: '1.1rem' }}>
            <ChipGroup
              label={dictionary.search.labelLanguages}
              values={languages}
              onToggle={(v) => toggleValue(languages, setLanguages, v)}
              options={LANGUAGE_OPTIONS}
              labels={dictionary.options.languages}
            />
          </div>
        </section>

        {/* --- Físico --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionPhysical}</h2>
          <div className={styles.fieldGrid}>
            <RangeSelect label={dictionary.search.labelHeight} minValue={minHeight} maxValue={maxHeight} onMinChange={setMinHeight} onMaxChange={setMaxHeight} options={HEIGHT_OPTIONS} anyLabel={dictionary.search.any} />
            <RangeSelect label={dictionary.search.labelWeight} minValue={minWeight} maxValue={maxWeight} onMinChange={setMinWeight} onMaxChange={setMaxWeight} options={WEIGHT_OPTIONS} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelBodyType} value={bodyType} onChange={setBodyType} options={BODY_TYPE_OPTIONS} labels={dictionary.options.bodyType} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelEthnicity} value={ethnicity} onChange={setEthnicity} options={ETHNICITY_OPTIONS} labels={dictionary.options.ethnicity} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelAppearance} value={appearanceRating} onChange={setAppearanceRating} options={APPEARANCE_RATING_OPTIONS} labels={dictionary.options.appearanceRating} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelHairColor} value={hairColor} onChange={setHairColor} options={HAIR_COLOR_OPTIONS} labels={dictionary.options.hairColor} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelEyeColor} value={eyeColor} onChange={setEyeColor} options={EYE_COLOR_OPTIONS} labels={dictionary.options.eyeColor} anyLabel={dictionary.search.any} />
            <ChipGroup label={dictionary.search.labelBodyArt} values={bodyArt} onToggle={(v) => toggleValue(bodyArt, setBodyArt, v)} options={BODY_ART_OPTIONS} labels={dictionary.options.bodyArt} />
          </div>
        </section>

        {/* --- Estilo de vida y familia --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionLifestyle}</h2>
          <div className={styles.fieldGrid}>
            <SelectField label={dictionary.search.labelSmoking} value={smokingHabit} onChange={setSmokingHabit} options={SMOKING_HABIT_OPTIONS} labels={dictionary.options.smokingHabit} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelDrinking} value={drinkingHabit} onChange={setDrinkingHabit} options={DRINKING_HABIT_OPTIONS} labels={dictionary.options.drinkingHabit} anyLabel={dictionary.search.any} />
            <ChipGroup label={dictionary.search.labelRelocation} values={relocationWillingness} onToggle={(v) => toggleValue(relocationWillingness, setRelocationWillingness, v)} options={RELOCATION_WILLINGNESS_OPTIONS} labels={dictionary.options.relocationWillingness} />
            <SelectField label={dictionary.search.labelMaritalStatus} value={maritalStatus} onChange={setMaritalStatus} options={MARITAL_STATUS_OPTIONS} labels={dictionary.options.maritalStatus} anyLabel={dictionary.search.any} />
            <div className={styles.field}>
              <label className={styles.label}>{dictionary.search.labelMaxChildren}</label>
              <select className={styles.select} value={maxChildren} onChange={(e) => setMaxChildren(e.target.value)}>
                <option value="">{dictionary.search.any}</option>
                {MAX_CHILDREN_OPTIONS.map((n) => (
                  <option key={n} value={n}>
                    {n}
                  </option>
                ))}
              </select>
            </div>
            <SelectField label={dictionary.search.labelOccupation} value={occupation} onChange={setOccupation} options={OCCUPATION_OPTIONS} labels={dictionary.options.occupation} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelEmploymentStatus} value={employmentStatus} onChange={setEmploymentStatus} options={EMPLOYMENT_STATUS_OPTIONS} labels={dictionary.options.employmentStatus} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelIncomeLevel} value={incomeLevel} onChange={setIncomeLevel} options={INCOME_LEVEL_OPTIONS} labels={dictionary.options.incomeLevel} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelLivingSituation} value={livingSituation} onChange={setLivingSituation} options={LIVING_SITUATION_OPTIONS} labels={dictionary.options.livingSituation} anyLabel={dictionary.search.any} />
          </div>
        </section>

        {/* --- Fondo, cultura y valores --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionBackground}</h2>
          <div className={styles.fieldGrid}>
            {isClient ? (
              <div className={styles.field}>
                <label className={styles.label}>{dictionary.search.labelNationality}</label>
                <select className={styles.select} value={nationality} onChange={(e) => setNationality(e.target.value)}>
                  <option value="">{dictionary.search.any}</option>
                  {countryOptions.map((c) => (
                    <option key={c.code} value={c.code}>
                      {c.label}
                    </option>
                  ))}
                </select>
              </div>
            ) : null}
            <SelectField label={dictionary.search.labelEducationLevel} value={educationLevel} onChange={setEducationLevel} options={EDUCATION_LEVEL_OPTIONS} labels={dictionary.options.educationLevel} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelEnglishAbility} value={englishAbility} onChange={setEnglishAbility} options={ENGLISH_ABILITY_OPTIONS} labels={dictionary.options.englishAbility} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelReligion} value={religion} onChange={setReligion} options={RELIGION_OPTIONS} labels={dictionary.options.religion} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelReligiousValues} value={religiousValues} onChange={setReligiousValues} options={RELIGIOUS_VALUES_OPTIONS} labels={dictionary.options.religiousValues} anyLabel={dictionary.search.any} />
            <SelectField label={dictionary.search.labelStarSign} value={starSign} onChange={setStarSign} options={STAR_SIGN_OPTIONS} labels={dictionary.options.starSign} anyLabel={dictionary.search.any} />
          </div>
        </section>

        {/* --- Estilo de vida adicional --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionLifestyleExtra}</h2>
          <div className={styles.fieldGrid}>
            <ChipGroup label={dictionary.search.labelFutureVision} values={futureVision} onToggle={(v) => toggleValue(futureVision, setFutureVision, v)} options={FUTURE_VISION_OPTIONS} labels={dictionary.options.futureVision} />
            <ChipGroup label={dictionary.search.labelSports} values={sports} onToggle={(v) => toggleValue(sports, setSports, v)} options={SPORTS_OPTIONS} labels={dictionary.options.sports} />
            <SelectField label={dictionary.search.labelLikesPets} value={likesPets} onChange={setLikesPets} options={LIKES_PETS_OPTIONS} labels={dictionary.options.likesPets} anyLabel={dictionary.search.any} />
            <ChipGroup label={dictionary.search.labelPetsOwned} values={petsOwned} onToggle={(v) => toggleValue(petsOwned, setPetsOwned, v)} options={PETS_OWNED_OPTIONS} labels={dictionary.options.petsOwned} />
            <SelectField label={dictionary.search.labelFavoriteSeason} value={favoriteSeason} onChange={setFavoriteSeason} options={FAVORITE_SEASON_OPTIONS} labels={dictionary.options.favoriteSeason} anyLabel={dictionary.search.any} />
            <ChipGroup label={dictionary.search.labelIdealVacationStyle} values={idealVacationStyle} onToggle={(v) => toggleValue(idealVacationStyle, setIdealVacationStyle, v)} options={IDEAL_VACATION_STYLE_OPTIONS} labels={dictionary.options.idealVacationStyle} />
            <ChipGroup label={dictionary.search.labelVacationActivities} values={vacationActivities} onToggle={(v) => toggleValue(vacationActivities, setVacationActivities, v)} options={VACATION_ACTIVITIES_OPTIONS} labels={dictionary.options.vacationActivities} />
          </div>
        </section>

        {/* --- Intereses (pertenencia; y comparador </o> de nivel para los
             que el catálogo marca con has_level=true, p.ej. "viajes >4" o
             "deportes <3") --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionInterests}</h2>
          {interestsError && <p className={styles.freeTextWarning}>{dictionary.search.interestsLoadError}</p>}
          {interestsByCategory.map((group) => {
            const simpleItems = group.items.filter((i) => !i.has_level);
            const leveledItems = group.items.filter((i) => i.has_level);
            return (
              <div key={group.key}>
                <p className={styles.categoryLabel}>{dictionary.options.interestCategories?.[group.key as keyof typeof dictionary.options.interestCategories] ?? group.key}</p>

                {simpleItems.length > 0 && (
                  <div className={styles.chipGroup}>
                    {simpleItems.map((item) => {
                      const active = selectedInterests.includes(item.key);
                      return (
                        <button
                          type="button"
                          key={item.key}
                          className={active ? `${styles.chip} ${styles.chipActive}` : styles.chip}
                          onClick={() => toggleValue(selectedInterests, setSelectedInterests, item.key)}
                        >
                          {item.label}
                        </button>
                      );
                    })}
                  </div>
                )}

                {leveledItems.map((item) => {
                  const row = interestLevels[item.key] ?? { comparator: 'none', value: '' };
                  const needsValue = row.comparator === 'gte' || row.comparator === 'lte' || row.comparator === 'eq';
                  return (
                    <div key={item.key} className={styles.comparatorRow}>
                      <span className={styles.label}>{item.label}</span>
                      <ComparatorSelect
                        value={row.comparator}
                        onChange={(c) => updateInterestLevel(item.key, { comparator: c })}
                        dictionary={dictionary}
                        includeAny
                      />
                      {needsValue ? (
                        <select
                          className={styles.select}
                          aria-label={dictionary.search.interestLevelAriaLabel}
                          value={row.value}
                          onChange={(e) => updateInterestLevel(item.key, { value: e.target.value })}
                        >
                          {INTEREST_LEVEL_OPTIONS.map((n) => (
                            <option key={n} value={n}>
                              {n}
                            </option>
                          ))}
                        </select>
                      ) : (
                        <span />
                      )}
                      <span />
                    </div>
                  );
                })}
              </div>
            );
          })}
        </section>

        {/* --- Personalidad (Big Five, comparador) --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionPersonality}</h2>
          {PERSONALITY_TRAIT_KEYS.map((trait) => {
            const row = personality[trait] ?? { comparator: 'none', value: '' };
            const needsValue = row.comparator === 'gte' || row.comparator === 'lte' || row.comparator === 'eq';
            return (
              <div key={trait} className={styles.comparatorRow}>
                <span className={styles.label}>{dictionary.options.personalityTraits?.[trait as keyof typeof dictionary.options.personalityTraits] ?? trait}</span>
                <ComparatorSelect
                  value={row.comparator}
                  onChange={(c) => updatePersonality(trait, { comparator: c })}
                  dictionary={dictionary}
                  includeAny={false}
                />
                {needsValue ? (
                  <select
                    className={styles.select}
                    value={row.value}
                    onChange={(e) => updatePersonality(trait, { value: e.target.value })}
                  >
                    {PERSONALITY_SCORE_OPTIONS.map((n) => (
                      <option key={n} value={n}>
                        {n}
                      </option>
                    ))}
                  </select>
                ) : (
                  <span />
                )}
                <span />
              </div>
            );
          })}
        </section>

        {/* --- Orden --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionSort}</h2>
          <div className={styles.fieldGrid}>
            <div className={styles.field}>
              <label className={styles.label}>{dictionary.search.labelSort}</label>
              <select className={styles.select} value={sort} onChange={(e) => setSort(e.target.value)}>
                <option value="">{dictionary.search.sortRecent}</option>
                <option value="age_asc">{dictionary.search.sortAgeAsc}</option>
                <option value="age_desc">{dictionary.search.sortAgeDesc}</option>
              </select>
            </div>
          </div>
        </section>

        <div className={styles.actionsRow}>
          <button type="button" className={styles.resetBtn} onClick={handleReset}>
            {dictionary.search.reset}
          </button>
          <button type="submit" className={styles.submitBtn}>
            {dictionary.search.submit}
          </button>
        </div>
      </form>
    </main>
  );
}
