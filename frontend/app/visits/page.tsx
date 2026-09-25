'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, VisitsResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

type Tab = 'received' | 'sent' | 'mutual';

export default function VisitsPage() {
  const { dictionary } = useI18n();
  const [tab, setTab] = useState<Tab>('received');
  const [data, setData] = useState<VisitsResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    apiFetch<VisitsResponse>(`/visits/${tab}?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError(dictionary.visits.errorUnauthorized);
        } else {
          setError(dictionary.visits.loadError);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, page]);

  function changeTab(nextTab: Tab) {
    setTab(nextTab);
    setPage(1);
  }

  const visits = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  const emptyTitle: Record<Tab, string> = {
    received: dictionary.visits.emptyReceived,
    sent: dictionary.visits.emptySent,
    mutual: dictionary.visits.emptyMutual,
  };
  const emptySubtitle: Record<Tab, string> = {
    received: dictionary.visits.emptyReceivedSubtitle,
    sent: dictionary.visits.emptySentSubtitle,
    mutual: dictionary.visits.emptyMutualSubtitle,
  };

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.visits.title}</h1>

      <div className={styles.tabs} role="tablist">
        <button
          type="button"
          className={tab === 'received' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('received')}
        >
          {dictionary.visits.tabReceived}
        </button>
        <button
          type="button"
          className={tab === 'sent' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('sent')}
        >
          {dictionary.visits.tabSent}
        </button>
        <button
          type="button"
          className={tab === 'mutual' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('mutual')}
        >
          {dictionary.visits.tabMutual}
        </button>
      </div>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>{dictionary.visits.loading}</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {visits.length === 0 ? (
            <div className={styles.emptyState}>
              <p className={styles.emptyStateTitle}>{emptyTitle[tab]}</p>
              <p className={styles.emptyStateSubtitle}>{emptySubtitle[tab]}</p>
            </div>
          ) : (
            <div className={styles.grid}>
              {visits.map((item) => (
                <Link
                  key={item.profile_id}
                  href={`/profiles/${item.profile_id}`}
                  className={styles.card}
                >
                  <div className={styles.imageContainer}>
                    {item.photo_url ? (
                      <img
                        src={item.photo_url}
                        alt={item.display_name}
                        className={styles.photoImg}
                      />
                    ) : (
                      <div className={styles.photoPlaceholder}>{dictionary.common.noPhoto}</div>
                    )}
                  </div>

                  <div className={styles.cardInfo}>
                    <h3 className={styles.name}>
                      {item.display_name}
                      <span className={styles.age}> · {item.age}</span>
                    </h3>
                    <p className={styles.details}>
                      {[item.region, item.country_code].filter(Boolean).join(', ')}
                    </p>
                  </div>
                </Link>
              ))}
            </div>
          )}

          {totalPages > 1 && (
            <div className={styles.pagination}>
              <button
                type="button"
                className={styles.pageButton}
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                {dictionary.common.paginationPrev}
              </button>
              <span className={styles.pageInfo}>
                {dictionary.visits.paginationInfo
                  .replace('{page}', String(data?.page ?? page))
                  .replace('{totalPages}', String(totalPages))
                  .replace('{total}', String(data?.total ?? 0))}
              </span>
              <button
                type="button"
                className={styles.pageButton}
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
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
