'use client';

// Antes cada pantalla que pintaba tarjetas de perfil (Home) pedía sus propios
// /likes/sent, /favorites, /likes/received, /favorites/received y mantenía
// su propio Set de ids "dados like" / "recibidos". Se centraliza aquí para
// que Home, Likes, Visits, Favorites y Activity compartan un único caché en
// memoria (y una sola tanda de peticiones) durante la sesión de navegación.

import { useEffect, useSyncExternalStore } from 'react';
import { apiFetch, LikesResponse, FavoritesResponse } from '@/lib/api';

interface InteractionsCache {
  likedIds: Set<string>;
  favoritedIds: Set<string>;
  receivedLikeIds: Set<string>;
  receivedFavIds: Set<string>;
  loaded: boolean;
}

const emptyCache: InteractionsCache = {
  likedIds: new Set(),
  favoritedIds: new Set(),
  receivedLikeIds: new Set(),
  receivedFavIds: new Set(),
  loaded: false,
};

let cache: InteractionsCache | null = null;
let inFlight: Promise<InteractionsCache> | null = null;
const listeners = new Set<() => void>();

function emitChange() {
  listeners.forEach((listener) => listener());
}

function getSnapshot(): InteractionsCache {
  return cache ?? emptyCache;
}

function subscribe(listener: () => void): () => void {
  listeners.add(listener);
  return () => listeners.delete(listener);
}

function loadInteractions(): Promise<InteractionsCache> {
  if (cache?.loaded) return Promise.resolve(cache);
  if (inFlight) return inFlight;

  inFlight = Promise.all([
    apiFetch<LikesResponse>('/likes/sent?page_size=100').catch(() => null),
    apiFetch<FavoritesResponse>('/favorites?page_size=100').catch(() => null),
    apiFetch<LikesResponse>('/likes/received?page_size=100').catch(() => null),
    apiFetch<FavoritesResponse>('/favorites/received?page_size=100').catch(() => null),
  ]).then(([sentLikes, favs, recLikes, recFavs]) => {
    const next: InteractionsCache = {
      likedIds: new Set(sentLikes?.items.map((i) => i.profile_id) ?? []),
      favoritedIds: new Set(favs?.items.map((i) => i.profile_id) ?? []),
      receivedLikeIds: new Set(recLikes?.items.map((i) => i.profile_id) ?? []),
      receivedFavIds: new Set(recFavs?.items.map((i) => i.profile_id) ?? []),
      loaded: true,
    };
    cache = next;
    inFlight = null;
    emitChange();
    return next;
  });

  return inFlight;
}

export function useProfileInteractions() {
  const snapshot = useSyncExternalStore(subscribe, getSnapshot);

  useEffect(() => {
    loadInteractions();
  }, []);

  function toggleLike(id: string, isLiked: boolean) {
    if (!cache) return;
    const next = new Set(cache.likedIds);
    if (isLiked) next.add(id);
    else next.delete(id);
    cache = { ...cache, likedIds: next };
    emitChange();
  }

  function toggleFavorite(id: string, isFav: boolean) {
    if (!cache) return;
    const next = new Set(cache.favoritedIds);
    if (isFav) next.add(id);
    else next.delete(id);
    cache = { ...cache, favoritedIds: next };
    emitChange();
  }

  return {
    likedIds: snapshot.likedIds,
    favoritedIds: snapshot.favoritedIds,
    receivedLikeIds: snapshot.receivedLikeIds,
    receivedFavIds: snapshot.receivedFavIds,
    toggleLike,
    toggleFavorite,
  };
}
