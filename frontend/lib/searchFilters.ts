// Filtros de la búsqueda avanzada guardados en el navegador.
//
// Los escribe SOLO la pantalla Búsqueda (al enviar y al limpiar) y los leen
// Búsqueda (para rellenar el formulario), Discover y Quick Match.
// Se guardan como la query string que ya entiende /search/profiles
// (p. ej. "min_age=25&gender=female"). Lo que no esté en ella no se filtra:
// el backend usa entonces el género buscado y la edad de Partner del perfil.
//
// La clave lleva el id del usuario, para que dos cuentas en el mismo navegador
// no hereden los filtros de la otra. Es un dato por-navegador; si hiciera falta
// entre dispositivos, necesitaría un endpoint.

import { apiFetch, type MeResponse } from '@/lib/api';

const KEY_PREFIX = 'searchFilters:';
const LEGACY_KEY = 'discoveryFilters';

async function currentKey(): Promise<string | null> {
  try {
    // apiFetch cachea los GET 60 s y deduplica las peticiones en vuelo.
    const me = await apiFetch<MeResponse>('/auth/me');
    return KEY_PREFIX + me.id;
  } catch {
    return null;
  }
}

// Devuelve la query string guardada ('' si no hay ninguna).
export async function loadSavedFilters(): Promise<string> {
  const key = await currentKey();
  if (!key) return '';
  try {
    return window.localStorage.getItem(key) ?? '';
  } catch {
    return '';
  }
}

// Guarda la query string. Vacía = borra los filtros guardados.
export async function saveFilters(query: string): Promise<void> {
  const key = await currentKey();
  if (!key) return;
  try {
    if (query) window.localStorage.setItem(key, query);
    else window.localStorage.removeItem(key);
    window.localStorage.removeItem(LEGACY_KEY);
  } catch {
    // localStorage no disponible (modo privado, etc.): no es crítico.
  }
}

export function clearSavedFilters(): Promise<void> {
  return saveFilters('');
}
