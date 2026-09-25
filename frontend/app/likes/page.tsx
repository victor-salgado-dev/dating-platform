'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import {
  apiFetch,
  ApiError,
  LikesResponse,
  MatchesResponse,
  LikeItem,
  MatchItem,
} from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

interface PhotoItem {
  id: string;
  url: string;
  position: number;
}

type Tab = 'sent' | 'received' | 'mutual';

// Forma común de tarjeta: /likes/{sent,received} devuelven LikeItem (con
// liked_at) y /matches devuelve MatchItem (con matched_at y un
// conversation_id opcional). Normalizamos a esta forma para poder pintar
// la misma tarjeta en las tres tabs.
interface CardItem {
  profile_id: string;
  display_name: string;
  age: number;
  country_code: string;
  region: string | null;
  has_photo: boolean;
  conversation_id: string | null;
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

export default function LikesPage() {
  const { dictionary } = useI18n();
  const [tab, setTab] = useState<Tab>('received');
  const [items, setItems] = useState<CardItem[]>([]);
  const [page, setPage] = useState(1);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    // "Mutuos" en likes equivale a los matches ya existentes: reutilizamos
    // GET /matches en lugar de un /likes/mutual (que no existe).
    const path = tab === 'mutual' ? `/matches?page=${page}` : `/likes/${tab}?page=${page}`;

    apiFetch<LikesResponse | MatchesResponse>(path)
      .then((data) => {
        const normalized: CardItem[] = (data.items as Array<LikeItem | MatchItem>).map((item) => ({
          profile_id: item.profile_id,
          display_name: item.display_name,
          age: item.age,
          country_code: item.country_code,
          region: item.region,
          has_photo: item.has_photo,
          conversation_id: 'conversation_id' in item ? item.conversation_id ?? null : null,
        }));

        setItems(normalized);
        setCurrentPage(data.page);
        setTotalPages(data.total_pages);
        setTotal(data.total);
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError(dictionary.likes.errorUnauthorized);
        } else {
          setError(dictionary.likes.loadError);
        }
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [tab, page]);

  function changeTab(nextTab: Tab) {
    setTab(nextTab);
    setPage(1);
  }

  const emptyTitle: Record<Tab, string> = {
    received: dictionary.likes.emptyReceived,
    sent: dictionary.likes.emptySent,
    mutual: dictionary.likes.emptyMutual,
  };
  const emptySubtitle: Record<Tab, string> = {
    received: dictionary.likes.emptySubtitle,
    sent: dictionary.likes.emptySubtitle,
    mutual: dictionary.likes.emptyMutualSubtitle,
  };

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.likes.title}</h1>

      <div className={styles.tabs} role="tablist">
        <button
          type="button"
          className={tab === 'received' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('received')}
        >
          {dictionary.likes.tabReceived}
        </button>
        <button
          type="button"
          className={tab === 'sent' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('sent')}
        >
          {dictionary.likes.tabSent}
        </button>
        <button
          type="button"
          className={tab === 'mutual' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('mutual')}
        >
          {dictionary.likes.tabMutual}
        </button>
      </div>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>{dictionary.likes.loading}</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {items.length === 0 ? (
            <div className={styles.emptyState}>
              <p className={styles.emptyStateTitle}>{emptyTitle[tab]}</p>
              <p className={styles.emptyStateSubtitle}>{emptySubtitle[tab]}</p>
            </div>
          ) : (
            <div className={styles.grid}>
              {items.map((item) => (
                <div key={item.profile_id} className={styles.card}>
                  <Link href={`/profiles/${item.profile_id}`} className={styles.cardLink}>
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

                  {item.conversation_id && (
                    <Link
                      href={`/messages/${item.conversation_id}`}
                      className={styles.openConversation}
                    >
                      {dictionary.likes.openConversation}
                    </Link>
                  )}
                </div>
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
                {dictionary.likes.paginationInfo
                  .replace('{page}', String(currentPage))
                  .replace('{totalPages}', String(totalPages))
                  .replace('{total}', String(total))}
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
