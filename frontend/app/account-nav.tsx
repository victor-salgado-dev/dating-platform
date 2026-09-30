'use client';

// Estado de la cuenta del header a nivel de módulo, cargado una sola vez por
// sesión de navegación y compartido entre todas las instancias del hook.
//
// Antes cada carga hacía tres peticiones (auth/me + profiles/me +
// profiles/me/photos) y calculaba el % de perfil completado en el navegador.
// Ahora GET /auth/me devuelve además el avatar y el porcentaje (calculado en
// el servidor), así que es UNA petición.

import { useEffect, useSyncExternalStore } from 'react';
import { apiFetch, type MeResponse } from '@/lib/api';

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

// El color depende solo del porcentaje (misma escala que antes).
function completionFor(percent: number): AccountCompletion {
  let color = '#ef4444'; // rojo
  if (percent >= 50 && percent < 80) color = '#eab308'; // amarillo
  if (percent >= 80) color = '#22c55e'; // verde
  return { percent, color };
}

function loadAccountData(): Promise<AccountData> {
  if (loaded) return Promise.resolve(state);
  if (inFlight) return inFlight;

  inFlight = apiFetch<MeResponse>('/auth/me')
    .then((me) => {
      const next: AccountData = {
        authenticated: true,
        photo: me.photo_url ?? null,
        completion: completionFor(me.profile_completion ?? 0),
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
