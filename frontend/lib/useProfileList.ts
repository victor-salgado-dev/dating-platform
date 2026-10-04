'use client';

// Carga una página de perfiles para cualquiera de las pestañas de Home. Los
// cuatro endpoints devuelven ya el mismo shape (SearchResponse, con los flags
// de like/favorito incluidos), así que aquí no hay ningún mapeo.
//
// - Sin AbortController a propósito: apiFetch deduplica peticiones GET
//   idénticas compartiendo la misma promesa, y abortarla dejaría colgado a
//   cualquier otro consumidor. En su lugar se descartan las respuestas
//   obsoletas con un flag `cancelled` por efecto.
// - Mientras carga otra página de la MISMA pestaña se conservan los datos
//   anteriores (la grilla no se vacía). Al cambiar de pestaña no se
//   reutilizan, para no mostrar perfiles de otra lista.
// - La pestaña "online" salta la caché de 60 s de apiFetch: es un dato que
//   caduca rápido.

import { useEffect, useState } from 'react';
import { apiFetch, type SearchResponse, type SearchResultItem } from '@/lib/api';

export type HomeTab = 'recommended' | 'popular' | 'online' | 'new';

export const HOME_TABS: HomeTab[] = ['recommended', 'popular', 'online', 'new'];
export const PAGE_SIZE = 24;

const ENDPOINTS: Record<HomeTab, string> = {
  recommended: '/search/recommended',
  popular: '/search/popular',
  online: '/search/online-now',
  new: '/search/new-members',
};

interface Loaded {
  tab: HomeTab;
  items: SearchResultItem[];
  totalPages: number;
}

export function useProfileList(tab: HomeTab, page: number) {
  const [loaded, setLoaded] = useState<Loaded | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const [reloadKey, setReloadKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(false);

    const path = `${ENDPOINTS[tab]}?page=${page}&page_size=${PAGE_SIZE}`;
    const init: RequestInit | undefined = tab === 'online' ? { cache: 'no-store' } : undefined;

    apiFetch<SearchResponse>(path, init)
      .then((res) => {
        if (cancelled) return;
        setLoaded({ tab, items: res.items ?? [], totalPages: res.total_pages ?? 1 });
      })
      .catch(() => {
        if (!cancelled) setError(true);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [tab, page, reloadKey]);

  const visible = loaded && loaded.tab === tab ? loaded : null;

  return {
    items: visible?.items ?? [],
    totalPages: visible?.totalPages ?? 1,
    loading,
    error,
    reload: () => setReloadKey((k) => k + 1),
  };
}
