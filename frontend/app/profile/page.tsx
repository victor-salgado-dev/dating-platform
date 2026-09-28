'use client';

import { Fragment, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';

import {
  apiFetch,
  ApiError,
  FullProfileEnvelope,
  ProfilePhoto,
  PublicProfile,
  ProfileLanguage,
  InterestDefinition,
  ProfileInterest,
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
  DESIRED_TRAITS_OPTIONS,
  PARTNER_MAY_HAVE_CHILDREN_OPTIONS,
  PARTNER_RELIGION_PREFERENCE_OPTIONS,
  FIRST_MEETING_PREFERENCE_OPTIONS,
  DESIRED_LIVING_PLACE_OPTIONS,
  PARTNER_IMPORTANCE_FIELDS,
  LANGUAGE_OPTIONS,
} from '@/lib/profileOptions';
import { useI18n } from '@/lib/i18n/context';
import { tOption, tOptionList } from '@/lib/i18n/options';
import styles from '../profiles/[id]/page.module.css';

// Evitamos el `unused var` conservando los arrays que no usamos aquí
// directamente (siguen siendo la fuente de verdad de los valores), porque
// en la siguiente iteración los consumirá `profile/edit`.
void HAS_CHILDREN_OPTIONS;
void WANTS_CHILDREN_OPTIONS;
void RELATIONSHIP_GOAL_OPTIONS;
void BODY_TYPE_OPTIONS;
void ETHNICITY_OPTIONS;
void APPEARANCE_RATING_OPTIONS;
void HAIR_COLOR_OPTIONS;
void EYE_COLOR_OPTIONS;
void BODY_ART_OPTIONS;
void SMOKING_HABIT_OPTIONS;
void DRINKING_HABIT_OPTIONS;
void RELOCATION_WILLINGNESS_OPTIONS;
void MARITAL_STATUS_OPTIONS;
void OCCUPATION_OPTIONS;
void EMPLOYMENT_STATUS_OPTIONS;
void INCOME_LEVEL_OPTIONS;
void LIVING_SITUATION_OPTIONS;
void EDUCATION_LEVEL_OPTIONS;
void ENGLISH_ABILITY_OPTIONS;
void RELIGION_OPTIONS;
void RELIGIOUS_VALUES_OPTIONS;
void STAR_SIGN_OPTIONS;
void FUTURE_VISION_OPTIONS;
void SPORTS_OPTIONS;
void LIKES_PETS_OPTIONS;
void PETS_OWNED_OPTIONS;
void FAVORITE_SEASON_OPTIONS;
void IDEAL_VACATION_STYLE_OPTIONS;
void VACATION_ACTIVITIES_OPTIONS;
void DESIRED_TRAITS_OPTIONS;
void PARTNER_MAY_HAVE_CHILDREN_OPTIONS;
void PARTNER_RELIGION_PREFERENCE_OPTIONS;
void FIRST_MEETING_PREFERENCE_OPTIONS;
void DESIRED_LIVING_PLACE_OPTIONS;
void LANGUAGE_OPTIONS;

export default function MyProfilePage() {
  const { dictionary } = useI18n();
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [languages, setLanguages] = useState<ProfileLanguage[]>([]);
  const [interestCatalog, setInterestCatalog] = useState<InterestDefinition[]>([]);
  const [myInterests, setMyInterests] = useState<ProfileInterest[]>([]);
  const [personality, setPersonality] = useState<PersonalityResponse | null>(null);
  const [partnerPrefs, setPartnerPrefs] = useState<PartnerPreferences | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<FullProfileEnvelope>('/profiles/me/full')
      .then((data) => {
        setProfile(data.profile);
        setPhotos(data.photos);
        setLanguages(data.languages);
        setInterestCatalog(data.interest_catalog);
        setMyInterests(data.interests);
        setPersonality(data.personality);
        setPartnerPrefs(data.partner_preferences);
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setError(dictionary.profile.errorNotCreated);
        } else {
          setError(dictionary.profile.loadError);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const interestByKey = useMemo(() => {
    const map: Record<string, InterestDefinition> = {};
    for (const d of interestCatalog) map[d.key] = d;
    return map;
  }, [interestCatalog]);

  const leveledInterests = useMemo(
    () =>
      myInterests
        .filter((i) => i.level != null)
        .map((i) => ({ ...i, def: interestByKey[i.interest_key] }))
        .filter((i) => i.def)
        .sort((a, b) => (b.level ?? 0) - (a.level ?? 0)),
    [myInterests, interestByKey]
  );

  const concreteInterests = useMemo(
    () =>
      myInterests
        .filter((i) => i.level == null)
        .map((i) => interestByKey[i.interest_key])
        .filter((d): d is InterestDefinition => Boolean(d)),
    [myInterests, interestByKey]
  );

  return (
    <main className={styles.main}>
      <Link href="/discover" className={styles.back}>
        {dictionary.common.backToResults}
      </Link>

      {loading && <p>{dictionary.common.loading}</p>}
      {error && (
        <>
          <p className={styles.error}>{error}</p>
          <Link href="/profile/edit">{dictionary.profile.edit}</Link>
        </>
      )}

      {profile && (
        <article>
          <h1>
            {profile.display_name}, {profile.age}
          </h1>

          <p>
            <Link href="/profile/edit" className={styles.editProfileButton}>
              {dictionary.profile.edit}
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
            <dt>{dictionary.profile.fieldGender}</dt><dd>{profile.gender}</dd>
            <dt>{dictionary.profile.fieldHasChildren}</dt><dd>{tOption(dictionary, 'hasChildren', profile.has_children) ?? dictionary.common.notProvided}</dd>
            <dt>{dictionary.profile.fieldWantsChildren}</dt><dd>{tOption(dictionary, 'wantsChildren', profile.wants_children) ?? dictionary.common.notProvided}</dd>
            {profile.relationship_goals && profile.relationship_goals.length > 0 && (
              <><dt>{dictionary.profile.fieldLookingFor}</dt><dd>{tOptionList(dictionary, 'relationshipGoal', profile.relationship_goals)}</dd></>
            )}
          </dl>

          {/* --- Idiomas --- */}
          {languages.length > 0 && (
            <>
              <h2>{dictionary.profile.sectionLanguages}</h2>
              <ul>
                {languages.map((l) => (
                  <li key={l.language_code}>
                    {tOption(dictionary, 'languages', l.language_code) ?? l.language_code}
                    {l.level != null ? ` — ${l.level}/5` : ''}
                  </li>
                ))}
              </ul>
            </>
          )}

          {/* --- Físico y apariencia --- */}
          <h2>{dictionary.profile.sectionPhysical}</h2>
          <dl className={styles.details}>
            {profile.height != null && (<><dt>{dictionary.profile.fieldHeight}</dt><dd>{profile.height} cm</dd></>)}
            {profile.weight != null && (<><dt>{dictionary.profile.fieldWeight}</dt><dd>{profile.weight} kg</dd></>)}
            {profile.body_type && (<><dt>{dictionary.profile.fieldBodyType}</dt><dd>{tOption(dictionary, 'bodyType', profile.body_type)}</dd></>)}
            {profile.ethnicity && (<><dt>{dictionary.profile.fieldEthnicity}</dt><dd>{tOption(dictionary, 'ethnicity', profile.ethnicity)}</dd></>)}
            {profile.appearance_rating && (<><dt>{dictionary.profile.fieldAppearance}</dt><dd>{tOption(dictionary, 'appearanceRating', profile.appearance_rating)}</dd></>)}
            {profile.hair_color && (<><dt>{dictionary.profile.fieldHairColor}</dt><dd>{tOption(dictionary, 'hairColor', profile.hair_color)}</dd></>)}
            {profile.eye_color && (<><dt>{dictionary.profile.fieldEyeColor}</dt><dd>{tOption(dictionary, 'eyeColor', profile.eye_color)}</dd></>)}
            {profile.body_art && profile.body_art.length > 0 && (<><dt>{dictionary.profile.fieldBodyArt}</dt><dd>{tOptionList(dictionary, 'bodyArt', profile.body_art)}</dd></>)}
          </dl>

          {/* --- Estilo de vida y familia --- */}
          <h2>{dictionary.profile.sectionLifestyle}</h2>
          <dl className={styles.details}>
            {profile.smoking_habit && (<><dt>{dictionary.profile.fieldSmoking}</dt><dd>{tOption(dictionary, 'smokingHabit', profile.smoking_habit)}</dd></>)}
            {profile.drinking_habit && (<><dt>{dictionary.profile.fieldDrinking}</dt><dd>{tOption(dictionary, 'drinkingHabit', profile.drinking_habit)}</dd></>)}
            {profile.relocation_willingness && profile.relocation_willingness.length > 0 && (<><dt>{dictionary.profile.fieldRelocation}</dt><dd>{tOptionList(dictionary, 'relocationWillingness', profile.relocation_willingness)}</dd></>)}
            {profile.marital_status && (<><dt>{dictionary.profile.fieldMaritalStatus}</dt><dd>{tOption(dictionary, 'maritalStatus', profile.marital_status)}</dd></>)}
            {profile.children_count != null && (<><dt>{dictionary.profile.fieldChildrenCount}</dt><dd>{profile.children_count}</dd></>)}
            {profile.youngest_child_age != null && (<><dt>{dictionary.profile.fieldYoungestChildAge}</dt><dd>{profile.youngest_child_age}</dd></>)}
            {profile.oldest_child_age != null && (<><dt>{dictionary.profile.fieldOldestChildAge}</dt><dd>{profile.oldest_child_age}</dd></>)}
            {profile.occupation && (<><dt>{dictionary.profile.fieldOccupation}</dt><dd>{tOption(dictionary, 'occupation', profile.occupation)}</dd></>)}
            {profile.employment_status && (<><dt>{dictionary.profile.fieldEmploymentStatus}</dt><dd>{tOption(dictionary, 'employmentStatus', profile.employment_status)}</dd></>)}
            {profile.income_level && (<><dt>{dictionary.profile.fieldIncomeLevel}</dt><dd>{tOption(dictionary, 'incomeLevel', profile.income_level)}</dd></>)}
            {profile.living_situation && (<><dt>{dictionary.profile.fieldLivingSituation}</dt><dd>{tOption(dictionary, 'livingSituation', profile.living_situation)}</dd></>)}
          </dl>

          {/* --- Fondo, cultura y valores --- */}
          <h2>{dictionary.profile.sectionCulture}</h2>
          <dl className={styles.details}>
            {profile.nationality && (<><dt>{dictionary.profile.fieldNationality}</dt><dd>{profile.nationality}</dd></>)}
            {profile.education_level && (<><dt>{dictionary.profile.fieldEducation}</dt><dd>{tOption(dictionary, 'educationLevel', profile.education_level)}</dd></>)}
            {profile.english_ability && (<><dt>{dictionary.profile.fieldEnglish}</dt><dd>{tOption(dictionary, 'englishAbility', profile.english_ability)}</dd></>)}
            {profile.religion && (<><dt>{dictionary.profile.fieldReligion}</dt><dd>{tOption(dictionary, 'religion', profile.religion)}</dd></>)}
            {profile.religious_values && (<><dt>{dictionary.profile.fieldReligiosity}</dt><dd>{tOption(dictionary, 'religiousValues', profile.religious_values)}</dd></>)}
            {profile.star_sign && (<><dt>{dictionary.profile.fieldStarSign}</dt><dd>{tOption(dictionary, 'starSign', profile.star_sign)}</dd></>)}
          </dl>

          {/* --- Über mich / estilo de vida --- */}
          <h2>{dictionary.profile.sectionAboutMe}</h2>
          <dl className={styles.details}>
            {profile.future_vision && profile.future_vision.length > 0 && (<><dt>{dictionary.profile.fieldFutureVision}</dt><dd>{tOptionList(dictionary, 'futureVision', profile.future_vision)}</dd></>)}
            {profile.sports && profile.sports.length > 0 && (<><dt>{dictionary.profile.fieldSports}</dt><dd>{tOptionList(dictionary, 'sports', profile.sports)}</dd></>)}
            {profile.likes_pets && (<><dt>{dictionary.profile.fieldPets}</dt><dd>{tOption(dictionary, 'likesPets', profile.likes_pets)}</dd></>)}
            {profile.pets_owned && profile.pets_owned.length > 0 && (<><dt>{dictionary.profile.fieldPetsOwned}</dt><dd>{tOptionList(dictionary, 'petsOwned', profile.pets_owned)}</dd></>)}
            {profile.favorite_season && (<><dt>{dictionary.profile.fieldFavoriteSeason}</dt><dd>{tOption(dictionary, 'favoriteSeason', profile.favorite_season)}</dd></>)}
            {profile.ideal_vacation_style && profile.ideal_vacation_style.length > 0 && (<><dt>{dictionary.profile.fieldIdealVacation}</dt><dd>{tOptionList(dictionary, 'idealVacationStyle', profile.ideal_vacation_style)}</dd></>)}
            {profile.vacation_activities && profile.vacation_activities.length > 0 && (<><dt>{dictionary.profile.fieldVacationActivities}</dt><dd>{tOptionList(dictionary, 'vacationActivities', profile.vacation_activities)}</dd></>)}
          </dl>
          {profile.dream_wish && <p className={styles.bio}>&ldquo;{profile.dream_wish}&rdquo;</p>}

          {/* --- Intereses: exactamente el formato pedido --- */}
          {leveledInterests.length > 0 && (
            <>
              <h2>{dictionary.profile.sectionInterests}</h2>
              <ul>
                {leveledInterests.map((i) => (
                  <li key={i.interest_key}>
                    {i.def.label} — {i.level}/5
                  </li>
                ))}
              </ul>
            </>
          )}
          {concreteInterests.length > 0 && (
            <>
              <h2>{dictionary.profile.sectionAlsoLikes}</h2>
              <ul>
                {concreteInterests.map((d) => (
                  <li key={d.key}>{d.label}</li>
                ))}
              </ul>
            </>
          )}

          {/* --- Personalidad --- */}
          {personality && personality.trait_scores.length > 0 && (
            <>
              <h2>{dictionary.profile.sectionPersonality}</h2>
              <dl className={styles.details}>
                {personality.trait_scores.map((score) => (
                  <Fragment key={score.trait_key}>
                    <dt>{tOption(dictionary, 'personalityTraits', score.trait_key) ?? score.trait_key}</dt>
                    <dd>
                      {dictionary.profile.personalityScore
                        .replace('{avg}', score.average_score.toFixed(2))
                        .replace('{count}', String(score.answered_count))}
                    </dd>
                  </Fragment>
                ))}
              </dl>
            </>
          )}

          {/* --- Preferencias de pareja --- */}
          {partnerPrefs && (
            <>
              <h2>{dictionary.profile.sectionPartnerPrefs}</h2>
              <dl className={styles.details}>
                {(partnerPrefs.age_min != null || partnerPrefs.age_max != null) && (
                  <><dt>{dictionary.profile.fieldPartnerAge}</dt><dd>{partnerPrefs.age_min ?? '?'} – {partnerPrefs.age_max ?? '?'}</dd></>
                )}
                {(partnerPrefs.height_min != null || partnerPrefs.height_max != null) && (
                  <><dt>{dictionary.profile.fieldPartnerHeight}</dt><dd>{partnerPrefs.height_min ?? '?'} – {partnerPrefs.height_max ?? '?'} cm</dd></>
                )}
                {partnerPrefs.desired_traits && partnerPrefs.desired_traits.length > 0 && (
                  <><dt>{dictionary.profile.fieldPartnerTraits}</dt><dd>{tOptionList(dictionary, 'desiredTraits', partnerPrefs.desired_traits)}</dd></>
                )}
                {partnerPrefs.partner_may_have_children && (
                  <><dt>{dictionary.profile.fieldPartnerChildren}</dt><dd>{tOption(dictionary, 'partnerMayHaveChildren', partnerPrefs.partner_may_have_children)}</dd></>
                )}
                {partnerPrefs.partner_religion_preference && (
                  <><dt>{dictionary.profile.fieldPartnerReligion}</dt><dd>{tOption(dictionary, 'partnerReligionPreference', partnerPrefs.partner_religion_preference)}</dd></>
                )}
                {partnerPrefs.first_meeting_preference && (
                  <><dt>{dictionary.profile.fieldFirstMeeting}</dt><dd>{tOption(dictionary, 'firstMeetingPreference', partnerPrefs.first_meeting_preference)}</dd></>
                )}
                {partnerPrefs.desired_living_place && partnerPrefs.desired_living_place.length > 0 && (
                  <><dt>{dictionary.profile.fieldDesiredLivingPlace}</dt><dd>{tOptionList(dictionary, 'desiredLivingPlace', partnerPrefs.desired_living_place)}</dd></>
                )}
                {PARTNER_IMPORTANCE_FIELDS.filter((f) => (partnerPrefs as unknown as Record<string, number | null>)[f.key] != null).map((f) => (
                  <Fragment key={f.key}>
                    <dt>{tOption(dictionary, 'partnerImportance', f.key) ?? f.key}</dt>
                    <dd>{(partnerPrefs as unknown as Record<string, number | null>)[f.key]}/5</dd>
                  </Fragment>
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
