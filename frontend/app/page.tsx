'use client';

import React, { useEffect, useState, useRef } from 'react';
import { apiFetch, SearchResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { ProfileCard } from '@/components/ProfileCard';
import { useProfileInteractions } from '@/lib/useProfileInteractions';
import styles from './page.module.css';

// Pestañas de Home. Solo "recommended" tiene datos reales hoy: el resto
// necesita sort_by/order en /search/profiles (backend), que todavía no
// existe -> se muestran como "próximamente" en vez de fingir que funcionan.
type HomeTab = 'recommended' | 'popular' | 'online' | 'new';

const COMING_SOON_TABS: HomeTab[] = ['popular', 'online', 'new'];

// -----------------------------------------------------------------------------
// Página Principal
// -----------------------------------------------------------------------------
export default function HomePage() {
  const { dictionary } = useI18n();
  const [activeTab, setActiveTab] = useState<HomeTab>('recommended');
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  const { likedIds, favoritedIds, receivedLikeIds, receivedFavIds, toggleLike, toggleFavorite } =
    useProfileInteractions();

  const [randomAdIndex, setRandomAdIndex] = useState(-1);

  // Guarda para evitar que React StrictMode dispare la búsqueda dos veces al montar
  const lastFetchedPage = useRef<number | null>(null);

  // Cargar perfiles (con guarda anti-duplicado). Solo la pestaña
  // "recommended" pide datos reales; el resto son placeholders hasta que
  // el backend soporte ordenar por popularidad / online / fecha.
  useEffect(() => {
    if (activeTab !== 'recommended') return;
    if (lastFetchedPage.current === page) return;
    lastFetchedPage.current = page;

    setRandomAdIndex(Math.floor(Math.random() * 20) + 2);
    setLoading(true);

    apiFetch<SearchResponse>(`/search/profiles?page=${page}&page_size=24`)
      .then(setData)
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  }, [page, activeTab]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      <div
        style={{
          display: 'flex',
          gap: '0.5rem',
          padding: '0 1rem',
          marginBottom: '1.5rem',
          overflowX: 'auto',
        }}
      >
        {(['recommended', 'popular', 'online', 'new'] as HomeTab[]).map((tab) => {
          const isComingSoon = COMING_SOON_TABS.includes(tab);
          const isActive = activeTab === tab;
          const label = dictionary.home.tabs[tab];
          return (
            <button
              key={tab}
              type="button"
              disabled={isComingSoon}
              onClick={() => setActiveTab(tab)}
              title={isComingSoon ? dictionary.home.tabComingSoonHint : undefined}
              style={{
                padding: '0.5rem 1rem',
                borderRadius: '999px',
                border: isActive ? '1px solid #111' : '1px solid #e5e7eb',
                background: isActive ? '#111' : '#fff',
                color: isComingSoon ? '#9ca3af' : isActive ? '#fff' : '#374151',
                fontWeight: 600,
                fontSize: '0.9rem',
                whiteSpace: 'nowrap',
                cursor: isComingSoon ? 'not-allowed' : 'pointer',
                opacity: isComingSoon ? 0.6 : 1,
                display: 'flex',
                alignItems: 'center',
                gap: '0.35rem',
              }}
            >
              {label}
              {isComingSoon && (
                <span style={{ fontSize: '0.7rem', fontWeight: 500 }}>· {dictionary.home.tabComingSoonBadge}</span>
              )}
            </button>
          );
        })}
      </div>

      {activeTab !== 'recommended' ? (
        <div style={{ textAlign: 'center', padding: '4rem 1rem', color: '#666' }}>
          <p style={{ fontSize: '1.1rem', fontWeight: 600, marginBottom: '0.5rem' }}>
            {dictionary.home.tabComingSoonBadge}
          </p>
          <p>{dictionary.home.tabComingSoonHint}</p>
        </div>
      ) : (
        <>
          {loading && (
            <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>
              {dictionary.home.loading}
            </p>
          )}

          <div className={styles.grid}>
            {profiles.map((profile, index) => {
              const isPremium = index === 0;
              const showAdSquare = index === randomAdIndex;
              const showBannerHorizontal = index === 14;

              const isLiked = likedIds.has(profile.profile_id);
              const isFavorited = favoritedIds.has(profile.profile_id);
              const gotLike = receivedLikeIds.has(profile.profile_id);
              const gotFavorite = receivedFavIds.has(profile.profile_id);

              return (
                <React.Fragment key={profile.profile_id}>
                  {showAdSquare && <div className={styles.adSquare}>{dictionary.discover.sponsoredRandom}</div>}
                  {showBannerHorizontal && <div className={styles.adBanner}>{dictionary.discover.bannerMid}</div>}

                  <ProfileCard
                    profile={profile}
                    isPremium={isPremium}
                    initialLiked={isLiked}
                    initialFavorited={isFavorited}
                    receivedLike={gotLike}
                    receivedFavorite={gotFavorite}
                    onToggleLike={toggleLike}
                    onToggleFavorite={toggleFavorite}
                  />
                </React.Fragment>
              );
            })}
          </div>

          {!loading && profiles.length > 0 && (
            <>
              <div className={styles.adBanner} style={{ marginTop: '3rem' }}>
                {dictionary.home.bannerBottom}
              </div>

              {totalPages > 1 && (
                <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '1rem', margin: '2rem 0' }}>
                  <button
                    type="button"
                    onClick={() => setPage((p) => Math.max(1, p - 1))}
                    disabled={page <= 1}
                    style={{ padding: '0.75rem 1.5rem', cursor: page <= 1 ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
                  >
                    {dictionary.common.paginationPrev}
                  </button>
                  <span style={{ fontWeight: 600, color: '#444' }}>
                    {dictionary.common.paginationPage
                      .replace('{page}', String(page))
                      .replace('{totalPages}', String(totalPages))}
                  </span>
                  <button
                    type="button"
                    onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
                    disabled={page >= totalPages}
                    style={{ padding: '0.75rem 1.5rem', cursor: page >= totalPages ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
                  >
                    {dictionary.common.paginationNext}
                  </button>
                </div>
              )}
            </>
          )}
        </>
      )}
    </main>
  );
}
