'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import {
  apiFetch,
  ApiError,
  ProfilePhoto,
  PublicProfile,
  HobbyDefinition,
  ProfileHobby,
  PersonalityResponse,
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
  PARTNER_IMPORTANCE_FIELDS,
  PERSONALITY_TRAIT_LABELS,
  Option,
} from '@/lib/profileOptions';
import styles from '../profiles/[id]/page.module.css';

// label(s) busca la etiqueta legible de un valor (o lista de valores)
// dentro de una lista de opciones; si no la encuentra, muestra el valor
// en crudo en vez de nada, para no ocultar datos por un desajuste de
// catálogo.
function label(value: string | null, options: Option[]): string | null {
  if (!value) return null;
  return options.find((o) => o.value === value)?.label ?? value;
}

function labelList(values: string[] | null, options: Option[]): string | null {
  if (!values || values.length === 0) return null;
  return values.map((v) => options.find((o) => o.value === v)?.label ?? v).join(', ');
}

export default function MyProfilePage() {
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [hobbyCatalog, setHobbyCatalog] = useState<HobbyDefinition[]>([]);
  const [myHobbies, setMyHobbies] = useState<ProfileHobby[]>([]);
  const [personality, setPersonality] = useState<PersonalityResponse | null>(null);
  const [partnerPrefs, setPartnerPrefs] = useState<PartnerPreferences | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<HobbyDefinition[]>('/catalog/hobbies').then(setHobbyCatalog).catch(() => setHobbyCatalog([]));

    apiFetch<PublicProfile>('/profiles/me')
      .then(async (data) => {
        setProfile(data);
        try {
          setPhotos(await apiFetch<ProfilePhoto[]>('/profiles/me/photos'));
        } catch {
          setPhotos([]);
        }
        try {
          setMyHobbies(await apiFetch<ProfileHobby[]>('/profiles/me/hobbies'));
        } catch {
          setMyHobbies([]);
        }
        try {
          setPersonality(await apiFetch<PersonalityResponse>('/profiles/me/personality'));
        } catch {
          setPersonality(null);
        }
        try {
          setPartnerPrefs(await apiFetch<PartnerPreferences>('/profiles/me/partner-preferences'));
        } catch {
          setPartnerPrefs(null);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setError('Aún no has creado tu perfil.');
        } else {
          setError('No se pudo cargar tu perfil.');
        }
      })
      .finally(() => setLoading(false));
  }, []);

  const hobbyLabel = (key: string) => hobbyCatalog.find((h) => h.key === key)?.label ?? key;
  const likedHobbies = myHobbies.filter((h) => h.liked);
  const dislikedHobbies = myHobbies.filter((h) => !h.liked);

  return (
    <main className={styles.main}>
      <Link href="/discover" className={styles.back}>
        &larr; Volver a resultados
      </Link>

      {loading && <p>Cargando…</p>}
      {error && (
        <>
          <p className={styles.error}>{error}</p>
          <Link href="/profile/edit">Modificar perfil</Link>
        </>
      )}

      {profile && (
        <article>
          <h1>
            {profile.display_name}, {profile.age}
          </h1>

          <p>
            <Link href="/profile/edit" className={styles.editProfileButton}>
              Modificar perfil
            </Link>
          </p>

          <p className={styles.location}>
            {[profile.region, profile.country_code].filter(Boolean).join(', ')}
          </p>

          {photos.length > 0 && (
            <div className={styles.photos}>
              {photos.map((photo) => (
                <img key={photo.id} src={photo.url} alt="" className={styles.photo} />
              ))}
            </div>
          )}

          {profile.profile_quote && <p className={styles.bio}>&ldquo;{profile.profile_quote}&rdquo;</p>}
          {profile.bio && <p className={styles.bio}>{profile.bio}</p>}

          {/* --- Básico --- */}
          <dl className={styles.details}>
            <dt>Género</dt><dd>{profile.gender}</dd>
            <dt>¿Tiene hijos?</dt><dd>{label(profile.has_children, HAS_CHILDREN_OPTIONS) ?? 'No indicado'}</dd>
            <dt>¿Quiere tener hijos?</dt><dd>{label(profile.wants_children, WANTS_CHILDREN_OPTIONS) ?? 'No indicado'}</dd>
            {profile.relationship_goal && (<><dt>Busca</dt><dd>{label(profile.relationship_goal, RELATIONSHIP_GOAL_OPTIONS)}</dd></>)}
            {profile.languages && profile.languages.length > 0 && (<><dt>Idiomas</dt><dd>{profile.languages.join(', ')}</dd></>)}
            {profile.interests && profile.interests.length > 0 && (<><dt>Intereses</dt><dd>{profile.interests.join(', ')}</dd></>)}
          </dl>

          {/* --- Físico y apariencia --- */}
          <h2>Físico y apariencia</h2>
          <dl className={styles.details}>
            {profile.height != null && (<><dt>Altura</dt><dd>{profile.height} cm</dd></>)}
            {profile.weight != null && (<><dt>Peso</dt><dd>{profile.weight} kg</dd></>)}
            {profile.body_type && (<><dt>Complexión</dt><dd>{label(profile.body_type, BODY_TYPE_OPTIONS)}</dd></>)}
            {profile.ethnicity && (<><dt>Etnia</dt><dd>{label(profile.ethnicity, ETHNICITY_OPTIONS)}</dd></>)}
            {profile.appearance_rating && (<><dt>Aspecto</dt><dd>{label(profile.appearance_rating, APPEARANCE_RATING_OPTIONS)}</dd></>)}
            {profile.hair_color && (<><dt>Pelo</dt><dd>{label(profile.hair_color, HAIR_COLOR_OPTIONS)}</dd></>)}
            {profile.eye_color && (<><dt>Ojos</dt><dd>{label(profile.eye_color, EYE_COLOR_OPTIONS)}</dd></>)}
            {profile.body_art && profile.body_art.length > 0 && (<><dt>Piercings / tatuajes</dt><dd>{labelList(profile.body_art, BODY_ART_OPTIONS)}</dd></>)}
          </dl>

          {/* --- Estilo de vida y familia --- */}
          <h2>Estilo de vida y familia</h2>
          <dl className={styles.details}>
            {profile.smoking_habit && (<><dt>Fuma</dt><dd>{label(profile.smoking_habit, SMOKING_HABIT_OPTIONS)}</dd></>)}
            {profile.drinking_habit && (<><dt>Bebe alcohol</dt><dd>{label(profile.drinking_habit, DRINKING_HABIT_OPTIONS)}</dd></>)}
            {profile.relocation_willingness && profile.relocation_willingness.length > 0 && (<><dt>Dispuesto/a a mudarse</dt><dd>{labelList(profile.relocation_willingness, RELOCATION_WILLINGNESS_OPTIONS)}</dd></>)}
            {profile.marital_status && (<><dt>Estado civil</dt><dd>{label(profile.marital_status, MARITAL_STATUS_OPTIONS)}</dd></>)}
            {profile.children_count != null && (<><dt>Número de hijos</dt><dd>{profile.children_count}</dd></>)}
            {profile.occupation && (<><dt>Ocupación</dt><dd>{label(profile.occupation, OCCUPATION_OPTIONS)}</dd></>)}
            {profile.employment_status && (<><dt>Situación laboral</dt><dd>{label(profile.employment_status, EMPLOYMENT_STATUS_OPTIONS)}</dd></>)}
            {profile.income_level && (<><dt>Nivel de ingresos</dt><dd>{label(profile.income_level, INCOME_LEVEL_OPTIONS)}</dd></>)}
            {profile.living_situation && (<><dt>Con quién vive</dt><dd>{label(profile.living_situation, LIVING_SITUATION_OPTIONS)}</dd></>)}
          </dl>

          {/* --- Fondo, cultura y valores --- */}
          <h2>Fondo, cultura y valores</h2>
          <dl className={styles.details}>
            {profile.nationality && (<><dt>Nacionalidad</dt><dd>{profile.nationality}</dd></>)}
            {profile.education_level && (<><dt>Nivel educativo</dt><dd>{label(profile.education_level, EDUCATION_LEVEL_OPTIONS)}</dd></>)}
            {profile.english_ability && (<><dt>Nivel de inglés</dt><dd>{label(profile.english_ability, ENGLISH_ABILITY_OPTIONS)}</dd></>)}
            {profile.religion && (<><dt>Religión</dt><dd>{label(profile.religion, RELIGION_OPTIONS)}</dd></>)}
            {profile.religious_values && (<><dt>Religiosidad</dt><dd>{label(profile.religious_values, RELIGIOUS_VALUES_OPTIONS)}</dd></>)}
            {profile.star_sign && (<><dt>Signo</dt><dd>{label(profile.star_sign, STAR_SIGN_OPTIONS)}</dd></>)}
          </dl>

          {/* --- Über mich / estilo de vida --- */}
          <h2>Über mich</h2>
          <dl className={styles.details}>
            {profile.future_vision && profile.future_vision.length > 0 && (<><dt>Cómo te imaginas el futuro</dt><dd>{labelList(profile.future_vision, FUTURE_VISION_OPTIONS)}</dd></>)}
            {profile.sports && profile.sports.length > 0 && (<><dt>Deportes</dt><dd>{labelList(profile.sports, SPORTS_OPTIONS)}</dd></>)}
            {profile.likes_pets && (<><dt>Mascotas</dt><dd>{label(profile.likes_pets, LIKES_PETS_OPTIONS)}</dd></>)}
            {profile.pets_owned && profile.pets_owned.length > 0 && (<><dt>Mascotas que tiene</dt><dd>{labelList(profile.pets_owned, PETS_OWNED_OPTIONS)}</dd></>)}
            {profile.favorite_season && (<><dt>Estación favorita</dt><dd>{label(profile.favorite_season, FAVORITE_SEASON_OPTIONS)}</dd></>)}
            {profile.ideal_vacation_style && profile.ideal_vacation_style.length > 0 && (<><dt>Vacaciones ideales</dt><dd>{labelList(profile.ideal_vacation_style, IDEAL_VACATION_STYLE_OPTIONS)}</dd></>)}
            {profile.vacation_activities && profile.vacation_activities.length > 0 && (<><dt>Actividades de vacaciones</dt><dd>{labelList(profile.vacation_activities, VACATION_ACTIVITIES_OPTIONS)}</dd></>)}
          </dl>
          {profile.dream_wish && <p className={styles.bio}>&ldquo;{profile.dream_wish}&rdquo;</p>}

          {/* --- Hobbies --- */}
          {(likedHobbies.length > 0 || dislikedHobbies.length > 0) && (
            <>
              <h2>Hobbies</h2>
              {likedHobbies.length > 0 && (
                <ul>
                  {likedHobbies.map((h) => (
                    <li key={h.hobby_key}>
                      {hobbyLabel(h.hobby_key)}{h.intensity != null ? ` — intensidad ${h.intensity}/5` : ''}
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}

          {/* --- Personalidad --- */}
          {personality && personality.trait_scores.length > 0 && (
            <>
              <h2>Personalidad</h2>
              <dl className={styles.details}>
                {personality.trait_scores.map((score) => (
                  <>
                    <dt key={`${score.trait_key}-dt`}>{PERSONALITY_TRAIT_LABELS[score.trait_key] ?? score.trait_key}</dt>
                    <dd key={`${score.trait_key}-dd`}>{score.average_score.toFixed(2)} / 5 ({score.answered_count} respuestas)</dd>
                  </>
                ))}
              </dl>
            </>
          )}

          {/* --- Preferencias de pareja --- */}
          {partnerPrefs && (
            <>
              <h2>Lo que busco en una pareja</h2>
              <dl className={styles.details}>
                {(partnerPrefs.age_min != null || partnerPrefs.age_max != null) && (
                  <><dt>Edad</dt><dd>{partnerPrefs.age_min ?? '?'} – {partnerPrefs.age_max ?? '?'}</dd></>
                )}
                {(partnerPrefs.height_min != null || partnerPrefs.height_max != null) && (
                  <><dt>Altura</dt><dd>{partnerPrefs.height_min ?? '?'} – {partnerPrefs.height_max ?? '?'} cm</dd></>
                )}
                {partnerPrefs.desired_traits && partnerPrefs.desired_traits.length > 0 && (
                  <><dt>Rasgos que busco</dt><dd>{partnerPrefs.desired_traits.join(', ')}</dd></>
                )}
                {PARTNER_IMPORTANCE_FIELDS.filter((f) => (partnerPrefs as unknown as Record<string, number | null>)[f.key] != null).map((f) => (
                  <>
                    <dt key={`${f.key}-dt`}>{f.label}</dt>
                    <dd key={`${f.key}-dd`}>{(partnerPrefs as unknown as Record<string, number | null>)[f.key]}/5</dd>
                  </>
                ))}
              </dl>
              {partnerPrefs.about_partner_text && <p className={styles.bio}>{partnerPrefs.about_partner_text}</p>}
            </>
          )}
        </article>
      )}
    </main>
  );
}
