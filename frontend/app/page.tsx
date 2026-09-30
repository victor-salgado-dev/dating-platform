'use client';

import React, { Suspense, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { useI18n } from '@/lib/i18n/context';
import { ProfileCard } from '@/components/ProfileCard';
import { HOME_TABS, useProfileList, type HomeTab } from '@/lib/useProfileList';
import styles from './page.module.css';

// La pestaña y la página viven en la URL (/?tab=online&page=2): las pestañas
// no recargan la app, pero se pueden compartir, funcionan con el botón
// atrás y sobreviven a un F5.

function parseTab(value: string | null | undefined): HomeTab {
  return HOME_TABS.includes(value as HomeTab) ? (value as HomeTab) : 'recommended';
}

function parsePage(value: string | null | undefined): number {
  const n = parseInt(value ?? '1', 10);
  return Number.isFinite(n) && n >= 1 ? n : 1;
}

function buildHref(tab: HomeTab, page: number): string {
  const params = new URLSearchParams();
  if (tab !== 'recommended') params.set('tab', tab);
  if (page > 1) params.set('page', String(page));
  const qs = params.toString();
  return qs ? `/?${qs}` : '/';
}

function HomeContent() {
  const { dictionary } = useI18n();
  const router = useRouter();
  const searchParams = useSearchParams();

  const tab = parseTab(searchParams?.get('tab'));
  const page = parsePage(searchParams?.get('page'));

  const { items: profiles, totalPages, loading, error, reload } = useProfileList(tab, page);

  // Posición del anuncio cuadrado: se sortea al cambiar de pestaña/página,
  // NO en cada fetch, para que no salte mientras cargan los datos.
  const [adIndex, setAdIndex] = useState(-1);
  useEffect(() => {
    setAdIndex(Math.floor(Math.random() * 20) + 2);
  }, [tab, page]);

  // Al paginar, vuelve arriba (no en el primer montaje, para respetar la
  // restauración de scroll del navegador).
  const firstRender = useRef(true);
  useEffect(() => {
    if (firstRender.current) {
      firstRender.current = false;
      return;
    }
    window.scrollTo({ top: 0 });
  }, [page]);

  const goToPage = (next: number) => router.push(buildHref(tab, next), { scroll: false });

  const showInitialLoading = loading && profiles.length === 0;

  return (
    <main className={styles.main}>
      <div className={styles.tabs}>
        {HOME_TABS.map((t) => (
          <Link
            key={t}
            href={buildHref(t, 1)}
            scroll={false}
            className={`${styles.tab} ${t === tab ? styles.tabActive : ''}`}
            aria-current={t === tab ? 'page' : undefined}
          >
            {dictionary.home.tabs[t]}
          </Link>
        ))}
      </div>

      {showInitialLoading && <p className={styles.status}>{dictionary.home.loading}</p>}

      {error && (
        <div className={styles.error}>
          <p>{dictionary.discover.loadError}</p>
          <button type="button" className={styles.pageBtn} onClick={reload}>
            {dictionary.home.retry}
          </button>
        </div>
      )}

      {!error && !loading && profiles.length === 0 && <p className={styles.status}>{dictionary.home.empty}</p>}

      {!error && profiles.length > 0 && (
        <>
          <div className={`${styles.grid} ${loading ? styles.gridDim : ''}`}>
            {profiles.map((profile, index) => (
              <React.Fragment key={profile.profile_id}>
                {index === adIndex && <div className={styles.adSquare}>{dictionary.discover.sponsoredRandom}</div>}
                {index === 14 && <div className={styles.adBanner}>{dictionary.discover.bannerMid}</div>}

                <ProfileCard profile={profile} isPremium={index === 0} />
              </React.Fragment>
            ))}
          </div>

          <div className={styles.adBanner} style={{ marginTop: '3rem' }}>
            {dictionary.home.bannerBottom}
          </div>

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button
                type="button"
                className={styles.pageBtn}
                onClick={() => goToPage(Math.max(1, page - 1))}
                disabled={page <= 1 || loading}
              >
                {dictionary.common.paginationPrev}
              </button>
              <span className={styles.pageInfo}>
                {dictionary.common.paginationPage
                  .replace('{page}', String(page))
                  .replace('{totalPages}', String(totalPages))}
              </span>
              <button
                type="button"
                className={styles.pageBtn}
                onClick={() => goToPage(Math.min(totalPages, page + 1))}
                disabled={page >= totalPages || loading}
              >
                {dictionary.common.paginationNext}
              </button>
            </div>
          )}
        </>
      )}
    </main>
  );
}

// useSearchParams exige un límite de Suspense para poder compilar la página.
export default function HomePage() {
  return (
    <Suspense fallback={null}>
      <HomeContent />
    </Suspense>
  );
}
