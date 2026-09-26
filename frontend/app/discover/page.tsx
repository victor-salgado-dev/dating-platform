'use client';

import React, { useEffect, useState, useRef, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';

import { apiFetch, ApiError, SearchResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { ProfileCard } from '@/components/ProfileCard';
import { useProfileInteractions } from '@/lib/useProfileInteractions';
import styles from './page.module.css';

// -----------------------------------------------------------------------------
// Componente interno con lógica de filtros y peticiones
// -----------------------------------------------------------------------------
function DiscoverContent() {
  const searchParams = useSearchParams();
  const searchString = searchParams.toString();
  const { dictionary } = useI18n();

  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [lastSearch, setLastSearch] = useState(searchString);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [randomAdIndex, setRandomAdIndex] = useState(-1);

  const { likedIds, favoritedIds, receivedLikeIds, receivedFavIds, toggleLike, toggleFavorite } =
    useProfileInteractions();

  // Guarda contra React Strict Mode / peticiones duplicadas
  const lastSearchKey = useRef<string | null>(null);

  if (searchString !== lastSearch) {
    setLastSearch(searchString);
    setPage(1);
  }

  // Guarda el filtro aplicado para que Quick Match pueda reutilizarlo sin
  // tener que aplicarlo dos veces. Es un dato por-navegador (no viaja entre
  // dispositivos); si eso hace falta más adelante, necesitará un endpoint.
  useEffect(() => {
    if (!searchString) return;
    try {
      window.localStorage.setItem('discoveryFilters', searchString);
    } catch {
      // localStorage no disponible (modo privado, SSR, etc.) — no es crítico
    }
  }, [searchString]);

  // Cargar perfiles según los filtros (de la URL, vienen de /search) y la página
  useEffect(() => {
    const currentKey = `${page}-${searchString}`;
    if (lastSearchKey.current === currentKey) return;
    lastSearchKey.current = currentKey;

    let isMounted = true;
    setLoading(true);
    setError(null);
    setRandomAdIndex(Math.floor(Math.random() * 20) + 2);

    const params = new URLSearchParams(searchString);
    params.set('page', page.toString());
    params.set('page_size', '24');

    apiFetch<SearchResponse>(`/search/profiles?${params.toString()}`)
      .then((res) => {
        if (isMounted) setData(res);
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        if (err instanceof ApiError && err.status === 401) {
          setError(dictionary.discover.errorUnauthorized);
        } else if (err instanceof ApiError && err.status === 429) {
          setError(dictionary.errors.rateLimited);
        } else {
          setError(dictionary.discover.loadError);
        }
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });

    return () => {
      isMounted = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page, searchString]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <>
      {loading && (
        <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>{dictionary.discover.loading}</p>
      )}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {profiles.length === 0 ? (
            <div className={styles.emptyState}>
              <p style={{ fontSize: '1.2rem', marginBottom: '1rem' }}>{dictionary.discover.empty}</p>
              <Link
                href="/search"
                style={{
                  color: 'var(--brand-primary)',
                  textDecoration: 'none',
                  fontWeight: 'bold',
                  padding: '0.5rem 1rem',
                  border: '1px solid var(--brand-primary)',
                  borderRadius: '4px',
                }}
              >
                {dictionary.discover.changeFilters}
              </Link>
            </div>
          ) : (
            <div className={styles.grid}>
              {profiles.map((profile, index) => {
                const isPremium = index === 0;
                const showAdSquare = index === randomAdIndex;
                const showBannerHorizontal = index === 14;

                return (
                  <React.Fragment key={profile.profile_id}>
                    {showAdSquare && <div className={styles.adSquare}>{dictionary.discover.sponsoredRandom}</div>}
                    {showBannerHorizontal && <div className={styles.adBanner}>{dictionary.discover.bannerMid}</div>}

                    <ProfileCard
                      profile={profile}
                      isPremium={isPremium}
                      initialLiked={likedIds.has(profile.profile_id)}
                      initialFavorited={favoritedIds.has(profile.profile_id)}
                      receivedLike={receivedLikeIds.has(profile.profile_id)}
                      receivedFavorite={receivedFavIds.has(profile.profile_id)}
                      onToggleLike={toggleLike}
                      onToggleFavorite={toggleFavorite}
                    />
                  </React.Fragment>
                );
              })}
            </div>
          )}

          {totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '1rem', margin: '3rem 0' }}>
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                style={{ padding: '0.75rem 1.5rem', cursor: page <= 1 ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
              >
                {dictionary.common.paginationPrev}
              </button>
              <span style={{ fontWeight: 600, color: '#444' }}>
                {dictionary.common.paginationPage.replace('{page}', String(data?.page ?? page)).replace('{totalPages}', String(totalPages))}
              </span>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                style={{ padding: '0.75rem 1.5rem', cursor: page >= totalPages ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
              >
                {dictionary.common.paginationNext}
              </button>
            </div>
          )}
        </>
      )}
    </>
  );
}

// -----------------------------------------------------------------------------
// Export principal de la página Discover
// -----------------------------------------------------------------------------
export default function DiscoverPage() {
  const { dictionary } = useI18n();
  return (
    <main className={styles.main}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '1.8rem', margin: 0, color: '#111827' }}>{dictionary.discover.title}</h1>
        <Link
          href="/search"
          style={{
            textDecoration: 'none',
            background: 'var(--brand-primary)',
            color: 'white',
            padding: '0.6rem 1.2rem',
            borderRadius: '8px',
            fontWeight: 'bold',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          {dictionary.discover.filtersLink}
        </Link>
      </div>

      <Suspense fallback={<p style={{ textAlign: 'center' }}>{dictionary.common.loading}</p>}>
        <DiscoverContent />
      </Suspense>
    </main>
  );
}
