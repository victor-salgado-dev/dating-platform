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
  LANGUAGE_OPTIONS,
} from '@/lib/profileOptions';
import { useI18n } from '@/lib/i18n/context';
import { tOption } from '@/lib/i18n/options';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
import LocationAutocomplete from '@/components/LocationAutocomplete';
import { PlaceSuggestion } from '@/lib/geocoding';
import styles from './page.module.css';

// =====================================================================
// Tipos y helpers de conversión
// =====================================================================

type FormState = {
  display_name: string;
  birth_date: string;
  gender: string;
  country_code: string;
  region: string;
  relationship_goals: string[];
  has_children: string;
  wants_children: string;
  bio: string;

  height: string;
  weight: string;
  body_type: string;
  ethnicity: string;
  appearance_rating: string;
  hair_color: string;
  eye_color: string;
  body_art: string[];

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

  nationality: string;
  education_level: string;
  english_ability: string;
  religion: string;
  religious_values: string;
  star_sign: string;

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

// =====================================================================
// Componente
// =====================================================================

export default function EditProfilePage() {
  const router = useRouter();
  const { dictionary } = useI18n();
  const [tab, setTab] = useState<
    'basic' | 'physical' | 'lifestyle' | 'background' | 'uber' | 'languages' | 'interests' | 'personality' | 'partner'
  >('basic');

  const [form, setForm] = useState<FormState>(emptyForm);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [creating, setCreating] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const [primaryBusyId, setPrimaryBusyId] = useState<string | null>(null);

  const [myLanguages, setMyLanguages] = useState<ProfileLanguage[]>([]);
  const [languageQuery, setLanguageQuery] = useState('');

  const [interestCatalog, setInterestCatalog] = useState<InterestDefinition[]>([]);
  const [myInterests, setMyInterests] = useState<ProfileInterest[]>([]);
  const [interestQuery, setInterestQuery] = useState('');

  const [personalityCatalog, setPersonalityCatalog] = useState<PersonalityStatement[]>([]);
  const [personalityAnswers, setPersonalityAnswers] = useState<Record<string, number>>({});
  const [traitScores, setTraitScores] = useState<PersonalityTraitScore[]>([]);

  const [partnerForm, setPartnerForm] = useState<PartnerFormState>(emptyPartnerForm);
  const [partnerSaving, setPartnerSaving] = useState(false);
  const [partnerSaved, setPartnerSaved] = useState(false);

  useEffect(() => {
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
          setError(dictionary.profileEdit.errorLoad);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
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
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorSave);
    } finally {
      setSaving(false);
    }
  }

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
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorUploadPhoto);
    } finally {
      setUploading(false);
    }
  }

  async function handleDeletePhoto(photoID: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/photos/${photoID}`, { method: 'DELETE' });
      const refreshed = await apiFetch<ProfilePhoto[]>('/profiles/me/photos');
      setPhotos(refreshed);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorDeletePhoto);
    }
  }

  async function handleSetPrimary(photoID: string) {
    setError(null);
    setPrimaryBusyId(photoID);
    try {
      await apiFetch<void>(`/profiles/me/photos/${photoID}/primary`, {
        method: 'PUT',
      });
      const refreshed = await apiFetch<ProfilePhoto[]>('/profiles/me/photos');
      setPhotos(refreshed);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorSetPrimary);
    } finally {
      setPrimaryBusyId(null);
    }
  }

  async function handleAddLanguage(code: string) {
    setError(null);
    try {
      const savedLang = await apiFetch<ProfileLanguage>(`/profiles/me/languages/${code}`, {
        method: 'PUT',
        body: JSON.stringify({ level: null }),
      });
      setMyLanguages((current) => [...current.filter((l) => l.language_code !== code), savedLang]);
      setLanguageQuery('');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorAddLanguage);
    }
  }

  async function handleLanguageLevelChange(code: string, level: number | null) {
    setError(null);
    try {
      const savedLang = await apiFetch<ProfileLanguage>(`/profiles/me/languages/${code}`, {
        method: 'PUT',
        body: JSON.stringify({ level }),
      });
      setMyLanguages((current) => current.map((l) => (l.language_code === code ? savedLang : l)));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorSaveLevel);
    }
  }

  async function handleRemoveLanguage(code: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/languages/${code}`, { method: 'DELETE' });
      setMyLanguages((current) => current.filter((l) => l.language_code !== code));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorRemoveLanguage);
    }
  }

  const languageSuggestions = useMemo(() => {
    const q = languageQuery.trim().toLowerCase();
    if (q === '') return [];
    const selected = new Set(myLanguages.map((l) => l.language_code));
    return LANGUAGE_OPTIONS.filter(
      (code) => !selected.has(code) && (tOption(dictionary, 'languages', code) ?? code).toLowerCase().includes(q)
    ).slice(0, 8);
  }, [languageQuery, myLanguages, dictionary]);

  async function handleSetInterest(key: string, level: number | null) {
    setError(null);
    try {
      const savedItem = await apiFetch<ProfileInterest>(`/profiles/me/interests/${key}`, {
        method: 'PUT',
        body: JSON.stringify({ level }),
      });
      setMyInterests((current) => [...current.filter((i) => i.interest_key !== key), savedItem]);
      setInterestQuery('');
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorSaveInterest);
    }
  }

  async function handleRemoveInterest(key: string) {
    setError(null);
    try {
      await apiFetch<void>(`/profiles/me/interests/${key}`, { method: 'DELETE' });
      setMyInterests((current) => current.filter((i) => i.interest_key !== key));
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorRemoveInterest);
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
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorSaveAnswer);
    }
  }

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
      const savedPrefs = await apiFetch<PartnerPreferences>('/profiles/me/partner-preferences', {
        method: 'PATCH',
        body: JSON.stringify(body),
      });
      setPartnerForm(partnerFormFromPreferences(savedPrefs));
      setPartnerSaved(true);
    } catch (err) {
      setError(err instanceof ApiError ? err.message : dictionary.profileEdit.errorSavePartner);
    } finally {
      setPartnerSaving(false);
    }
  }

  function renderSelect<K extends keyof FormState>(
    field: K,
    label: string,
    category: keyof Dictionary['options'],
    values: string[]
  ) {
    return (
      <label>
        {label}
        <select value={form[field] as string} onChange={(e) => updateField(field, e.target.value as never)}>
          <option value="">{dictionary.profileEdit.preferNotToSay}</option>
          {values.map((value) => (
            <option key={value} value={value}>{tOption(dictionary, category, value)}</option>
          ))}
        </select>
      </label>
    );
  }

  function renderCheckboxGroup(
    field: MultiSelectField,
    label: string,
    category: keyof Dictionary['options'],
    values: string[]
  ) {
    return (
      <fieldset className={styles.checkboxFieldset}>
        <legend>{label}</legend>
        <div className={styles.checkboxGrid}>
          {values.map((value) => (
            <label key={value} className={styles.checkboxLabel}>
              <input
                type="checkbox"
                checked={form[field].includes(value)}
                onChange={() => toggleMulti(field, value)}
              />
              {tOption(dictionary, category, value)}
            </label>
          ))}
        </div>
      </fieldset>
    );
  }

  if (loading) {
    return <main className={styles.main}><p>{dictionary.common.loading}</p></main>;
  }

  const disabledUntilCreated = creating;

  const tabs: { id: typeof tab; label: string }[] = [
    { id: 'basic', label: dictionary.profileEdit.tabBasic },
    { id: 'physical', label: dictionary.profileEdit.tabPhysical },
    { id: 'lifestyle', label: dictionary.profileEdit.tabLifestyle },
    { id: 'background', label: dictionary.profileEdit.tabBackground },
    { id: 'uber', label: dictionary.profileEdit.tabAboutMe },
    { id: 'languages', label: dictionary.profileEdit.tabLanguages },
    { id: 'interests', label: dictionary.profileEdit.tabInterests },
    { id: 'personality', label: dictionary.profileEdit.tabPersonality },
    { id: 'partner', label: dictionary.profileEdit.tabPartner },
  ];

  const locationSaved = dictionary.profileEdit.locationSaved
    .replace('{region}', form.region ? `${form.region}, ` : '')
    .replace('{country}', form.country_code);

  const primaryPhotoId = photos.reduce<ProfilePhoto | null>(
    (min, p) => (min === null || p.position < min.position ? p : min),
    null
  )?.id ?? null;

  return (
    <main className={styles.main}>
      <Link href="/profile">{dictionary.profileEdit.back}</Link>
      <h1>{creating ? dictionary.profileEdit.titleCreate : dictionary.profileEdit.titleEdit}</h1>

      <nav className={styles.tabs}>
        {tabs.map((t) => (
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
              <label>{dictionary.profileEdit.fieldDisplayName}<input value={form.display_name} onChange={(e) => updateField('display_name', e.target.value)} required /></label>
              {creating && <label>{dictionary.profileEdit.fieldBirthDate}<input type="date" value={form.birth_date} onChange={(e) => updateField('birth_date', e.target.value)} required /></label>}
              <label>{dictionary.profileEdit.fieldGender}
                <select value={form.gender} onChange={(e) => updateField('gender', e.target.value)} required>
                  <option value="">{dictionary.profileEdit.selectOption}</option>
                  {(['female', 'male', 'non_binary', 'other'] as const).map((value) => (
                    <option key={value} value={value}>{tOption(dictionary, 'gender', value)}</option>
                  ))}
                </select>
              </label>

              <label>{dictionary.profileEdit.fieldLocation}
                <LocationAutocomplete
                  initialValue={form.region && form.country_code ? `${form.region}, ${form.country_code}` : ''}
                  onSelect={handleLocationSelect}
                />
              </label>
              {form.country_code && (
                <p className={styles.hint}>{locationSaved}</p>
              )}

              <fieldset className={styles.checkboxFieldset}>
                <legend>{dictionary.profileEdit.fieldRelationshipGoals}</legend>
                <div className={styles.checkboxGrid}>
                  {RELATIONSHIP_GOAL_OPTIONS.map((value) => (
                    <label key={value} className={styles.checkboxLabel}>
                      <input
                        type="checkbox"
                        checked={form.relationship_goals.includes(value)}
                        onChange={() => toggleRelationshipGoal(value)}
                      />
                      {tOption(dictionary, 'relationshipGoal', value)}
                    </label>
                  ))}
                </div>
              </fieldset>

              {renderSelect('has_children', dictionary.profileEdit.fieldHasChildren, 'hasChildren', HAS_CHILDREN_OPTIONS)}
              {renderSelect('wants_children', dictionary.profileEdit.fieldWantsChildren, 'wantsChildren', WANTS_CHILDREN_OPTIONS)}
              <label>{dictionary.profileEdit.fieldBio}<textarea value={form.bio} onChange={(e) => updateField('bio', e.target.value)} maxLength={1000} rows={5} /></label>
            </>
          )}

          {tab === 'physical' && (
            <>
              <label>{dictionary.profileEdit.fieldHeight}<input type="number" min={50} max={300} value={form.height} onChange={(e) => updateField('height', e.target.value)} /></label>
              <label>{dictionary.profileEdit.fieldWeight}<input type="number" min={20} max={400} value={form.weight} onChange={(e) => updateField('weight', e.target.value)} /></label>
              {renderSelect('body_type', dictionary.profileEdit.fieldBodyType, 'bodyType', BODY_TYPE_OPTIONS)}
              {renderSelect('ethnicity', dictionary.profileEdit.fieldEthnicity, 'ethnicity', ETHNICITY_OPTIONS)}
              {renderSelect('appearance_rating', dictionary.profileEdit.fieldAppearance, 'appearanceRating', APPEARANCE_RATING_OPTIONS)}
              {renderSelect('hair_color', dictionary.profileEdit.fieldHairColor, 'hairColor', HAIR_COLOR_OPTIONS)}
              {renderSelect('eye_color', dictionary.profileEdit.fieldEyeColor, 'eyeColor', EYE_COLOR_OPTIONS)}
              {renderCheckboxGroup('body_art', dictionary.profileEdit.fieldBodyArt, 'bodyArt', BODY_ART_OPTIONS)}
            </>
          )}

          {tab === 'lifestyle' && (
            <>
              {renderSelect('smoking_habit', dictionary.profileEdit.fieldSmoking, 'smokingHabit', SMOKING_HABIT_OPTIONS)}
              {renderSelect('drinking_habit', dictionary.profileEdit.fieldDrinking, 'drinkingHabit', DRINKING_HABIT_OPTIONS)}
              {renderCheckboxGroup('relocation_willingness', dictionary.profileEdit.fieldRelocation, 'relocationWillingness', RELOCATION_WILLINGNESS_OPTIONS)}
              {renderSelect('marital_status', dictionary.profileEdit.fieldMaritalStatus, 'maritalStatus', MARITAL_STATUS_OPTIONS)}
              <label>{dictionary.profileEdit.fieldChildrenCount}<input type="number" min={0} max={25} value={form.children_count} onChange={(e) => updateField('children_count', e.target.value)} /></label>
              <label>{dictionary.profileEdit.fieldYoungestChildAge}<input type="number" min={0} max={100} value={form.youngest_child_age} onChange={(e) => updateField('youngest_child_age', e.target.value)} /></label>
              <label>{dictionary.profileEdit.fieldOldestChildAge}<input type="number" min={0} max={100} value={form.oldest_child_age} onChange={(e) => updateField('oldest_child_age', e.target.value)} /></label>
              {renderSelect('occupation', dictionary.profileEdit.fieldOccupation, 'occupation', OCCUPATION_OPTIONS)}
              {renderSelect('employment_status', dictionary.profileEdit.fieldEmploymentStatus, 'employmentStatus', EMPLOYMENT_STATUS_OPTIONS)}
              {renderSelect('income_level', dictionary.profileEdit.fieldIncomeLevel, 'incomeLevel', INCOME_LEVEL_OPTIONS)}
              {renderSelect('living_situation', dictionary.profileEdit.fieldLivingSituation, 'livingSituation', LIVING_SITUATION_OPTIONS)}
            </>
          )}

          {tab === 'background' && (
            <>
              <label>{dictionary.profileEdit.fieldNationality}<input value={form.nationality} onChange={(e) => updateField('nationality', e.target.value.toUpperCase())} maxLength={2} /></label>
              {renderSelect('education_level', dictionary.profileEdit.fieldEducation, 'educationLevel', EDUCATION_LEVEL_OPTIONS)}
              {renderSelect('english_ability', dictionary.profileEdit.fieldEnglish, 'englishAbility', ENGLISH_ABILITY_OPTIONS)}
              {renderSelect('religion', dictionary.profileEdit.fieldReligion, 'religion', RELIGION_OPTIONS)}
              {renderSelect('religious_values', dictionary.profileEdit.fieldReligiousValues, 'religiousValues', RELIGIOUS_VALUES_OPTIONS)}
              {renderSelect('star_sign', dictionary.profileEdit.fieldStarSign, 'starSign', STAR_SIGN_OPTIONS)}
            </>
          )}

          {tab === 'uber' && (
            <>
              <label>{dictionary.profileEdit.fieldProfileQuote} <span className={styles.hint}>{dictionary.profileEdit.maxCharsShort}</span><textarea value={form.profile_quote} onChange={(e) => updateField('profile_quote', e.target.value)} maxLength={500} rows={3} /></label>
              {renderCheckboxGroup('future_vision', dictionary.profileEdit.fieldFutureVision, 'futureVision', FUTURE_VISION_OPTIONS)}
              {renderCheckboxGroup('sports', dictionary.profileEdit.fieldSports, 'sports', SPORTS_OPTIONS)}
              {renderSelect('likes_pets', dictionary.profileEdit.fieldLikesPets, 'likesPets', LIKES_PETS_OPTIONS)}
              {renderCheckboxGroup('pets_owned', dictionary.profileEdit.fieldPetsOwned, 'petsOwned', PETS_OWNED_OPTIONS)}
              {renderSelect('favorite_season', dictionary.profileEdit.fieldFavoriteSeason, 'favoriteSeason', FAVORITE_SEASON_OPTIONS)}
              {renderCheckboxGroup('ideal_vacation_style', dictionary.profileEdit.fieldIdealVacation, 'idealVacationStyle', IDEAL_VACATION_STYLE_OPTIONS)}
              {renderCheckboxGroup('vacation_activities', dictionary.profileEdit.fieldVacationActivities, 'vacationActivities', VACATION_ACTIVITIES_OPTIONS)}
              <label>{dictionary.profileEdit.fieldDreamWish} <span className={styles.hint}>{dictionary.profileEdit.maxCharsShort}</span><textarea value={form.dream_wish} onChange={(e) => updateField('dream_wish', e.target.value)} maxLength={500} rows={3} /></label>
            </>
          )}

          {saved && <p className={styles.success}>{dictionary.profileEdit.saved}</p>}
          <button type="submit" disabled={saving}>{saving ? dictionary.profileEdit.saving : dictionary.profileEdit.save}</button>
        </form>
      )}

      {/* --- Idiomas --- */}
      {tab === 'languages' && (
        disabledUntilCreated ? (
          <p className={styles.hint}>{dictionary.profileEdit.languagesMustSaveFirst}</p>
        ) : (
          <div className={styles.hobbyList}>
            {myLanguages.length > 0 && (
              <fieldset className={styles.checkboxFieldset}>
                <legend>{dictionary.profileEdit.languagesLegend}</legend>
                {myLanguages.map((l) => (
                  <div key={l.language_code} className={styles.hobbyRow}>
                    <span className={styles.hobbyLabel}>
                      {tOption(dictionary, 'languages', l.language_code) ?? l.language_code}
                    </span>
                    <select
                      value={l.level ?? ''}
                      onChange={(e) => handleLanguageLevelChange(l.language_code, e.target.value ? Number(e.target.value) : null)}
                    >
                      <option value="">{dictionary.profileEdit.noLevel}</option>
                      {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                    </select>
                    <button type="button" onClick={() => handleRemoveLanguage(l.language_code)}>{dictionary.profileEdit.remove}</button>
                  </div>
                ))}
              </fieldset>
            )}

            <label>{dictionary.profileEdit.addLanguage}
              <input
                value={languageQuery}
                onChange={(e) => setLanguageQuery(e.target.value)}
                placeholder={dictionary.profileEdit.searchLanguage}
              />
            </label>
            {languageSuggestions.length > 0 && (
              <ul className={styles.suggestionList}>
                {languageSuggestions.map((code) => (
                  <li key={code}>
                    <button type="button" onClick={() => handleAddLanguage(code)}>
                      {tOption(dictionary, 'languages', code) ?? code}
                    </button>
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
          <p className={styles.hint}>{dictionary.profileEdit.interestsMustSaveFirst}</p>
        ) : (
          <div className={styles.hobbyList}>
            <h2>{dictionary.profileEdit.primaryInterests}</h2>
            <p className={styles.hint}>{dictionary.profileEdit.primaryInterestsHint}</p>
            {Object.entries(
              leveledInterests.reduce<Record<string, InterestDefinition[]>>((acc, d) => {
                (acc[d.category] ??= []).push(d);
                return acc;
              }, {})
            ).map(([category, defs]) => (
              <fieldset key={category} className={styles.checkboxFieldset}>
                <legend>{tOption(dictionary, 'interestCategories', category) ?? category}</legend>
                {defs.map((d) => (
                  <div key={d.key} className={styles.hobbyRow}>
                    <span className={styles.hobbyLabel}>{d.label}</span>
                    <select
                      value={myInterestByKey[d.key]?.level ?? ''}
                      onChange={(e) => (e.target.value ? handleSetInterest(d.key, Number(e.target.value)) : handleRemoveInterest(d.key))}
                    >
                      <option value="">{dictionary.profileEdit.unmarked}</option>
                      {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                    </select>
                  </div>
                ))}
              </fieldset>
            ))}

            <h2>{dictionary.profileEdit.otherInterests}</h2>
            <p className={styles.hint}>{dictionary.profileEdit.otherInterestsHint}</p>

            {selectedConcreteInterests.length > 0 && (
              <div className={styles.chipRow}>
                {selectedConcreteInterests.map((d) => (
                  <span key={d.key} className={styles.chip}>
                    {d.label}
                    <button
                      type="button"
                      onClick={() => handleRemoveInterest(d.key)}
                      aria-label={dictionary.profileEdit.removeAria.replace('{label}', d.label)}
                    >×</button>
                  </span>
                ))}
              </div>
            )}

            <label>{dictionary.profileEdit.searchInterest}
              <input
                value={interestQuery}
                onChange={(e) => setInterestQuery(e.target.value)}
                placeholder={dictionary.profileEdit.searchInterestPlaceholder}
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
          <p className={styles.hint}>{dictionary.profileEdit.personalityMustSaveFirst}</p>
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
                    {tOption(dictionary, 'personalityTraits', trait) ?? trait}
                    {score && (
                      <span className={styles.hint}>
                        {dictionary.profileEdit.personalityAverage
                          .replace('{avg}', score.average_score.toFixed(2))
                          .replace('{count}', String(score.answered_count))}
                      </span>
                    )}
                  </legend>
                  {statements.map((s) => (
                    <div key={s.key} className={styles.hobbyRow}>
                      <span className={styles.hobbyLabel}>{s.label}</span>
                      <select
                        value={personalityAnswers[s.key] ?? ''}
                        onChange={(e) => handlePersonalityChange(s.key, Number(e.target.value))}
                      >
                        <option value="">{dictionary.profileEdit.unanswered}</option>
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
          <p className={styles.hint}>{dictionary.profileEdit.partnerMustSaveFirst}</p>
        ) : (
          <form onSubmit={handleSavePartnerPreferences} className={styles.form}>
            <label>{dictionary.profileEdit.fieldPartnerAgeMin}<input type="number" min={18} max={120} value={partnerForm.age_min} onChange={(e) => updatePartnerField('age_min', e.target.value)} /></label>
            <label>{dictionary.profileEdit.fieldPartnerAgeMax}<input type="number" min={18} max={120} value={partnerForm.age_max} onChange={(e) => updatePartnerField('age_max', e.target.value)} /></label>
            <label>{dictionary.profileEdit.fieldPartnerHeightMin}<input type="number" min={50} max={300} value={partnerForm.height_min} onChange={(e) => updatePartnerField('height_min', e.target.value)} /></label>
            <label>{dictionary.profileEdit.fieldPartnerHeightMax}<input type="number" min={50} max={300} value={partnerForm.height_max} onChange={(e) => updatePartnerField('height_max', e.target.value)} /></label>

            <fieldset className={styles.checkboxFieldset}>
              <legend>{dictionary.profileEdit.fieldPartnerTraits}</legend>
              <div className={styles.checkboxGrid}>
                {DESIRED_TRAITS_OPTIONS.map((value) => (
                  <label key={value} className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={partnerForm.desired_traits.includes(value)}
                      onChange={() => togglePartnerMulti('desired_traits', value)}
                    />
                    {tOption(dictionary, 'desiredTraits', value)}
                  </label>
                ))}
              </div>
            </fieldset>

            <label>{dictionary.profileEdit.fieldPartnerMayHaveChildren}
              <select value={partnerForm.partner_may_have_children} onChange={(e) => updatePartnerField('partner_may_have_children', e.target.value)}>
                <option value="">{dictionary.profileEdit.preferNotToSay}</option>
                {PARTNER_MAY_HAVE_CHILDREN_OPTIONS.map((value) => (
                  <option key={value} value={value}>{tOption(dictionary, 'partnerMayHaveChildren', value)}</option>
                ))}
              </select>
            </label>

            <label>{dictionary.profileEdit.fieldPartnerReligion}
              <select value={partnerForm.partner_religion_preference} onChange={(e) => updatePartnerField('partner_religion_preference', e.target.value)}>
                <option value="">{dictionary.profileEdit.preferNotToSay}</option>
                {PARTNER_RELIGION_PREFERENCE_OPTIONS.map((value) => (
                  <option key={value} value={value}>{tOption(dictionary, 'partnerReligionPreference', value)}</option>
                ))}
              </select>
            </label>

            <label>{dictionary.profileEdit.fieldPartnerAboutText} <span className={styles.hint}>{dictionary.profileEdit.maxCharsLong}</span>
              <textarea value={partnerForm.about_partner_text} onChange={(e) => updatePartnerField('about_partner_text', e.target.value)} maxLength={1000} rows={4} />
            </label>

            <label>{dictionary.profileEdit.fieldFirstMeeting}
              <select value={partnerForm.first_meeting_preference} onChange={(e) => updatePartnerField('first_meeting_preference', e.target.value)}>
                <option value="">{dictionary.profileEdit.preferNotToSay}</option>
                {FIRST_MEETING_PREFERENCE_OPTIONS.map((value) => (
                  <option key={value} value={value}>{tOption(dictionary, 'firstMeetingPreference', value)}</option>
                ))}
              </select>
            </label>

            <fieldset className={styles.checkboxFieldset}>
              <legend>{dictionary.profileEdit.fieldDesiredLivingPlace}</legend>
              <div className={styles.checkboxGrid}>
                {DESIRED_LIVING_PLACE_OPTIONS.map((value) => (
                  <label key={value} className={styles.checkboxLabel}>
                    <input
                      type="checkbox"
                      checked={partnerForm.desired_living_place.includes(value)}
                      onChange={() => togglePartnerMulti('desired_living_place', value)}
                    />
                    {tOption(dictionary, 'desiredLivingPlace', value)}
                  </label>
                ))}
              </div>
            </fieldset>

            <fieldset className={styles.checkboxFieldset}>
              <legend>{dictionary.profileEdit.fieldPartnerImportance} <span className={styles.hint}>{dictionary.profileEdit.partnerImportanceHint}</span></legend>
              {PARTNER_IMPORTANCE_FIELDS.map((f) => (
                <div key={f.key} className={styles.hobbyRow}>
                  <span className={styles.hobbyLabel}>{tOption(dictionary, 'partnerImportance', f.key) ?? f.key}</span>
                  <select
                    value={partnerForm[f.key as keyof PartnerFormState]}
                    onChange={(e) => updatePartnerField(f.key as keyof PartnerFormState, e.target.value)}
                  >
                    <option value="">{dictionary.profileEdit.unanswered}</option>
                    {[1, 2, 3, 4, 5].map((n) => <option key={n} value={n}>{n}</option>)}
                  </select>
                </div>
              ))}
            </fieldset>

            {partnerSaved && <p className={styles.success}>{dictionary.profileEdit.partnerSaved}</p>}
            <button type="submit" disabled={partnerSaving}>{partnerSaving ? dictionary.profileEdit.saving : dictionary.profileEdit.partnerSave}</button>
          </form>
        )
      )}

      {/* --- Fotos: visibles en cualquier pestaña, al final --- */}
      <section className={styles.photos}>
        <h2>{dictionary.profileEdit.photosTitle}</h2>
        {photos.length > 0 && (
          <div className={styles.photoGrid}>
            {photos.map((photo) => (
              <div key={photo.id} className={styles.photo}>
                <img src={photo.url} alt="" />
                <div>
                  {photo.id === primaryPhotoId && (
                    <span className={styles.hint}>{dictionary.profileEdit.primaryPhoto}</span>
                  )}
                  {photos.length > 1 && photo.id !== primaryPhotoId && (
                    <button
                      type="button"
                      onClick={() => handleSetPrimary(photo.id)}
                      disabled={primaryBusyId === photo.id}
                    >
                      {primaryBusyId === photo.id
                        ? dictionary.profileEdit.settingPrimary
                        : dictionary.profileEdit.setPrimary}
                    </button>
                  )}
                  <button type="button" onClick={() => handleDeletePhoto(photo.id)}>{dictionary.profileEdit.deletePhoto}</button>
                </div>
              </div>
            ))}
          </div>
        )}
        {!creating && (
          <label className={styles.upload}>
            {dictionary.profileEdit.addPhoto}
            <input type="file" accept="image/jpeg,image/png,image/webp" onChange={handleUpload} disabled={uploading} />
          </label>
        )}
        {creating && <p className={styles.hint}>{dictionary.profileEdit.photosMustSaveFirst}</p>}
      </section>
    </main>
  );
}
