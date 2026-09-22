'use client';

import { ChangeEvent, FormEvent, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import {
  apiFetch,
  ApiError,
  ProfilePhoto,
  PublicProfile,
  ProfileLanguage,
  InterestDefinition,
  ProfileInterest,
  PersonalityStatement,
  PersonalityResponse,
  PersonalityTraitScore,
  PartnerPreferences,
} from '@/lib/api';
import {
  Option,
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
  DESIRED_TRAITS_OPTIONS,
  PARTNER_MAY_HAVE_CHILDREN_OPTIONS,
  PARTNER_RELIGION_PREFERENCE_OPTIONS,
  FIRST_MEETING_PREFERENCE_OPTIONS,
  DESIRED_LIVING_PLACE_OPTIONS,
  PARTNER_IMPORTANCE_FIELDS,
  PERSONALITY_TRAIT_LABELS,
  INTEREST_CATEGORY_LABELS,
  LANGUAGE_OPTIONS,
} from '@/lib/profileOptions';
import LocationAutocomplete from '@/components/LocationAutocomplete';
import { PlaceSuggestion } from '@/lib/geocoding';
import styles from './page.module.css';

// =====================================================================
// Tipos y helpers de conversión
// =====================================================================

type FormState = {
  // Básico
  display_name: string;
  birth_date: string;
  gender: string;
  country_code: string;
  region: string;
  relationship_goals: string[];
  has_children: string;
  wants_children: string;
  bio: string;

  // Físico y apariencia
  height: string;
  weight: string;
  body_type: string;
  ethnicity: string;
  appearance_rating: string;
  hair_color: string;
  eye_color: string;
  body_art: string[];

  // Estilo de vida y familia
  smoking_habit: string;
  drinking_habit: string;
  relocation_willingness: string[];
  marital_status: string;
  children_count: string;
  youngest_child_age: string;
  oldest_child_age: string;
  occupation: string;
  employment_status: string;
  income_level: string;
  living_situation: string;

  // Fondo, cultura y valores
  nationality: string;
  education_level: string;
  english_ability: string;
  religion: string;
  religious_values: string;
  star_sign: string;

  // Über mich / estilo de vida
  future_vision: string[];
  sports: string[];
  likes_pets: string;
  pets_owned: string[];
  favorite_season: string;
  ideal_vacation_style: string[];
  vacation_activities: string[];
  profile_quote: string;
  dream_wish: string;
};

type MultiSelectField = Extract<
  keyof FormState,
  'body_art' | 'relocation_willingness' | 'future_vision' | 'sports' | 'pets_owned' | 'ideal_vacation_style' | 'vacation_activities'
>;

const emptyForm: FormState = {
  display_name: '', birth_date: '', gender: '', country_code: '', region: '',
  relationship_goals: [], has_children: '', wants_children: '', bio: '',

  height: '', weight: '', body_type: '', ethnicity: '', appearance_rating: '',
  hair_color: '', eye_color: '', body_art: [],

  smoking_habit: '', drinking_habit: '', relocation_willingness: [], marital_status: '',
  children_count: '', youngest_child_age: '', oldest_child_age: '',
  occupation: '', employment_status: '', income_level: '', living_situation: '',

  nationality: '', education_level: '', english_ability: '',
  religion: '', religious_values: '', star_sign: '',

  future_vision: [], sports: [], likes_pets: '', pets_owned: [],
  favorite_season: '', ideal_vacation_style: [], vacation_activities: [],
  profile_quote: '', dream_wish: '',
};

function nullableString(value: string) {
  return value.trim() || null;
}

function nullableNumber(value: string): number | null {
  const trimmed = value.trim();
  if (trimmed === '') return null;
  const n = Number(trimmed);
  return Number.isNaN(n) ? null : n;
}

function formFromProfile(profile: PublicProfile): FormState {
  return {
    display_name: profile.display_name,
    birth_date: '',
    gender: profile.gender,
    country_code: profile.country_code,
    region: profile.region ?? '',
    relationship_goals: profile.relationship_goals ?? [],
    has_children: profile.has_children ?? '',
    wants_children: profile.wants_children ?? '',
    bio: profile.bio ?? '',

    height: profile.height?.toString() ?? '',
    weight: profile.weight?.toString() ?? '',
    body_type: profile.body_type ?? '',
    ethnicity: profile.ethnicity ?? '',
    appearance_rating: profile.appearance_rating ?? '',
    hair_color: profile.hair_color ?? '',
    eye_color: profile.eye_color ?? '',
    body_art: profile.body_art ?? [],

    smoking_habit: profile.smoking_habit ?? '',
    drinking_habit: profile.drinking_habit ?? '',
    relocation_willingness: profile.relocation_willingness ?? [],
    marital_status: profile.marital_status ?? '',
    children_count: profile.children_count?.toString() ?? '',
    youngest_child_age: profile.youngest_child_age?.toString() ?? '',
    oldest_child_age: profile.oldest_child_age?.toString() ?? '',
    occupation: profile.occupation ?? '',
    employment_status: profile.employment_status ?? '',
    income_level: profile.income_level ?? '',
    living_situation: profile.living_situation ?? '',

    nationality: profile.nationality ?? '',
    education_level: profile.education_level ?? '',
    english_ability: profile.english_ability ?? '',
    religion: profile.religion ?? '',
    religious_values: profile.religious_values ?? '',
    star_sign: profile.star_sign ?? '',

    future_vision: profile.future_vision ?? [],
    sports: profile.sports ?? [],
    likes_pets: profile.likes_pets ?? '',
    pets_owned: profile.pets_owned ?? [],
    favorite_season: profile.favorite_season ?? '',
    ideal_vacation_style: profile.ideal_vacation_style ?? [],
    vacation_activities: profile.vacation_activities ?? [],
    profile_quote: profile.profile_quote ?? '',
    dream_wish: profile.dream_wish ?? '',
  };
}

type PartnerFormState = {
  age_min: string;
  age_max: string;
  height_min: string;
  height_max: string;
  desired_traits: string[];
  partner_may_have_children: string;
  partner_religion_preference: string;
  about_partner_text: string;
  first_meeting_preference: string;
  desired_living_place: string[];
  importance_shared_thoughts: string;
  importance_shared_hobbies: string;
  importance_intimacy: string;
  importance_romantic_love: string;
  importance_financial_security: string;
  importance_fun: string;
  importance_shared_friends: string;
  importance_shared_humor: string;
  importance_personal_space: string;
  importance_independence: string;
};

const emptyPartnerForm: PartnerFormState = {
  age_min: '', age_max: '', height_min: '', height_max: '',
  desired_traits: [], partner_may_have_children: '', partner_religion_preference: '',
  about_partner_text: '', first_meeting_preference: '', desired_living_place: [],
  importance_shared_thoughts: '', importance_shared_hobbies: '', importance_intimacy: '',
  importance_romantic_love: '', importance_financial_security: '', importance_fun: '',
  importance_shared_friends: '', importance_shared_humor: '', importance_personal_space: '',
  importance_independence: '',
};

function partnerFormFromPreferences(pp: PartnerPreferences): PartnerFormState {
  return {
    age_min: pp.age_min?.toString() ?? '',
    age_max: pp.age_max?.toString() ?? '',
    height_min: pp.height_min?.toString() ?? '',
    height_max: pp.height_max?.toString() ?? '',
    desired_traits: pp.desired_traits ?? [],
    partner_may_have_children: pp.partner_may_have_children ?? '',
    partner_religion_preference: pp.partner_religion_preference ?? '',
    about_partner_text: pp.about_partner_text ?? '',
    first_meeting_preference: pp.first_meeting_preference ?? '',
    desired_living_place: pp.desired_living_place ?? [],
    importance_shared_thoughts: pp.importance_shared_thoughts?.toString() ?? '',
    importance_shared_hobbies: pp.importance_shared_hobbies?.toString() ?? '',
    importance_intimacy: pp.importance_intimacy?.toString() ?? '',
    importance_romantic_love: pp.importance_romantic_love?.toString() ?? '',
    importance_financial_security: pp.importance_financial_security?.toString() ?? '',
    importance_fun: pp.importance_fun?.toString() ?? '',
    importance_shared_friends: pp.importance_shared_friends?.toString() ?? '',
    importance_shared_humor: pp.importance_shared_humor?.toString() ?? '',
    importance_personal_space: pp.importance_personal_space?.toString() ?? '',
    importance_independence: pp.importance_independence?.toString() ?? '',
  };
}

const TABS = [
  { id: 'basic', label: 'Básico' },
  { id: 'physical', label: 'Físico' },
  { id: 'lifestyle', label: 'Estilo de vida' },
  { id: 'background', label: 'Fondo y cultura' },
  { id: 'uber', label: 'Über mich' },
  { id: 'languages', label: 'Idiomas' },
  { id: 'interests', label: 'Intereses' },
  { id: 'personality', label: 'Personalidad' },
  { id: 'partner', label: 'Pareja' },
] as const;

type TabId = (typeof TABS)[number]['id'];

// =====================================================================
// Componente
// =====================================================================

export default function EditProfilePage() {
  const router = useRouter();
  const [tab, setTab] = useState<TabId>('basic');

  const [form, setForm] = useState<FormState>(emptyForm);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [creating, setCreating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);

  // Idiomas
  const [myLanguages, setMyLanguages] = useState<ProfileLanguage[]>([]);
  const [languageQuery, setLanguageQuery] = useState('');

  // Intereses
  const [interestCatalog, setInterestCatalog] = useState<InterestDefinition[]>([]);
  const [myInterests, setMyInterests] = useState<ProfileInterest[]>([]);
  const [interestQuery, setInterestQuery] = useState('');

  // Personalidad
  const [personalityCatalog, setPersonalityCatalog] = useState<PersonalityStatement[]>([]);
  const [personalityAnswers, setPersonalityAnswers] = useState<Record<string, number>>({});
  const [traitScores, setTraitScores] = useState<PersonalityTraitScore[]>([]);

  // Preferencias de pareja
  const [partnerForm, setPartnerForm] = useState<PartnerFormState>(emptyPartnerForm);
  const [partnerSaving, setPartnerSaving] = useState(false);
  const [partnerSaved, setPartnerSaved] = useState(false);

  useEffect(() => {
    // Los catálogos son públicos y no dependen de tener perfil creado.
    apiFetch<InterestDefinition[]>('/catalog/interests').then(setInterestCatalog).catch(() => setInterestCatalog([]));
    apiFetch<PersonalityStatement[]>('/catalog/personality-statements')
      .then(setPersonalityCatalog)
      .catch(() => setPersonalityCatalog([]));

    apiFetch<PublicProfile>('/profiles/me')
      .then(async (profile) => {
        setForm(formFromProfile(profile));

        try {
          setPhotos(await apiFetch<ProfilePhoto[]>('/profiles/me/photos'));
        } catch {
          setPhotos([]);
        }

        try {
          setMyLanguages(await apiFetch<ProfileLanguage[]>('/profiles/me/languages'));
        } catch {
          setMyLanguages([]);
        }

        try {
          setMyInterests(await apiFetch<ProfileInterest[]>('/profiles/me/interests'));
        } catch {
          setMyInterests([]);
        }

        try {
          const personality = await apiFetch<PersonalityResponse>('/profiles/me/personality');
          const answers: Record<string, number> = {};
          for (const a of personality.answers) answers[a.statement_key] = a.score;
          setPersonalityAnswers(answers);
          setTraitScores(personality.trait_scores);
        } catch {
          setPersonalityAnswers({});
          setTraitScores([]);
        }

        try {
          const prefs = await apiFetch<PartnerPreferences>('/profiles/me/partner-preferences');
          setPartnerForm(partnerFormFromPreferences(prefs));
        } catch {
          setPartnerForm(emptyPartnerForm);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setCreating(true);
        } else if (err instanceof ApiError && err.status === 401) {
          router.push('/login');
        } else {
          setError('No se pudo cargar el perfil.');
        }
      })
      .finally(() => setLoading(false));
  }, [router]);

  function updateField<K extends keyof FormState>(field: K, value: FormState[K]) {
    setForm((current) => ({ ...current, [field]: value }));
  }

  function toggleMulti(field: MultiSelectField, value: string) {
    setForm((current) => {
      const arr = current[field];
      const next = arr.includes(value) ? arr.filter((v) => v !== value) : [...arr, value];
      return { ...current, [field]: next };
    });
  }

  function toggleRelationshipGoal(value: string) {
    setForm((current) => {
      const next = current.relationship_goals.includes(value)
        ? current.relationship_goals.filter((v) => v !== value)
        : [...current.relationship_goals, value];
      return { ...current, relationship_goals: next };
    });
  }

  function handleLocationSelect(place: PlaceSuggestion) {
    updateField('region', [place.city, place.state].filter(Boolean).join(', '));
    updateField('country_code', place.countryCode);
  }

  // --- Guardar el perfil ---

  async function handleSubmit(event: FormEvent) {
    event.preventDefault();
    setSaving(true);
    setError(null);
    setSaved(false);

    const body = {
      display_name: form.display_name,
      gender: form.gender,
      country_code: form.country_code,
      region: nullableString(form.region),
      relationship_goals: form.relationship_goals,
      has_children: nullableString(form.has_children),
      wants_children: nullableString(form.wants_children),
      bio: nullableString(form.bio),

      height: nullableNumber(form.height),
      weight: nullableNumber(form.weight),
      body_type: nullableString(form.body_type),
      ethnicity: nullableString(form.ethnicity),
      appearance_rating: nullableString(form.appearance_rating),
      hair_color: nullableString(form.hair_color),
      eye_color: nullableString(form.eye_color),
      body_art: form.body_art,

      smoking_habit: nullableString(form.smoking_habit),
      drinking_habit: nullableString(form.drinking_habit),
      relocation_willingness: form.relocation_willingness,
      marital_status: nullableString(form.marital_status),
      children_count: nullableNumber(form.children_count),
      youngest_child_age: nullableNumber(form.youngest_child_age),
      oldest_child_age: nullableNumber(form.oldest_child_age),
      occupation: nullableString(form.occupation),
      employment_status: nullableString(form.employment_status),
      income_level: nullableString(form.income_level),
      living_situation: nullableString(form.living_situation),

      nationality: nullableString(form.nationality),
      education_level: nullableString(form.education_level),
      english_ability: nullableString(form.english_ability),
      religion: nullableString(form.religion),
      religious_values: nullableString(form.religious_values),
      star_sign: nullableString(form.star_sign),

      future_vision: form.future_vision,
      sports: form.sports,
      likes_pets: nullableString(form.likes_pets),
      pets_owned: form.pets_owned,
      favorite_season: nullableString(form.favorite_season),
      ideal_vacation_style: form.ideal_vacation_style,
      vacation_activities: form.vacation_activities,
      profile_quote: nullableString(form.profile_quote),
      dream_wish: nullableString(form.dream_wish),

      ...(creating ? { birth_date: form.birth_date } : {}),
    };

    try {
      const profile = await apiFetch<PublicProfile>('/profiles/me', {
        method: creating ? 'POST' : 'PATCH',
        body: JSON.stringify(body),
      });
      setForm(formFromProfile(profile));
      setCreating(false);
      setSaved(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo guardar el perfil.');
    } finally {
      setSaving(false);
    }
  }

  // --- Fotos ---

  async function handleUpload(event: ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file) return;

    setUploading(true);
    setError(null);
    const data = new FormData();
    data.append('photo', file);
    try {
      const photo = await apiFetch<ProfilePhoto>('/profiles/me/photos', {
        method: 'POST',
        body: data,
      });
      setPhotos((current) => [...current, photo]);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo subir la foto.');
    } finally {
      setUploading(false);
    }
  }

  async function handleDeletePhoto(photoID: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/photos/${photoID}`, { method: 'DELETE' });
      setPhotos((current) => current.filter((photo) => photo.id !== photoID));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo eliminar la foto.');
    }
  }

  // --- Idiomas: cada cambio se guarda al momento ---

  async function handleAddLanguage(code: string) {
    setError(null);
    try {
      const saved = await apiFetch<ProfileLanguage>(`/profiles/me/languages/${code}`, {
        method: 'PUT',
        body: JSON.stringify({ level: null }),
      });
      setMyLanguages((current) => [...current.filter((l) => l.language_code !== code), saved]);
      setLanguageQuery('');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo añadir el idioma.');
    }
  }

  async function handleLanguageLevelChange(code: string, level: number | null) {
    setError(null);
    try {
      const saved = await apiFetch<ProfileLanguage>(`/profiles/me/languages/${code}`, {
        method: 'PUT',
        body: JSON.stringify({ level }),
      });
      setMyLanguages((current) => current.map((l) => (l.language_code === code ? saved : l)));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo guardar el nivel.');
    }
  }

  async function handleRemoveLanguage(code: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/languages/${code}`, { method: 'DELETE' });
      setMyLanguages((current) => current.filter((l) => l.language_code !== code));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo quitar el idioma.');
    }
  }

  const languageSuggestions = useMemo(() => {
    const q = languageQuery.trim().toLowerCase();
    if (q === '') return [];
    const selected = new Set(myLanguages.map((l) => l.language_code));
    return LANGUAGE_OPTIONS.filter((o) => !selected.has(o.value) && o.label.toLowerCase().includes(q)).slice(0, 8);
  }, [languageQuery, myLanguages]);

  // --- Intereses: cada cambio se guarda al momento ---

  async function handleSetInterest(key: string, level: number | null) {
    setError(null);
    try {
      const saved = await apiFetch<ProfileInterest>(`/profiles/me/interests/${key}`, {
        method: 'PUT',
        body: JSON.stringify({ level }),
      });
      setMyInterests((current) => [...current.filter((i) => i.interest_key !== key), saved]);
      setInterestQuery('');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo guardar el interés.');
    }
  }

  async function handleRemoveInterest(key: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/interests/${key}`, { method: 'DELETE' });
      setMyInterests((current) => current.filter((i) => i.interest_key !== key));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo quitar el interés.');
    }
  }

  const leveledInterests = useMemo(() => interestCatalog.filter((d) => d.has_level), [interestCatalog]);
  const concreteInterests = useMemo(() => interestCatalog.filter((d) => !d.has_level), [interestCatalog]);
  const myInterestByKey = useMemo(() => {
    const map: Record<string, ProfileInterest> = {};
    for (const i of myInterests) map[i.interest_key] = i;
    return map;
  }, [myInterests]);

  const selectedConcreteInterests = useMemo(
    () => concreteInterests.filter((d) => myInterestByKey[d.key]),
    [concreteInterests, myInterestByKey]
  );

  const concreteSuggestions = useMemo(() => {
    const q = interestQuery.trim().toLowerCase();
    if (q === '') return [];
    return concreteInterests.filter((d) => !myInterestByKey[d.key] && d.label.toLowerCase().includes(q)).slice(0, 8);
  }, [interestQuery, concreteInterests, myInterestByKey]);

  // --- Personalidad: cada respuesta se guarda al momento ---

  async function handlePersonalityChange(key: string, score: number) {
    setError(null);
    try {
      await apiFetch(`/profiles/me/personality/${key}`, {
        method: 'PUT',
        body: JSON.stringify({ score }),
      });
      setPersonalityAnswers((current) => ({ ...current, [key]: score }));
      const refreshed = await apiFetch<PersonalityResponse>('/profiles/me/personality');
      setTraitScores(refreshed.trait_scores);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudo guardar la respuesta.');
    }
  }

  // --- Preferencias de pareja: un único guardado para toda la pestaña ---

  function updatePartnerField<K extends keyof PartnerFormState>(field: K, value: PartnerFormState[K]) {
    setPartnerForm((current) => ({ ...current, [field]: value }));
  }

  function togglePartnerMulti(field: 'desired_traits' | 'desired_living_place', value: string) {
    setPartnerForm((current) => {
      const arr = current[field];
      const next = arr.includes(value) ? arr.filter((v) => v !== value) : [...arr, value];
      return { ...current, [field]: next };
    });
  }

  async function handleSavePartnerPreferences(event: FormEvent) {
    event.preventDefault();
    setPartnerSaving(true);
    setPartnerSaved(false);
    setError(null);

    const body = {
      age_min: nullableNumber(partnerForm.age_min),
      age_max: nullableNumber(partnerForm.age_max),
      height_min: nullableNumber(partnerForm.height_min),
      height_max: nullableNumber(partnerForm.height_max),
      desired_traits: partnerForm.desired_traits,
      partner_may_have_children: nullableString(partnerForm.partner_may_have_children),
      partner_religion_preference: nullableString(partnerForm.partner_religion_preference),
      about_partner_text: nullableString(partnerForm.about_partner_text),
      first_meeting_preference: nullableString(partnerForm.first_meeting_preference),
      desired_living_place: partnerForm.desired_living_place,
      importance_shared_thoughts: nullableNumber(partnerForm.importance_shared_thoughts),
      importance_shared_hobbies: nullableNumber(partnerForm.importance_shared_hobbies),
      importance_intimacy: nullableNumber(partnerForm.importance_intimacy),
      importance_romantic_love: nullableNumber(partnerForm.importance_romantic_love),
      importance_financial_security: nullableNumber(partnerForm.importance_financial_security),
      importance_fun: nullableNumber(partnerForm.importance_fun),
      importance_shared_friends: nullableNumber(partnerForm.importance_shared_friends),
      importance_shared_humor: nullableNumber(partnerForm.importance_shared_humor),
      importance_personal_space: nullableNumber(partnerForm.importance_personal_space),
      importance_independence: nullableNumber(partnerForm.importance_independence),
    };

    try {
      const saved = await apiFetch<PartnerPreferences>('/profiles/me/partner-preferences', {
        method: 'PATCH',
        body: JSON.stringify(body),
      });
      setPartnerForm(partnerFormFromPreferences(saved));
      setPartnerSaved(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'No se pudieron guardar las preferencias de pareja.');
    } finally {
      setPartnerSaving(false);
    }
  }

  // --- Helpers de render ---

  function renderSelect(field: keyof FormState, label: string, options: Option[]) {
    return (
      <label>
        {label}
        <select value={form[field] as string} onChange={(e) => updateField(field, e.target.value as never)}>
          <option value="">Prefiero no decirlo</option>
          {options.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
      </label>
    );
  }

  function renderCheckboxGroup(field: MultiSelectField, label: string, options: Option[]) {
    return (
      <fieldset className={styles.checkboxFieldset}>
        <legend>{label}</legend>
        <div className={styles.checkboxGrid}>
          {options.map((opt) => (
            <label key={opt.value} className={styles.checkboxLabel}>
              <input
                type="checkbox"
                checked={form[field].includes(opt.value)}
                onChange={() => toggleMulti(field, opt.value)}
              />
              {opt.label}
            </label>
          ))}
        </div>
      </fieldset>
    );
  }

  if (loading) {
    return <main className={styles.main}><p>Cargando…</p></main>;
  }

  const disabledUntilCreated = creating;

  return (
    <main className={styles.main}>
      <Link href="/profile">&larr; Volver a mi perfil</Link>
      <h1>{creating ? 'Crear perfil' : 'Modificar perfil'}</h1>

      <nav className={styles.tabs}>
        {TABS.map((t) => (
          <button
            key={t.id}
            type="button"
            className={tab === t.id ? styles.tabActive : styles.tab}
            onClick={() => setTab(t.id)}
          >
            {t.label}
          </button>
        ))}
      </nav>

      {error && <p className={styles.error}>{error}</p>}

      {/* --- Básico / Físico / Estilo de vida / Fondo / Über mich: un único form --- */}
      {tab !== 'languages' && tab !== 'interests' && tab !== 'personality' && tab !== 'partner' && (
        <form onSubmit={handleSubmit} className={styles.form}>
          {tab === 'basic' && (
            <>
              <label>Nombre visible<input value={form.display_name} onChange={(e) => updateField('display_name', e.target.value)} required /></label>
              {creating && <label>Fecha de nacimiento<input type="date" value={form.birth_date} onChange={(e) => updateField('birth_date', e.target.value)} required /></label>}
              <label>Género
                <select value={form.gender} onChange={(e) => updateField('gender', e.target.value)} required>
                  <option value="">Selecciona una opción</option>
                  <option value="female">Mujer</option><option value="male">Hombre</option>
                  <option value="non_binary">No binario</option><option value="other">Otro</option>
                </select>
              </label>

              <label>Ubicación
                <LocationAutocomplete
                  initialValue={form.region && form.country_code ? `${form.region}, ${form.country_code}` : ''}
                  onSelect={handleLocationSelect}
                />
              </label>
              {form.country_code && (
                <p className={styles.hint}>Guardado: {form.region ? `${form.region}, ` : ''}{form.country_code}</p>
              )}

              <fieldset className={styles.checkboxFieldset}>
                <legend>¿Qué buscas? (puedes elegir varias)</legend>
                <div className={styles.checkboxGrid}>
                  {RELATIONSHIP_GOAL_OPTIONS.map((opt) => (
                    <label key={opt.value} className={styles.checkboxLabel}>
                      <input
                        type="checkbox"
                        checked={form.relationship_goals.includes(opt.value)}
                        onChange={() => toggleRelationshipGoal(opt.value)}
                      />
                      {opt.label}
                    </label>
                  ))}
                </div>
              </fieldset>

              {renderSelect('has_children', '¿Tienes hijos?', HAS_CHILDREN_OPTIONS)}
              {renderSelect('wants_children', '¿Quieres tener hijos?', WANTS_CHILDREN_OPTIONS)}
              <label>Biografía<textarea value={form.bio} onChange={(e) => updateField('bio', e.target.value)} maxLength={1000} rows={5} /></label>
            </>
          )}

          {tab === 'physical' && (
            <>
              <label>Altura (cm)<input type="number" min={50} max={300} value={form.height} onChange={(e) => updateField('height', e.target.value)} /></label>
              <label>Peso (kg)<input type="number" min={20} max={400} value={form.weight} onChange={(e) => updateField('weight', e.target.value)} /></label>
              {renderSelect('body_type', 'Complexión', BODY_TYPE_OPTIONS)}
              {renderSelect('ethnicity', 'Etnia', ETHNICITY_OPTIONS)}
              {renderSelect('appearance_rating', 'Cómo describirías tu aspecto', APPEARANCE_RATING_OPTIONS)}
              {renderSelect('hair_color', 'Color de pelo', HAIR_COLOR_OPTIONS)}
              {renderSelect('eye_color', 'Color de ojos', EYE_COLOR_OPTIONS)}
              {renderCheckboxGroup('body_art', 'Piercings / tatuajes', BODY_ART_OPTIONS)}
            </>
          )}

          {tab === 'lifestyle' && (
            <>
              {renderSelect('smoking_habit', '¿Fumas?', SMOKING_HABIT_OPTIONS)}
              {renderSelect('drinking_habit', '¿Bebes alcohol?', DRINKING_HABIT_OPTIONS)}
              {renderCheckboxGroup('relocation_willingness', 'Disposición a mudarte', RELOCATION_WILLINGNESS_OPTIONS)}
              {renderSelect('marital_status', 'Estado civil', MARITAL_STATUS_OPTIONS)}
              <label>Número de hijos<input type="number" min={0} max={25} value={form.children_count} onChange={(e) => updateField('children_count', e.target.value)} /></label>
              <label>Edad del hijo/a más pequeño/a<input type="number" min={0} max={100} value={form.youngest_child_age} onChange={(e) => updateField('youngest_child_age', e.target.value)} /></label>
              <label>Edad del hijo/a más mayor<input type="number" min={0} max={100} value={form.oldest_child_age} onChange={(e) => updateField('oldest_child_age', e.target.value)} /></label>
              {renderSelect('occupation', 'Ocupación', OCCUPATION_OPTIONS)}
              {renderSelect('employment_status', 'Situación laboral', EMPLOYMENT_STATUS_OPTIONS)}
              {renderSelect('income_level', 'Nivel de ingresos', INCOME_LEVEL_OPTIONS)}
              {renderSelect('living_situation', 'Con quién vives', LIVING_SITUATION_OPTIONS)}
            </>
          )}

          {tab === 'background' && (
            <>
              <label>Nacionalidad (código de dos letras)<input value={form.nationality} onChange={(e) => updateField('nationality', e.target.value.toUpperCase())} maxLength={2} /></label>
              {renderSelect('education_level', 'Nivel educativo', EDUCATION_LEVEL_OPTIONS)}
              {renderSelect('english_ability', 'Nivel de inglés', ENGLISH_ABILITY_OPTIONS)}
              {renderSelect('religion', 'Religión', RELIGION_OPTIONS)}
              {renderSelect('religious_values', 'Nivel de religiosidad', RELIGIOUS_VALUES_OPTIONS)}
              {renderSelect('star_sign', 'Signo del zodiaco', STAR_SIGN_OPTIONS)}
            </>
          )}

          {tab === 'uber' && (
            <>
              <label>Frase de presentación <span className={styles.hint}>máx. 500 caracteres</span><textarea value={form.profile_quote} onChange={(e) => updateField('profile_quote', e.target.value)} maxLength={500} rows={3} /></label>
              {renderCheckboxGroup('future_vision', '¿Cómo te imaginas el futuro?', FUTURE_VISION_OPTIONS)}
              {renderCheckboxGroup('sports', '¿Qué deporte practicas?', SPORTS_OPTIONS)}
              {renderSelect('likes_pets', '¿Te gustan las mascotas?', LIKES_PETS_OPTIONS)}
              {renderCheckboxGroup('pets_owned', '¿Qué mascotas tienes?', PETS_OWNED_OPTIONS)}
              {renderSelect('favorite_season', 'Estación favorita', FAVORITE_SEASON_OPTIONS)}
              {renderCheckboxGroup('ideal_vacation_style', 'Vacaciones ideales', IDEAL_VACATION_STYLE_OPTIONS)}
              {renderCheckboxGroup('vacation_activities', 'Actividades de vacaciones favoritas', VACATION_ACTIVITIES_OPTIONS)}
              <label>Un sueño o algo que siempre has querido hacer <span className={styles.hint}>máx. 500 caracteres</span><textarea value={form.dream_wish} onChange={(e) => updateField('dream_wish', e.target.value)} maxLength={500} rows={3} /></label>
            </>
          )}

          {saved && <p className={styles.success}>Perfil guardado.</p>}
          <button type="submit" disabled={saving}>{saving ? 'Guardando…' : 'Guardar perfil'}</button>
        </form>
      )}

      {/* --- Idiomas --- */}
      {tab === 'languages' && (
        disabledUntilCreated ? (
          <p className={styles.hint}>Guarda el perfil antes de indicar tus idiomas.</p>
        ) : (
          <div className={styles.hobbyList}>
            {myLanguages.length > 0 && (
              <fieldset className={styles.checkboxFieldset}>
                <legend>Tus idiomas</legend>
                {myLanguages.map((l) => {
                  const opt = LANGUAGE_OPTIONS.find((o) => o.value === l.language_code);
                  return (
                    <div key={l.language_code} className={styles.hobbyRow}>
                      <span className={styles.hobbyLabel}>{opt?.label ?? l.language_code}</span>
                      <select
                        value={l.level ?? ''}
                        onChange={(e) => handleLanguageLevelChange(l.language_code, e.target.value ? Number(e.target.value) : null)}
                      >
                        <option value="">Sin nivel</option>
                        {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                      </select>
                      <button type="button" onClick={() => handleRemoveLanguage(l.language_code)}>Quitar</button>
                    </div>
                  );
                })}
              </fieldset>
            )}

            <label>Añadir idioma
              <input
                value={languageQuery}
                onChange={(e) => setLanguageQuery(e.target.value)}
                placeholder="Buscar idioma..."
              />
            </label>
            {languageSuggestions.length > 0 && (
              <ul className={styles.suggestionList}>
                {languageSuggestions.map((opt) => (
                  <li key={opt.value}>
                    <button type="button" onClick={() => handleAddLanguage(opt.value)}>{opt.label}</button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )
      )}

      {/* --- Intereses --- */}
      {tab === 'interests' && (
        disabledUntilCreated ? (
          <p className={styles.hint}>Guarda el perfil antes de indicar tus intereses.</p>
        ) : (
          <div className={styles.hobbyList}>
            <h2>Intereses principales</h2>
            <p className={styles.hint}>Indica cuánto te gusta cada uno (1 a 5). Los que dejes sin marcar no aparecerán en tu perfil.</p>
            {Object.entries(
              leveledInterests.reduce<Record<string, InterestDefinition[]>>((acc, d) => {
                (acc[d.category] ??= []).push(d);
                return acc;
              }, {})
            ).map(([category, defs]) => (
              <fieldset key={category} className={styles.checkboxFieldset}>
                <legend>{INTEREST_CATEGORY_LABELS[category] ?? category}</legend>
                {defs.map((d) => (
                  <div key={d.key} className={styles.hobbyRow}>
                    <span className={styles.hobbyLabel}>{d.label}</span>
                    <select
                      value={myInterestByKey[d.key]?.level ?? ''}
                      onChange={(e) => (e.target.value ? handleSetInterest(d.key, Number(e.target.value)) : handleRemoveInterest(d.key))}
                    >
                      <option value="">Sin marcar</option>
                      {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                    </select>
                  </div>
                ))}
              </fieldset>
            ))}

            <h2>Otros intereses</h2>
            <p className={styles.hint}>Cosas más concretas, sin nivel — solo se muestran las que marques.</p>

            {selectedConcreteInterests.length > 0 && (
              <div className={styles.chipRow}>
                {selectedConcreteInterests.map((d) => (
                  <span key={d.key} className={styles.chip}>
                    {d.label}
                    <button type="button" onClick={() => handleRemoveInterest(d.key)} aria-label={`Quitar ${d.label}`}>×</button>
                  </span>
                ))}
              </div>
            )}

            <label>Buscar interés
              <input
                value={interestQuery}
                onChange={(e) => setInterestQuery(e.target.value)}
                placeholder="Ej: guitarra, anime, sushi..."
              />
            </label>
            {concreteSuggestions.length > 0 && (
              <ul className={styles.suggestionList}>
                {concreteSuggestions.map((d) => (
                  <li key={d.key}>
                    <button type="button" onClick={() => handleSetInterest(d.key, null)}>{d.label}</button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )
      )}

      {/* --- Personalidad --- */}
      {tab === 'personality' && (
        disabledUntilCreated ? (
          <p className={styles.hint}>Guarda el perfil antes de contestar el test de personalidad.</p>
        ) : (
          <div className={styles.hobbyList}>
            {Object.entries(
              personalityCatalog.reduce<Record<string, PersonalityStatement[]>>((acc, s) => {
                (acc[s.trait_key] ??= []).push(s);
                return acc;
              }, {})
            ).map(([trait, statements]) => {
              const score = traitScores.find((t) => t.trait_key === trait);
              return (
                <fieldset key={trait} className={styles.checkboxFieldset}>
                  <legend>
                    {PERSONALITY_TRAIT_LABELS[trait] ?? trait}
                    {score && <span className={styles.hint}> — media: {score.average_score.toFixed(2)} ({score.answered_count} respuestas)</span>}
                  </legend>
                  {statements.map((s) => (
                    <div key={s.key} className={styles.hobbyRow}>
                      <span className={styles.hobbyLabel}>{s.label}</span>
                      <select
                        value={personalityAnswers[s.key] ?? ''}
                        onChange={(e) => handlePersonalityChange(s.key, Number(e.target.value))}
                      >
                        <option value="">Sin contestar</option>
                        {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                      </select>
                    </div>
                  ))}
                </fieldset>
              );
            })}
          </div>
        )
      )}

      {/* --- Preferencias de pareja --- */}
      {tab === 'partner' && (
        disabledUntilCreated ? (
          <p className={styles.hint}>Guarda el perfil antes de indicar tus preferencias de pareja.</p>
        ) : (
          <form onSubmit={handleSavePartnerPreferences} className={styles.form}>
            <label>Edad mínima<input type="number" min={18} max={120} value={partnerForm.age_min} onChange={(e) => updatePartnerField('age_min', e.target.value)} /></label>
            <label>Edad máxima<input type="number" min={18} max={120} value={partnerForm.age_max} onChange={(e) => updatePartnerField('age_max', e.target.value)} /></label>
            <label>Altura mínima (cm)<input type="number" min={50} max={300} value={partnerForm.height_min} onChange={(e) => updatePartnerField('height_min', e.target.value)} /></label>
            <label>Altura máxima (cm)<input type="number" min={50} max={300} value={partnerForm.height_max} onChange={(e) => updatePartnerField('height_max', e.target.value)} /></label>

            <fieldset className={styles.checkboxFieldset}>
              <legend>Rasgos que buscas</legend>
              <div className={styles.checkboxGrid}>
                {DESIRED_TRAITS_OPTIONS.map((opt) => (
                  <label key={opt.value} className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={partnerForm.desired_traits.includes(opt.value)}
                      onChange={() => togglePartnerMulti('desired_traits', opt.value)}
                    />
                    {opt.label}
                  </label>
                ))}
              </div>
            </fieldset>

            <label>¿Tu pareja puede tener hijos?
              <select value={partnerForm.partner_may_have_children} onChange={(e) => updatePartnerField('partner_may_have_children', e.target.value)}>
                <option value="">Prefiero no decirlo</option>
                {PARTNER_MAY_HAVE_CHILDREN_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </label>

            <label>Religión de tu pareja
              <select value={partnerForm.partner_religion_preference} onChange={(e) => updatePartnerField('partner_religion_preference', e.target.value)}>
                <option value="">Prefiero no decirlo</option>
                {PARTNER_RELIGION_PREFERENCE_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </label>

            <label>¿Qué buscas en una pareja? <span className={styles.hint}>máx. 1000 caracteres</span>
              <textarea value={partnerForm.about_partner_text} onChange={(e) => updatePartnerField('about_partner_text', e.target.value)} maxLength={1000} rows={4} />
            </label>

            <label>Dónde os gustaría conoceros
              <select value={partnerForm.first_meeting_preference} onChange={(e) => updatePartnerField('first_meeting_preference', e.target.value)}>
                <option value="">Prefiero no decirlo</option>
                {FIRST_MEETING_PREFERENCE_OPTIONS.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
              </select>
            </label>

            <fieldset className={styles.checkboxFieldset}>
              <legend>Dónde os gustaría vivir</legend>
              <div className={styles.checkboxGrid}>
                {DESIRED_LIVING_PLACE_OPTIONS.map((opt) => (
                  <label key={opt.value} className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={partnerForm.desired_living_place.includes(opt.value)}
                      onChange={() => togglePartnerMulti('desired_living_place', opt.value)}
                    />
                    {opt.label}
                  </label>
                ))}
              </div>
            </fieldset>

            <fieldset className={styles.checkboxFieldset}>
              <legend>¿Qué es importante en una relación? <span className={styles.hint}>1 = poco, 5 = mucho</span></legend>
              {PARTNER_IMPORTANCE_FIELDS.map((f) => (
                <div key={f.key} className={styles.hobbyRow}>
                  <span className={styles.hobbyLabel}>{f.label}</span>
                  <select
                    value={partnerForm[f.key as keyof PartnerFormState]}
                    onChange={(e) => updatePartnerField(f.key as keyof PartnerFormState, e.target.value)}
                  >
                    <option value="">Sin contestar</option>
                    {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                  </select>
                </div>
              ))}
            </fieldset>

            {partnerSaved && <p className={styles.success}>Preferencias guardadas.</p>}
            <button type="submit" disabled={partnerSaving}>{partnerSaving ? 'Guardando…' : 'Guardar preferencias'}</button>
          </form>
        )
      )}

      {/* --- Fotos: visibles en cualquier pestaña, al final --- */}
      <section className={styles.photos}>
        <h2>Fotos</h2>
        {photos.length > 0 && (
          <div className={styles.photoGrid}>
            {photos.map((photo) => (
              <div key={photo.id} className={styles.photo}>
                <img src={photo.url} alt="" />
                <button type="button" onClick={() => handleDeletePhoto(photo.id)}>Eliminar</button>
              </div>
            ))}
          </div>
        )}
        {!creating && (
          <label className={styles.upload}>
            Añadir foto
            <input type="file" accept="image/jpeg,image/png,image/webp" onChange={handleUpload} disabled={uploading} />
          </label>
        )}
        {creating && <p className={styles.hint}>Guarda el perfil antes de añadir fotos.</p>}
      </section>
    </main>
  );
}
