'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { ActivityItem, ActivityResponse, apiFetch, ApiError } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
import type { Locale } from '@/lib/i18n/config';
import styles from './page.module.css';

function ProfilePhoto({ profileId, name }: { profileId: string; name: string }) {
  const { dictionary } = useI18n();
  const [photoUrl, setPhotoUrl] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<{ id: string; url: string }[]>(`/profiles/${profileId}/photos`)
      .then((photos) => {
        if (photos[0]) setPhotoUrl(photos[0].url ?? `/api/v1/profiles/${profileId}/photos/${photos[0].id}/file`);
      })
      .catch(() => setPhotoUrl(null));
  }, [profileId]);

  return photoUrl ? (
    <img src={photoUrl} alt={name} className={styles.photo} />
  ) : (
    <div className={styles.photoPlaceholder}>{dictionary.common.noPhotoLower}</div>
  );
}

function eventLabel(type: ActivityItem['event_type'], dictionary: Dictionary): string {
  if (type === 'like_received') return dictionary.activity.eventLikeReceived;
  if (type === 'match_created') return dictionary.activity.eventMatchCreated;
  return dictionary.activity.eventFavoriteReceived;
}

function relativeDate(value: string, dictionary: Dictionary, locale: Locale): string {
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000));
  if (seconds < 60) return dictionary.activity.relativeNow;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return dictionary.activity.relativeMinutes.replace('{n}', String(minutes));
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return dictionary.activity.relativeHours.replace('{n}', String(hours));
  const days = Math.floor(hours / 24);
  if (days < 7) {
    return days === 1
      ? dictionary.activity.relativeDaysOne
      : dictionary.activity.relativeDaysOther.replace('{n}', String(days));
  }
  return new Date(value).toLocaleDateString(locale === 'en' ? 'en-GB' : 'es-ES');
}

export default function ActivityPage() {
  const { locale, dictionary } = useI18n();
  const [data, setData] = useState<ActivityResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    apiFetch<ActivityResponse>(`/activity?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        setError(
          err instanceof ApiError && err.status === 401
            ? dictionary.activity.errorUnauthorized
            : dictionary.activity.loadError
        );
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  const items = data?.items ?? [];
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.activity.title}</h1>
      {loading && <p>{dictionary.activity.loading}</p>}
      {error && <div className={styles.error}>{error}</div>}
      {!loading && !error && (items.length === 0 ? (
        <div className={styles.empty}>{dictionary.activity.empty}</div>
      ) : (
        <div className={styles.feed}>
          {items.map((item) => (
            <Link key={`${item.event_type}-${item.profile_id}-${item.created_at}`} href={`/profiles/${item.profile_id}`} className={styles.item}>
              <div className={styles.photoFrame}>
                {item.has_photo ? <ProfilePhoto profileId={item.profile_id} name={item.display_name} /> : <div className={styles.photoPlaceholder}>{dictionary.common.noPhotoLower}</div>}
              </div>
              <div className={styles.content}>
                <strong>{eventLabel(item.event_type, dictionary)}</strong>
                <span className={styles.name}>{item.display_name}, {item.age}</span>
                <span className={styles.meta}>{[item.region, item.country_code].filter(Boolean).join(', ')} · {relativeDate(item.created_at, dictionary, locale)}</span>
              </div>
              <span className={styles.chevron} aria-hidden="true">›</span>
            </Link>
          ))}
        </div>
      ))}
      {(data?.total_pages ?? 0) > 1 && <div className={styles.pagination}>
        <button type="button" className={styles.pageButton} disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>{dictionary.common.paginationPrevPlain}</button>
        <span className={styles.pageInfo}>
          {dictionary.activity.paginationInfo
            .replace('{page}', String(data?.page ?? page))
            .replace('{totalPages}', String(data?.total_pages ?? 0))
            .replace('{total}', String(data?.total ?? 0))}
        </span>
        <button type="button" className={styles.pageButton} disabled={page >= (data?.total_pages ?? 1)} onClick={() => setPage((value) => value + 1)}>{dictionary.common.paginationNextPlain}</button>
      </div>}
    </main>
  );
}
