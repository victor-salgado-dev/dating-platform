'use client';

import React, { useEffect, useState, useRef } from 'react';
import { apiFetch, NewMembersResponse, SearchResultItem } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { ProfileCard } from '@/components/ProfileCard';
import { useProfileInteractions } from '@/lib/useProfileInteractions';
import styles from '../page.module.css';

export default function OnlineNowPage() {
  const { dictionary } = useI18n();
  const [data, setData] = useState<NewMembersResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const { likedIds, favoritedIds, receivedLikeIds, receivedFavIds, toggleLike, toggleFavorite } =
    useProfileInteractions();

  const lastFetchedPage = useRef<number | null>(null);

  useEffect(() => {
    if (lastFetchedPage.current === page) return;
    lastFetchedPage.current = page;

    setLoading(true);
    setError(null);

    apiFetch<NewMembersResponse>(`/search/online-now?page=${page}&page_size=24`)
      .then(setData)
      .catch(() => {
        setError(dictionary.discover.loadError);
      })
      .finally(() => setLoading(false));
  }, [page, dictionary]);

  const profiles: SearchResultItem[] = (data?.items ?? []).map((item) => ({
    profile_id: item.profile_id,
    display_name: item.display_name,
    age: item.age,
    gender: item.gender,
    country_code: item.country_code,
    region: item.region,
    relationship_goals: null,
    has_photo: item.photo_url !== null,
    photo_url: item.photo_url,
    created_at: item.created_at,
  }));

  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      <h1 style={{ fontSize: '1.8rem', margin: 0, marginBottom: '1.5rem', color: '#111827' }}>
        {dictionary.home.tabs.online}
      </h1>

      {loading && (
        <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>
          {dictionary.home.loading}
        </p>
      )}

      {error && (
        <div style={{ textAlign: 'center', padding: '2rem', color: '#b91c1c', fontWeight: 600 }}>
          {error}
        </div>
      )}

      {!loading && !error && (
        <div className={styles.grid}>
          {profiles.map((profile, index) => {
            const isPremium = index === 0;

            return (
              <ProfileCard
                key={profile.profile_id}
                profile={profile}
                isPremium={isPremium}
                initialLiked={likedIds.has(profile.profile_id)}
                initialFavorited={favoritedIds.has(profile.profile_id)}
                receivedLike={receivedLikeIds.has(profile.profile_id)}
                receivedFavorite={receivedFavIds.has(profile.profile_id)}
                onToggleLike={toggleLike}
                onToggleFavorite={toggleFavorite}
              />
            );
          })}
        </div>
      )}

      {!loading && !error && profiles.length > 0 && totalPages > 1 && (
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
    </main>
  );
}
