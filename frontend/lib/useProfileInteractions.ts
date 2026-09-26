'use client';

// Antes cada pantalla que pintaba tarjetas de perfil (Home) pedía sus propios
// /likes/sent, /favorites, /likes/received, /favorites/received y mantenía
// su propio Set de ids "dados like" / "recibidos". Se centraliza aquí para
// que Home, Likes, Visits, Favorites y Activity compartan un único caché en
// memoria (y una sola tanda de peticiones) durante la sesión de navegación.

import { useEffect, useRef, useState } from 'react';
import { apiFetch, LikesResponse, FavoritesResponse } from '@/lib/api';

interface InteractionsCache {
  likedIds: Set<string>;
  favoritedIds: Set<string>;
  receivedLikeIds: Set<string>;
  receivedFavIds: Set<string>;
  loaded: boolean;
}

let cache: InteractionsCache | null = null;

export function useProfileInteractions() {
  const [likedIds, setLikedIds] = useState<Set<string>>(cache?.likedIds ?? new Set());
  const [favoritedIds, setFavoritedIds] = useState<Set<string>>(cache?.favoritedIds ?? new Set());
  const [receivedLikeIds, setReceivedLikeIds] = useState<Set<string>>(cache?.receivedLikeIds ?? new Set());
  const [receivedFavIds, setReceivedFavIds] = useState<Set<string>>(cache?.receivedFavIds ?? new Set());

  const fetched = useRef(false);

  useEffect(() => {
    if (cache?.loaded || fetched.current) return;
    fetched.current = true;

    Promise.all([
      apiFetch<LikesResponse>('/likes/sent?page_size=100').catch(() => null),
      apiFetch<FavoritesResponse>('/favorites?page_size=100').catch(() => null),
      apiFetch<LikesResponse>('/likes/received?page_size=100').catch(() => null),
      apiFetch<FavoritesResponse>('/favorites/received?page_size=100').catch(() => null),
    ]).then(([sentLikes, favs, recLikes, recFavs]) => {
      const nextLiked = new Set(sentLikes?.items.map((i) => i.profile_id) ?? []);
      const nextFavs = new Set(favs?.items.map((i) => i.profile_id) ?? []);
      const nextRecLikes = new Set(recLikes?.items.map((i) => i.profile_id) ?? []);
      const nextRecFavs = new Set(recFavs?.items.map((i) => i.profile_id) ?? []);

      cache = { likedIds: nextLiked, favoritedIds: nextFavs, receivedLikeIds: nextRecLikes, receivedFavIds: nextRecFavs, loaded: true };

      setLikedIds(nextLiked);
      setFavoritedIds(nextFavs);
      setReceivedLikeIds(nextRecLikes);
      setReceivedFavIds(nextRecFavs);
    });
  }, []);

  function toggleLike(id: string, isLiked: boolean) {
    setLikedIds((prev) => {
      const next = new Set(prev);
      if (isLiked) next.add(id);
      else next.delete(id);
      if (cache) cache.likedIds = next;
      return next;
    });
  }

  function toggleFavorite(id: string, isFav: boolean) {
    setFavoritedIds((prev) => {
      const next = new Set(prev);
      if (isFav) next.add(id);
      else next.delete(id);
      if (cache) cache.favoritedIds = next;
      return next;
    });
  }

  return { likedIds, favoritedIds, receivedLikeIds, receivedFavIds, toggleLike, toggleFavorite };
}
