'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, FavoritesResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

interface PhotoItem {
  id: string;
  url: string;
  position: number;
}

function ProfilePhoto({ profileId, name }: { profileId: string; name: string }) {
  const { dictionary } = useI18n();
  const [photoUrl, setPhotoUrl] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<PhotoItem[]>(`/profiles/${profileId}/photos`)
      .then((photos) => {
        if (photos && photos.length > 0) {
          const firstPhoto = photos[0];
          setPhotoUrl(firstPhoto.url ?? `/api/v1/profiles/${profileId}/photos/${firstPhoto.id}/file`);
        }
      })
      .catch(() => setPhotoUrl(null));
  }, [profileId]);

  if (!photoUrl) {
    return <div className={styles.photoPlaceholder}>{dictionary.common.noPhoto}</div>;
  }

  return <img src={photoUrl} alt={name} className={styles.photoImg} />;
}

export default function FavoritesPage() {
  const { dictionary } = useI18n();
  const [tab, setTab] = useState<'sent' | 'received'>('sent');
  const [data, setData] = useState<FavoritesResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    apiFetch<FavoritesResponse>(`/favorites/${tab}?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError(dictionary.favorites.errorUnauthorized);
        } else {
          setError(dictionary.favorites.loadError);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, page]);

  function changeTab(nextTab: 'sent' | 'received') {
    setTab(nextTab);
    setPage(1);
  }

  const favorites = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.favorites.title}</h1>

      <div className={styles.tabs} role="tablist">
        <button
          type="button"
          className={tab === 'sent' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('sent')}
        >
          {dictionary.favorites.tabSent}
        </button>
        <button
          type="button"
          className={tab === 'received' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('received')}
        >
          {dictionary.favorites.tabReceived}
        </button>
      </div>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>{dictionary.favorites.loading}</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {favorites.length === 0 ? (
            <div className={styles.emptyState}>
              <p className={styles.emptyStateTitle}>
                {tab === 'sent'
                  ? dictionary.favorites.emptySent
                  : dictionary.favorites.emptyReceived}
              </p>
              <p className={styles.emptyStateSubtitle}>
                {tab === 'sent'
                  ? dictionary.favorites.emptySentSubtitle
                  : dictionary.favorites.emptyReceivedSubtitle}
              </p>
            </div>
          ) : (
            <div className={styles.grid}>
              {favorites.map((item) => (
                <Link
                  key={item.profile_id}
                  href={`/profiles/${item.profile_id}`}
                  className={styles.card}
                >
                  <div className={styles.imageContainer}>
                    {item.has_photo ? (
                      <ProfilePhoto profileId={item.profile_id} name={item.display_name} />
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
                {dictionary.favorites.paginationInfo
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
