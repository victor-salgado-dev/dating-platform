'use client';

// Antes: cada cambio de pathname disparaba auth/me + profiles/me +
// profiles/me/photos de nuevo, y ponía authenticated en null de por
// medio (parpadeo del header). Ahora sigue el mismo patrón que
// useProfileInteractions.ts: estado a nivel de módulo, cargado una
// sola vez por sesión de navegación, compartido entre todas las
// instancias del hook (aquí solo hay una, en header-chrome, pero así
// queda listo por si profile/settings quieren reusarlo).

import { useEffect, useSyncExternalStore } from 'react';
import { apiFetch, PublicProfile, ProfilePhoto } from '@/lib/api';

export interface AccountCompletion {
  percent: number;
  color: string;
}

export interface AccountData {
  authenticated: boolean | null;
  photo: string | null;
  completion: AccountCompletion;
}

const DEFAULT_COMPLETION: AccountCompletion = { percent: 0, color: '#ef4444' };

const EMPTY_STATE: AccountData = {
  authenticated: null,
  photo: null,
  completion: DEFAULT_COMPLETION,
};

let state: AccountData = EMPTY_STATE;
let inFlight: Promise<AccountData> | null = null;
let loaded = false;
const listeners = new Set<() => void>();

function emitChange() {
  listeners.forEach((l) => l());
}

function getSnapshot(): AccountData {
  return state;
}

function getServerSnapshot(): AccountData {
  return EMPTY_STATE;
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function computeCompletion(profile: PublicProfile | null, photos: ProfilePhoto[]): AccountCompletion {
  let score = 0;
  const maxScore = 12; // Criterios totales

  if (profile) {
    score += 4; // Datos base (nombre, genero, fecha nac, pais) siempre existen si hay perfil
    if (profile.region) score++;
    if (profile.relationship_goals && profile.relationship_goals.length > 0) score++;
    if (profile.has_children !== null) score++;
    if (profile.bio) score++;
    if (profile.wants_children !== null) score++;
    if (profile.nationality) score++;
  }
  if (photos.length > 0) score += 2; // Extra por foto

  const percent = Math.round((score / maxScore) * 100);
  let color = '#ef4444'; // Rojo por defecto
  if (percent >= 50 && percent < 80) color = '#eab308'; // Amarillo
  if (percent >= 80) color = '#22c55e'; // Verde

  return { percent, color };
}

function loadAccountData(): Promise<AccountData> {
  if (loaded) return Promise.resolve(state);
  if (inFlight) return inFlight;

  inFlight = apiFetch('/auth/me')
    .then(async () => {
      const [profile, photos] = await Promise.all([
        apiFetch<PublicProfile>('/profiles/me').catch(() => null),
        apiFetch<ProfilePhoto[]>('/profiles/me/photos').catch(() => []),
      ]);

      const photo =
        photos.length > 0 ? photos[0].url ?? `/api/v1/profiles/me/photos/${photos[0].id}/file` : null;

      const next: AccountData = {
        authenticated: true,
        photo,
        completion: computeCompletion(profile, photos),
      };
      state = next;
      loaded = true;
      return next;
    })
    .catch(() => {
      const next: AccountData = { ...EMPTY_STATE, authenticated: false };
      state = next;
      loaded = true;
      return next;
    })
    .finally(() => {
      inFlight = null;
      emitChange();
    });

  return inFlight;
}

// Invalida el estado en memoria y fuerza una recarga en el próximo
// render que use el hook. Se llama tras login/logout (evento
// 'auth-change') y también se puede llamar manualmente tras editar el
// perfil o subir/borrar una foto (p.ej. desde profile/edit), para que
// el % de completado y el avatar del header se refresquen sin tener
// que recargar la página entera.
export function invalidateAccountData() {
  loaded = false;
  inFlight = null;
  state = EMPTY_STATE;
  emitChange();
}

export function useAccountData(): AccountData {
  const snapshot = useSyncExternalStore(subscribe, getSnapshot, getServerSnapshot);

  useEffect(() => {
    loadAccountData();
  }, []);

  useEffect(() => {
    function handleAuthChange() {
      invalidateAccountData();
      // Tras logout no hace falta recargar (authenticated: false ya es
      // el estado correcto y no hay sesión para pedir /auth/me). Tras
      // login sí: el próximo montaje de header-chrome (nueva página)
      // volverá a llamar a loadAccountData() y esta vez authenticated
      // será true.
    }
    window.addEventListener('auth-change', handleAuthChange);
    return () => window.removeEventListener('auth-change', handleAuthChange);
  }, []);

  return snapshot;
}

export async function logoutAndRedirect(router: { push: (href: string) => void }) {
  try {
    await apiFetch<void>('/auth/logout', { method: 'POST' });
  } catch {
    // Si falla, igualmente cerramos sesión en el cliente
  }
  window.dispatchEvent(new Event('auth-change'));
  router.push('/login');
}
