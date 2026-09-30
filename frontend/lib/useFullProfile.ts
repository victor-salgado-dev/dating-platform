'use client';

import { useEffect, useState } from 'react';
import { apiFetch, apiMutateQuiet, peekApiCache } from '@/lib/api';
import type {
  FullProfileEnvelope,
  PublicProfile,
  ProfilePhoto,
  ProfileLanguage,
  InterestDefinition,
  ProfileInterest,
  PersonalityResponse,
  PartnerPreferences,
} from '@/lib/api';

export interface FullProfileData {
  profile: PublicProfile | null;
  photos: ProfilePhoto[];
  languages: ProfileLanguage[];
  interestCatalog: InterestDefinition[];
  theirInterests: ProfileInterest[];
  personality: PersonalityResponse | null;
  partnerPrefs: PartnerPreferences | null;
  favorited: boolean;
  liked: boolean;
  matched: boolean;
  blocked: boolean;
  loading: boolean;
  notFound: boolean;
  unauthorized: boolean;
}

const EMPTY: FullProfileData = {
  profile: null,
  photos: [],
  languages: [],
  interestCatalog: [],
  theirInterests: [],
  personality: null,
  partnerPrefs: null,
  favorited: false,
  liked: false,
  matched: false,
  blocked: false,
  loading: true,
  notFound: false,
  unauthorized: false,
};

// Antes este archivo tenía su propia caché por perfil, que nunca se
// invalidaba: tras dar like o favorito desde una tarjeta, abrir ese perfil
// podía enseñar el estado anterior. Ahora se usa la caché de apiFetch (60 s,
// vaciada en cada mutación), que ya deduplica peticiones en vuelo.
//
// recordVisit=false pide /full?visit=0: el backend no registra la visita.
// Quick Match lo usa para no contar como visita a quien solo se precarga; y
// registra la visita real con recordProfileVisit cuando muestra la carta.
function fullPath(id: string, recordVisit: boolean) {
  return recordVisit ? `/profiles/${id}/full` : `/profiles/${id}/full?visit=0`;
}

function fromEnvelope(envelope: FullProfileEnvelope): Omit<FullProfileData, 'loading'> {
  return {
    profile: envelope.profile,
    photos: envelope.photos,
    languages: envelope.languages,
    interestCatalog: envelope.interest_catalog,
    theirInterests: envelope.interests,
    personality: envelope.personality,
    partnerPrefs: envelope.partner_preferences,
    favorited: envelope.favorited,
    liked: envelope.liked,
    matched: envelope.matched,
    blocked: envelope.blocked,
    notFound: false,
    unauthorized: false,
  };
}

// Lanza la carga de un perfil sin engancharse a ningún componente. Pensado
// para precargar "el siguiente" en Quick Match. Nunca registra visita.
export function preloadFullProfile(id: string | undefined) {
  if (!id) return;
  const path = fullPath(id, false);
  if (peekApiCache(path)) return;
  apiFetch<FullProfileEnvelope>(path).catch(() => {
    // Precarga silenciosa: si falla, el hook normal la reintentará
    // cuando el usuario realmente navegue a este perfil.
  });
}

// Registra la visita real a un perfil (Quick Match, que carga con visit=0).
// No vacía la caché de GET, para no tirar el perfil que se acaba de precargar.
export function recordProfileVisit(id: string | undefined) {
  if (!id) return;
  apiMutateQuiet<void>(`/visits/${id}`, { method: 'POST' }).catch(() => {
    // Best-effort: una visita no registrada no debe molestar al usuario.
  });
}

export function useFullProfile(id: string | undefined, options?: { recordVisit?: boolean }): FullProfileData {
  const recordVisit = options?.recordVisit ?? true;
  const [data, setData] = useState<FullProfileData>(EMPTY);

  useEffect(() => {
    if (!id) return;
    const path = fullPath(id, recordVisit);

    const cached = peekApiCache<FullProfileEnvelope>(path);
    if (cached) {
      setData({ ...fromEnvelope(cached), loading: false });
      return;
    }

    let isMounted = true;
    setData({ ...EMPTY, loading: true });

    apiFetch<FullProfileEnvelope>(path)
      .then((envelope) => {
        if (isMounted) setData({ ...fromEnvelope(envelope), loading: false });
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        const status = (err as { status?: number })?.status;
        setData({ ...EMPTY, loading: false, notFound: status === 404, unauthorized: status === 401 });
      });

    return () => {
      isMounted = false;
    };
  }, [id, recordVisit]);

  return data;
}
