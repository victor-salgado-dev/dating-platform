'use client';

import { useEffect, useState, useRef, type MouseEvent } from 'react';
import Link from 'next/link';

import { ActivityItem, ActivityResponse, apiFetch, ApiError } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import type { Dictionary } from '@/lib/i18n/dictionaries/es';
import type { Locale } from '@/lib/i18n/config';
import { PhotoGalleryModal } from '@/components/ProfileCard';
import { useProfileInteractions } from '@/lib/useProfileInteractions';
import { EmptyState, ErrorBanner } from '@/components/ListSectionState';
import LikesSection from '@/components/LikesSection';
import VisitsSection from '@/components/VisitsSection';
import FavoritesSection from '@/components/FavoritesSection';
import styles from './page.module.css';

type ActivityTab = 'feed' | 'likes' | 'visits' | 'favorites';

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
    return days === 1 ? dictionary.activity.relativeDaysOne : dictionary.activity.relativeDaysOther.replace('{n}', String(days));
  }
  return new Date(value).toLocaleDateString(locale === 'en' ? 'en-GB' : 'es-ES');
}

// Fila de "Novedades": mismo feed en lista de siempre, pero ahora con las
// mismas acciones (like / favorito / ver fotos) que las tarjetas de Home,
// en vez de ser solo un link al perfil.
function FeedRow({
  item,
  dictionary,
  locale,
}: {
  item: ActivityItem;
  dictionary: Dictionary;
  locale: Locale;
}) {
  const [isGalleryOpen, setIsGalleryOpen] = useState(false);
  const { likedIds, favoritedIds, toggleLike, toggleFavorite } = useProfileInteractions();
  const [liked, setLiked] = useState(false);
  const [favorited, setFavorited] = useState(false);

  useEffect(() => setLiked(likedIds.has(item.profile_id)), [likedIds, item.profile_id]);
  useEffect(() => setFavorited(favoritedIds.has(item.profile_id)), [favoritedIds, item.profile_id]);

  async function handleLike(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    const next = !liked;
    setLiked(next);
    toggleLike(item.profile_id, next);
    try {
      await apiFetch(`/likes/${item.profile_id}`, { method: next ? 'POST' : 'DELETE' });
    } catch {
      setLiked(!next);
      toggleLike(item.profile_id, !next);
    }
  }

  async function handleFavorite(e: MouseEvent) {
    e.preventDefault();
    e.stopPropagation();
    const next = !favorited;
    setFavorited(next);
    toggleFavorite(item.profile_id, next);
    try {
      await apiFetch(`/favorites/${item.profile_id}`, { method: next ? 'POST' : 'DELETE' });
    } catch {
      setFavorited(!next);
      toggleFavorite(item.profile_id, !next);
    }
  }

  return (
    <>
      <Link href={`/profiles/${item.profile_id}`} className={styles.item}>
        <div className={styles.photoFrame}>
          {item.photo_url ? (
            <img src={item.photo_url} alt={item.display_name} className={styles.photo} loading="lazy" />
          ) : (
            <div className={styles.photoPlaceholder}>{dictionary.common.noPhotoLower}</div>
          )}
        </div>
        <div className={styles.content}>
          <strong>{eventLabel(item.event_type, dictionary)}</strong>
          <span className={styles.name}>
            {item.display_name}, {item.age}
          </span>
          <span className={styles.meta}>
            {[item.region, item.country_code].filter(Boolean).join(', ')} · {relativeDate(item.created_at, dictionary, locale)}
          </span>
        </div>

        <div className={styles.rowActions}>
          {item.has_photo && (
            <button
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                setIsGalleryOpen(true);
              }}
              className={styles.rowActionBtn}
              title={dictionary.common.viewPhotos}
            >
              📸
            </button>
          )}
          <button
            onClick={handleLike}
            className={styles.rowActionBtn}
            style={liked ? { background: '#22c55e', border: 'none', boxShadow: '0 0 10px rgba(34,197,94,0.5)' } : undefined}
            title={liked ? dictionary.common.unlike : dictionary.common.like}
          >
            {liked ? '❤️' : '🤍'}
          </button>
          <button
            onClick={handleFavorite}
            className={styles.rowActionBtn}
            style={favorited ? { background: '#eab308', border: 'none', boxShadow: '0 0 10px rgba(234,179,8,0.5)' } : undefined}
            title={favorited ? dictionary.common.unfavorite : dictionary.common.favorite}
          >
            <span style={{ filter: favorited ? 'none' : 'grayscale(100%) opacity(0.6)' }}>⭐</span>
          </button>
        </div>

        <span className={styles.chevron} aria-hidden="true">›</span>
      </Link>

      {isGalleryOpen && (
        <PhotoGalleryModal profileId={item.profile_id} name={item.display_name} onClose={() => setIsGalleryOpen(false)} />
      )}
    </>
  );
}

function FeedTab() {
  const { locale, dictionary } = useI18n();
  const [data, setData] = useState<ActivityResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const lastFetchedPage = useRef<number | null>(null);

  useEffect(() => {
    if (lastFetchedPage.current === page) return;
    lastFetchedPage.current = page;

    setLoading(true);
    setError(null);
    apiFetch<ActivityResponse>(`/activity?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        setError(err instanceof ApiError && err.status === 401 ? dictionary.activity.errorUnauthorized : dictionary.activity.loadError);
      })
      .finally(() => setLoading(false));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [page]);

  const items = data?.items ?? [];

  return (
    <>
      {loading && <p>{dictionary.activity.loading}</p>}
      {error && <ErrorBanner message={error} className={styles.error} />}
      {!loading &&
        !error &&
        (items.length === 0 ? (
          <EmptyState title={dictionary.activity.empty} className={styles.empty} />
        ) : (
          <div className={styles.feed}>
            {items.map((item) => (
              <FeedRow key={`${item.event_type}-${item.profile_id}-${item.created_at}`} item={item} dictionary={dictionary} locale={locale} />
            ))}
          </div>
        ))}
      {(data?.total_pages ?? 0) > 1 && (
        <div className={styles.pagination}>
          <button type="button" className={styles.pageButton} disabled={page <= 1} onClick={() => setPage((v) => v - 1)}>
            {dictionary.common.paginationPrevPlain}
          </button>
          <span className={styles.pageInfo}>
            {dictionary.activity.paginationInfo
              .replace('{page}', String(data?.page ?? page))
              .replace('{totalPages}', String(data?.total_pages ?? 0))
              .replace('{total}', String(data?.total ?? 0))}
          </span>
          <button
            type="button"
            className={styles.pageButton}
            disabled={page >= (data?.total_pages ?? 1)}
            onClick={() => setPage((v) => v + 1)}
          >
            {dictionary.common.paginationNextPlain}
          </button>
        </div>
      )}
    </>
  );
}

export default function ActivityPage() {
  const { dictionary } = useI18n();
  const [tab, setTab] = useState<ActivityTab>('feed');

  const tabs: { key: ActivityTab; label: string }[] = [
    { key: 'feed', label: dictionary.activity.tabFeed },
    { key: 'likes', label: dictionary.nav.likes },
    { key: 'visits', label: dictionary.nav.visits },
    { key: 'favorites', label: dictionary.filters.favorites },
  ];

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>{dictionary.activity.title}</h1>

      <div className={styles.sectionTabs} role="tablist">
        {tabs.map((t) => (
          <button
            key={t.key}
            type="button"
            className={tab === t.key ? styles.sectionTabActive : styles.sectionTab}
            onClick={() => setTab(t.key)}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'feed' && <FeedTab />}
      {tab === 'likes' && <LikesSection />}
      {tab === 'visits' && <VisitsSection />}
      {tab === 'favorites' && <FavoritesSection />}
    </main>
  );
}
