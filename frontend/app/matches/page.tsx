'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';

import { apiFetch, SearchResponse, SearchResultItem, LikeResult } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { useFullProfile, preloadFullProfile, recordProfileVisit } from '@/lib/useFullProfile';
import { loadSavedFilters } from '@/lib/searchFilters';
import { FullProfileSections } from '@/components/FullProfileSections';
import styles from './page.module.css';

function shuffle<T>(arr: T[]): T[] {
  const copy = [...arr];
  for (let i = copy.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [copy[i], copy[j]] = [copy[j], copy[i]];
  }
  return copy;
}

export default function QuickMatchPage() {
  const { dictionary } = useI18n();
  // like/favorito de las cartas de esta sesión: se parte de los flags que
  // trae cada ficha y se sobreescribe al pulsar los botones.
  const [overrides, setOverrides] = useState<Record<string, { liked?: boolean; favorited?: boolean }>>({});
  const setFlag = useCallback((id: string, patch: { liked?: boolean; favorited?: boolean }) => {
    setOverrides((prev) => ({ ...prev, [id]: { ...prev[id], ...patch } }));
  }, []);

  const [queue, setQueue] = useState<SearchResultItem[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [fetchPage, setFetchPage] = useState(1);
  const [queueLoading, setQueueLoading] = useState(true);
  const [exhausted, setExhausted] = useState(false);
  const [activePhoto, setActivePhoto] = useState(0);
  const [likeBusy, setLikeBusy] = useState(false);
  const [favoriteBusy, setFavoriteBusy] = useState(false);
  const [showMatchNotice, setShowMatchNotice] = useState<string | null>(null);

  const seenIds = useRef<Set<string>>(new Set());

  // Filtros guardados desde Búsqueda. Lo que no esté en ellos lo completa el
  // backend con el género que buscas y la edad de tus Partner preferences.
  const fetchNextBatch = useCallback(
    async (pageToFetch: number) => {
      const savedFilters = await loadSavedFilters();
      const params = new URLSearchParams(savedFilters);
      params.set('page', String(pageToFetch));
      params.set('page_size', '50');
      params.set('include_total', 'false');

      try {
        const res = await apiFetch<SearchResponse>(`/search/profiles?${params.toString()}`);
        // Los flags vienen en cada ficha: ya no hace falta descargar antes
        // las listas de likes y favoritos para no repetir perfiles.
        const fresh = res.items.filter(
          (item) => !item.liked && !item.favorited && !seenIds.current.has(item.profile_id)
        );
        fresh.forEach((item) => seenIds.current.add(item.profile_id));

        setQueue((prev) => [...prev, ...shuffle(fresh)]);
        if (res.items.length === 0 || res.has_more === false || (res.has_more === undefined && pageToFetch >= res.total_pages)) {
          setExhausted(true);
        }
      } catch {
        setExhausted(true);
      } finally {
        setQueueLoading(false);
      }
    },
    []
  );

  // Primera carga
  const started = useRef(false);
  useEffect(() => {
    if (started.current) return;
    started.current = true;
    fetchNextBatch(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Pide más cuando quedan pocas cartas y todavía hay páginas
  useEffect(() => {
    if (exhausted || queueLoading) return;
    if (currentIndex >= queue.length - 5) {
      const next = fetchPage + 1;
      setFetchPage(next);
      setQueueLoading(true);
      fetchNextBatch(next);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentIndex, queue.length]);

  const current = queue[currentIndex];
  // visit=0: la visita real se registra abajo, cuando la carta se muestra; si
  // no, precargar la siguiente contaría como visita a alguien aún no visto.
  const full = useFullProfile(current?.profile_id, { recordVisit: false });

  // Mientras se mira la carta actual, se precarga la siguiente en segundo
  // plano: cuando el usuario le dé a "Siguiente" ya estará en caché y
  // useFullProfile la devuelve al instante, sin spinner.
  useEffect(() => {
    const next = queue[currentIndex + 1];
    if (next) preloadFullProfile(next.profile_id);
  }, [currentIndex, queue]);

  useEffect(() => {
    setActivePhoto(0);
    setShowMatchNotice(null);
  }, [current?.profile_id]);

  // Registra la visita cuando la carta ya está en pantalla.
  const visitedShown = full.profile?.id;
  useEffect(() => {
    if (visitedShown) recordProfileVisit(visitedShown);
  }, [visitedShown]);

  const liked = current ? overrides[current.profile_id]?.liked ?? current.liked : false;
  const favorited = current ? overrides[current.profile_id]?.favorited ?? current.favorited : false;

  async function handleLike() {
    if (!current || likeBusy) return;
    setLikeBusy(true);
    const next = !liked;
    setFlag(current.profile_id, { liked: next });
    try {
      // POST /likes/{id} devuelve { matched }: ya no hace falta pedir /matches/{id} aparte.
      const res = await apiFetch<LikeResult | undefined>(`/likes/${current.profile_id}`, { method: next ? 'POST' : 'DELETE' });
      if (next && res?.matched) setShowMatchNotice(current.display_name);
    } catch {
      setFlag(current.profile_id, { liked: !next });
    } finally {
      setLikeBusy(false);
    }
  }

  async function handleFavorite() {
    if (!current || favoriteBusy) return;
    setFavoriteBusy(true);
    const next = !favorited;
    setFlag(current.profile_id, { favorited: next });
    try {
      await apiFetch(`/favorites/${current.profile_id}`, { method: next ? 'POST' : 'DELETE' });
    } catch {
      setFlag(current.profile_id, { favorited: !next });
    } finally {
      setFavoriteBusy(false);
    }
  }

  function goPrev() {
    setCurrentIndex((i) => Math.max(0, i - 1));
  }
  function goNext() {
    setCurrentIndex((i) => Math.min(queue.length - 1, i + 1));
  }

  const noMoreCards = !queueLoading && currentIndex >= queue.length;

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.quickMatch.title}</h1>
      <p className={styles.hint}>{dictionary.quickMatch.filterHint}</p>

      {queueLoading && queue.length === 0 && <div className={styles.loadingState}>{dictionary.quickMatch.loading}</div>}

      {noMoreCards && (
        <div className={styles.emptyState}>
          <p>{dictionary.quickMatch.noMoreProfiles}</p>
          <Link href="/search" style={{ color: '#9f1239', fontWeight: 600 }}>
            {dictionary.quickMatch.adjustFilters}
          </Link>
        </div>
      )}

      {current && full.profile && (
        <>
          {showMatchNotice && (
            <div className={styles.matchNotice} role="status">
              <span>{dictionary.profilePublic.matchNotice.replace('{name}', showMatchNotice)}</span>
              <button type="button" onClick={() => setShowMatchNotice(null)}>
                {dictionary.common.close}
              </button>
            </div>
          )}

          <div className={styles.card}>
            {full.photos.length > 0 ? (
              <img src={full.photos[activePhoto]?.url ?? full.photos[0].url} alt="" className={styles.photoMain} />
            ) : (
              <div className={styles.photoPlaceholder}>{dictionary.common.noPhoto}</div>
            )}

            {full.photos.length > 1 && (
              <div className={styles.thumbRow}>
                {full.photos.map((p, idx) => (
                  <img
                    key={p.id}
                    src={p.url}
                    alt=""
                    loading="lazy"
                    className={idx === activePhoto ? `${styles.thumb} ${styles.thumbActive}` : styles.thumb}
                    onClick={() => setActivePhoto(idx)}
                  />
                ))}
              </div>
            )}

            <div className={styles.body}>
              <div className={styles.nameRow}>
                <h1>
                  {full.profile.display_name}, {full.profile.age}
                </h1>
              </div>
              <p className={styles.location}>
                {[full.profile.region, full.profile.country_code].filter(Boolean).join(', ')}
              </p>

              <Link href={`/profiles/${current.profile_id}`} className={styles.viewFullLink}>
                {dictionary.quickMatch.viewFullProfile} →
              </Link>

              <FullProfileSections
                profile={full.profile}
                languages={full.languages}
                interestCatalog={full.interestCatalog}
                theirInterests={full.theirInterests}
                personality={full.personality}
                partnerPrefs={full.partnerPrefs}
              />
            </div>
          </div>

          <div className={styles.actionBar}>
            <button type="button" className={styles.navBtn} onClick={goPrev} disabled={currentIndex === 0} aria-label={dictionary.quickMatch.prev}>
              ‹
            </button>
            <button
              type="button"
              className={favorited ? `${styles.bigActionBtn} ${styles.favoriteBtnActive}` : styles.bigActionBtn}
              onClick={handleFavorite}
              disabled={favoriteBusy}
              aria-label={favorited ? dictionary.quickMatch.unfavorite : dictionary.quickMatch.favorite}
            >
              ★
            </button>
            <button
              type="button"
              className={liked ? `${styles.bigActionBtn} ${styles.likeBtnActive}` : styles.bigActionBtn}
              onClick={handleLike}
              disabled={likeBusy}
              aria-label={liked ? dictionary.quickMatch.unlike : dictionary.quickMatch.like}
            >
              ♥
            </button>
            <button
              type="button"
              className={styles.navBtn}
              onClick={goNext}
              disabled={currentIndex >= queue.length - 1 && exhausted}
              aria-label={dictionary.quickMatch.next}
            >
              ›
            </button>
          </div>
        </>
      )}
    </main>
  );
}
