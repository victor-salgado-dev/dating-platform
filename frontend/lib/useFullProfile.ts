'use client';

import { useEffect, useState } from 'react';
import { apiFetch } from '@/lib/api';
import type {
  PublicProfile,
  ProfilePhoto,
  ProfileLanguage,
  InterestDefinition,
  ProfileInterest,
  PersonalityResponse,
  PartnerPreferences,
} from '@/lib/api';

// Envelope devuelto por GET /profiles/{id}/full (backend: Handler.GetPublicFull
// en internal/profiles/handler.go). Sustituye a las 6 peticiones
// separadas que hacía este archivo antes (profile, photos, languages,
// interests, personality, partner-preferences): ahora es una sola
// llamada HTTP. El catálogo de intereses NO viene en este envelope a
// propósito — no depende del perfil visitado, así que se sigue
// pidiendo y cacheando aparte más abajo (interestCatalogCache).
type FullProfileEnvelope = {
  profile: PublicProfile;
  photos: ProfilePhoto[];
  languages: ProfileLanguage[];
  interests: ProfileInterest[];
  personality: PersonalityResponse;
  partner_preferences: PartnerPreferences;
};

// Catálogo de intereses: no depende del id visitado, se pide una sola vez
// y se comparte entre tarjetas (evita re-pedirlo en cada swipe de Quick Match).
let interestCatalogCache: InterestDefinition[] | null = null;

// Caché en memoria por perfil visitado. Vive durante la sesión de
// navegación (igual que useProfileInteractions), así que volver a un
// perfil ya visto -p.ej. "Anterior" en Quick Match- no vuelve a lanzar
// la petición.
const profileDataCache = new Map<string, Omit<FullProfileData, 'loading'>>();

// Promesas en vuelo por perfil. Evita que dos componentes pidan el mismo
// perfil a la vez y lancen una segunda tanda de peticiones.
const profileDataPromises = new Map<string, Promise<Omit<FullProfileData, 'loading'>>>();

export interface FullProfileData {
  profile: PublicProfile | null;
  photos: ProfilePhoto[];
  languages: ProfileLanguage[];
  interestCatalog: InterestDefinition[];
  theirInterests: ProfileInterest[];
  personality: PersonalityResponse | null;
  partnerPrefs: PartnerPreferences | null;
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
  loading: true,
  notFound: false,
  unauthorized: false,
};

function fetchFullProfile(id: string): Promise<Omit<FullProfileData, 'loading'>> {
  // Si ya hay una petición en curso para este id, devolvemos esa misma
  // promesa. Así, si dos componentes se montan al mismo tiempo, solo se
  // hará una única carga.
  const existing = profileDataPromises.get(id);
  if (existing) return existing;

  const catalogPromise = interestCatalogCache
    ? Promise.resolve(interestCatalogCache)
    : apiFetch<InterestDefinition[]>('/catalog/interests')
        .then((c) => {
          interestCatalogCache = c;
          return c;
        })
        .catch(() => [] as InterestDefinition[]);

  const promise = apiFetch<FullProfileEnvelope>(`/profiles/${id}/full`)
    .then(async (envelope) => {
      const interestCatalog = await catalogPromise;

      const resolved: Omit<FullProfileData, 'loading'> = {
        profile: envelope.profile,
        photos: envelope.photos,
        languages: envelope.languages,
        interestCatalog,
        theirInterests: envelope.interests,
        personality: envelope.personality,
        partnerPrefs: envelope.partner_preferences,
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
