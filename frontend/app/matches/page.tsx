'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, MatchesResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import styles from './page.module.css';

export default function MatchesPage() {
  const { dictionary } = useI18n();
  const [data, setData] = useState<MatchesResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    apiFetch<MatchesResponse>(`/matches?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        setError(
          err instanceof ApiError && err.status === 401
            ? dictionary.matches.errorUnauthorized
            : dictionary.matches.loadError
        );
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  const matches = data?.items ?? [];
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.matches.title}</h1>
      {loading && <p>{dictionary.matches.loading}</p>}
      {error && <div className={styles.errorBanner}>{error}</div>}
      {!loading && !error && (matches.length === 0
        ? <div className={styles.emptyState}>{dictionary.matches.empty}</div>
        : <div className={styles.grid}>
          {matches.map((item) => <article key={item.profile_id} className={styles.card}>
            <Link href={`/profiles/${item.profile_id}`} className={styles.profileLink}>
              <div className={styles.imageContainer}>{item.photo_url ? <img src={item.photo_url} alt={item.display_name} className={styles.photoImg} /> : <div className={styles.photoPlaceholder}>{dictionary.common.noPhoto}</div>}</div>
              <div className={styles.cardInfo}><h2 className={styles.name}>{item.display_name}<span className={styles.age}> · {item.age}</span></h2><p className={styles.details}>{[item.region, item.country_code].filter(Boolean).join(', ')}</p></div>
            </Link>
            {item.conversation_id && <Link href={`/messages/${item.conversation_id}`} className={styles.messageLink}>{dictionary.matches.openConversation}</Link>}
          </article>)}
        </div>)}
      {(data?.total_pages ?? 0) > 1 && <div className={styles.pagination}>
        <button type="button" className={styles.pageButton} disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>{dictionary.common.paginationPrevPlain}</button>
        <span className={styles.pageInfo}>
          {dictionary.matches.paginationInfo
            .replace('{page}', String(data?.page ?? page))
            .replace('{totalPages}', String(data?.total_pages ?? 0))
            .replace('{total}', String(data?.total ?? 0))}
        </span>
        <button type="button" className={styles.pageButton} disabled={page >= (data?.total_pages ?? 1)} onClick={() => setPage((value) => value + 1)}>{dictionary.common.paginationNextPlain}</button>
      </div>}
    </main>
  );
}
