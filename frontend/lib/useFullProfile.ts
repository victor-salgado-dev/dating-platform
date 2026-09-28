'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
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

// Caché en memoria por perfil visitado. Vive durante la sesión de
// navegación (igual que useProfileInteractions), así que volver a un
// perfil ya visto -p.ej. "Anterior" en Quick Match- no vuelve a lanzar
// la petición.
const profileDataCache = new Map<string, Omit<FullProfileData, 'loading'>>();

// Promesas en vuelo por perfil. Evita que dos componentes pidan el mismo
// perfil a la vez y lancen una segunda tanda de peticiones.
const profileDataPromises = new Map<string, Promise<Omit<FullProfileData, 'loading'>>>();

function fetchFullProfile(id: string): Promise<Omit<FullProfileData, 'loading'>> {
  const existing = profileDataPromises.get(id);
  if (existing) return existing;

  const promise = apiFetch<FullProfileEnvelope>(`/profiles/${id}/full`)
    .then((envelope) => {
      const resolved: Omit<FullProfileData, 'loading'> = {
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
      profileDataCache.set(id, resolved);
      return resolved;
    })
    .catch((err) => {
      // Rechazamos la promesa para que el llamador pueda manejar el error.
      // El `finally` de abajo se encargará de limpiar el mapa de promesas.
      throw err;
    })
    .finally(() => {
      profileDataPromises.delete(id);
    });

  profileDataPromises.set(id, promise);
  return promise;
}

// Lanza la carga de un perfil sin engancharse a ningún componente ni
// devolver estado. Pensado para precargar "el siguiente" en Quick Match
// mientras el usuario todavía está mirando el actual. Si ya está en
// caché o ya se está pidiendo, no hace nada.
export function preloadFullProfile(id: string | undefined) {
  if (!id || profileDataCache.has(id)) return;
  fetchFullProfile(id).catch(() => {
    // Precarga silenciosa: si falla, el hook normal la reintentará
    // cuando el usuario realmente navegue a este perfil.
  });
}

export function useFullProfile(id: string | undefined): FullProfileData {
  const [data, setData] = useState<FullProfileData>(EMPTY);

  useEffect(() => {
    if (!id) return;

    const cached = profileDataCache.get(id);
    if (cached) {
      setData({ ...cached, loading: false });
      return;
    }

    let isMounted = true;
    setData({ ...EMPTY, loading: true });

    fetchFullProfile(id)
      .then((resolved) => {
        if (isMounted) setData({ ...resolved, loading: false });
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        const status = (err as { status?: number })?.status;
        setData({ ...EMPTY, loading: false, notFound: status === 404, unauthorized: status === 401 });
      });

    return () => {
      isMounted = false;
    };
  }, [id]);

  return data;
}
