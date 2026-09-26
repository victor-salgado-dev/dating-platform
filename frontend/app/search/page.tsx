'use client';

import { useEffect, useState, type FormEvent } from 'react';
import { useRouter } from 'next/navigation';

import { apiFetch, InterestDefinition } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
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

type Comparator = 'gte' | 'lte' | 'eq' | 'any';

interface HobbyRow {
  key: string;
  comparator: Comparator;
  value: string;
}

interface PersonalityRow {
  comparator: Comparator | 'none';
  value: string;
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

function RangeField({
  label,
  minValue,
  maxValue,
  onMinChange,
  onMaxChange,
  placeholderMin,
  placeholderMax,
}: {
  label: string;
  minValue: string;
  maxValue: string;
  onMinChange: (v: string) => void;
  onMaxChange: (v: string) => void;
  placeholderMin: string;
  placeholderMax: string;
}) {
  return (
    <div className={styles.field}>
      <label className={styles.label}>{label}</label>
      <div className={styles.rangeRow}>
        <input
          type="number"
          className={styles.input}
          placeholder={placeholderMin}
          value={minValue}
          onChange={(e) => onMinChange(e.target.value)}
        />
        <span className={styles.rangeSep}>–</span>
        <input
          type="number"
          className={styles.input}
          placeholder={placeholderMax}
          value={maxValue}
          onChange={(e) => onMaxChange(e.target.value)}
        />
      </div>
    </div>
  );
}

// Fila de comparador (usada para Hobbies e Personalidad): "campo" >= / <= / = "valor"
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
  const { dictionary } = useI18n();

  // --- Campos simples (texto/número) ---
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

  // --- Catálogo de intereses (endpoint asumido: GET /interests) ---
  const [interests, setInterests] = useState<InterestDefinition[]>([]);
  const [interestsError, setInterestsError] = useState(false);

  useEffect(() => {
    apiFetch<InterestDefinition[]>('/interests')
      .then(setInterests)
      .catch(() => setInterestsError(true));
  }, []);

  // --- Hobbies (dinámico, con comparador — no hay catálogo confirmado
  // todavía, así que la clave es texto libre por ahora) ---
  const [hobbies, setHobbies] = useState<HobbyRow[]>([]);

  // --- Personalidad (5 rasgos fijos, cada uno con su comparador) ---
  const [personality, setPersonality] = useState<Record<string, PersonalityRow>>(
    Object.fromEntries(PERSONALITY_TRAIT_KEYS.map((k) => [k, { comparator: 'none', value: '' }]))
  );

  function toggleValue(list: string[], setList: (v: string[]) => void, value: string) {
    setList(list.includes(value) ? list.filter((v) => v !== value) : [...list, value]);
  }

  function addHobbyRow() {
    setHobbies((rows) => [...rows, { key: '', comparator: 'any', value: '' }]);
  }
  function updateHobbyRow(index: number, patch: Partial<HobbyRow>) {
    setHobbies((rows) => rows.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  }
  function removeHobbyRow(index: number) {
    setHobbies((rows) => rows.filter((_, i) => i !== index));
  }

  function updatePersonality(trait: string, patch: Partial<PersonalityRow>) {
    setPersonality((prev) => ({ ...prev, [trait]: { ...prev[trait], ...patch } }));
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
    setIf('country', country.toUpperCase());
    setMulti('language', languages);
    setMulti('relationship_goals', relationshipGoals);
    setIf('has_children', hasChildren);
    setIf('wants_children', wantsChildren);
    setMulti('interests', selectedInterests);

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

    setIf('nationality', nationality.toUpperCase());
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

    // Hobbies: clave suelta (solo "le gusta") o con límites de intensidad.
    // "any" -> ?hobby=key ; gte/lte/eq -> ?hobby_{key}_min / _max
    const bareHobbies: string[] = [];
    hobbies.forEach((row) => {
      const key = row.key.trim();
      if (!key) return;
      if (row.comparator === 'any') {
        bareHobbies.push(key);
      } else if (row.value.trim() !== '') {
        if (row.comparator === 'gte' || row.comparator === 'eq') params.set(`hobby_${key}_min`, row.value.trim());
        if (row.comparator === 'lte' || row.comparator === 'eq') params.set(`hobby_${key}_max`, row.value.trim());
      }
    });
    setMulti('hobby', bareHobbies);

    // Personalidad: mismo patrón que hobbies, sin la opción "any" (no tiene
    // sentido pedir un rasgo sin ningún límite).
    PERSONALITY_TRAIT_KEYS.forEach((trait) => {
      const row = personality[trait];
      if (!row || row.comparator === 'none' || row.value.trim() === '') return;
      if (row.comparator === 'gte' || row.comparator === 'eq') params.set(`trait_${trait}_min`, row.value.trim());
      if (row.comparator === 'lte' || row.comparator === 'eq') params.set(`trait_${trait}_max`, row.value.trim());
    });

    if (sort) params.set('sort', sort);

    return params;
  }

  function handleSubmit(e: FormEvent) {
    e.preventDefault();
    router.push(`/discover?${buildParams().toString()}`);
  }

  function handleReset() {
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
    setHobbies([]);
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
            <RangeField
              label={dictionary.search.labelAge}
              minValue={minAge}
              maxValue={maxAge}
              onMinChange={setMinAge}
              onMaxChange={setMaxAge}
              placeholderMin={dictionary.search.placeholderAgeMin}
              placeholderMax={dictionary.search.placeholderAgeMax}
            />
            <div className={styles.field}>
              <label className={styles.label}>{dictionary.search.labelCountry}</label>
              <input
                type="text"
                className={styles.input}
                placeholder={dictionary.search.placeholderCountry}
                maxLength={2}
                value={country}
                onChange={(e) => setCountry(e.target.value.toUpperCase())}
              />
            </div>
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
            <RangeField label={dictionary.search.labelHeight} minValue={minHeight} maxValue={maxHeight} onMinChange={setMinHeight} onMaxChange={setMaxHeight} placeholderMin="cm" placeholderMax="cm" />
            <RangeField label={dictionary.search.labelWeight} minValue={minWeight} maxValue={maxWeight} onMinChange={setMinWeight} onMaxChange={setMaxWeight} placeholderMin="kg" placeholderMax="kg" />
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
              <input type="number" min={0} className={styles.input} value={maxChildren} onChange={(e) => setMaxChildren(e.target.value)} />
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
            <div className={styles.field}>
              <label className={styles.label}>{dictionary.search.labelNationality}</label>
              <input type="text" className={styles.input} maxLength={2} placeholder={dictionary.search.placeholderCountry} value={nationality} onChange={(e) => setNationality(e.target.value.toUpperCase())} />
            </div>
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

        {/* --- Intereses --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionInterests}</h2>
          {interestsError && <p className={styles.freeTextWarning}>{dictionary.search.interestsLoadError}</p>}
          {interestsByCategory.map((group) => (
            <div key={group.key}>
              <p className={styles.categoryLabel}>{dictionary.options.interestCategories?.[group.key as keyof typeof dictionary.options.interestCategories] ?? group.key}</p>
              <div className={styles.chipGroup}>
                {group.items.map((item) => {
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
            </div>
          ))}
        </section>

        {/* --- Hobbies (comparador, catálogo sin confirmar) --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionHobbies}</h2>
          <p className={styles.freeTextWarning}>{dictionary.search.hobbiesFreeTextWarning}</p>
          {hobbies.map((row, index) => (
            <div key={index} className={styles.comparatorRow}>
              <input
                type="text"
                className={styles.input}
                placeholder={dictionary.search.hobbyKeyPlaceholder}
                value={row.key}
                onChange={(e) => updateHobbyRow(index, { key: e.target.value })}
              />
              <ComparatorSelect value={row.comparator} onChange={(c) => updateHobbyRow(index, { comparator: c as Comparator })} dictionary={dictionary} includeAny />
              {row.comparator !== 'any' && (
                <input
                  type="number"
                  className={styles.input}
                  placeholder={dictionary.search.intensityPlaceholder}
                  value={row.value}
                  onChange={(e) => updateHobbyRow(index, { value: e.target.value })}
                />
              )}
              {row.comparator === 'any' && <span />}
              <button type="button" className={styles.removeBtn} onClick={() => removeHobbyRow(index)} title={dictionary.search.removeRow}>
                ✕
              </button>
            </div>
          ))}
          <button type="button" className={styles.addBtn} onClick={addHobbyRow}>
            + {dictionary.search.addHobby}
          </button>
        </section>

        {/* --- Personalidad (Big Five, comparador) --- */}
        <section className={styles.section}>
          <h2 className={styles.sectionTitle}>{dictionary.search.sectionPersonality}</h2>
          {PERSONALITY_TRAIT_KEYS.map((trait) => (
            <div key={trait} className={styles.comparatorRow}>
              <span className={styles.label}>{dictionary.options.personalityTraits?.[trait as keyof typeof dictionary.options.personalityTraits] ?? trait}</span>
              <ComparatorSelect
                value={personality[trait]?.comparator ?? 'none'}
                onChange={(c) => updatePersonality(trait, { comparator: c })}
                dictionary={dictionary}
                includeAny={false}
              />
              {personality[trait]?.comparator !== 'none' && (
                <input
                  type="number"
                  className={styles.input}
                  placeholder={dictionary.search.scorePlaceholder}
                  value={personality[trait]?.value ?? ''}
                  onChange={(e) => updatePersonality(trait, { value: e.target.value })}
                />
              )}
              {personality[trait]?.comparator === 'none' && <span />}
              <span />
            </div>
          ))}
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
