'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, FavoritesResponse } from '@/lib/api';
import styles from './page.module.css';

interface PhotoItem {
  id: string;
  url: string;
  position: number;
}

function ProfilePhoto({ profileId, name }: { profileId: string; name: string }) {
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
    return <div className={styles.photoPlaceholder}>Sin Foto</div>;
  }

  return <img src={photoUrl} alt={name} className={styles.photoImg} />;
}

export default function FavoritesPage() {
  const [data, setData] = useState<FavoritesResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    apiFetch<FavoritesResponse>(`/favorites?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver tus favoritos.');
        } else {
          setError('No se pudieron cargar los favoritos.');
        }
      })
      .finally(() => setLoading(false));
  }, [page]);

  const favorites = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>Mis Favoritos</h1>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>Cargando favoritos…</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {favorites.length === 0 ? (
            <div className={styles.emptyState}>
              <p className={styles.emptyStateTitle}>Todavía no has añadido ningún favorito.</p>
              <p className={styles.emptyStateSubtitle}>
                Explora perfiles en el inicio y pulsa la estrella ★ para guardarlos aquí.
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
                      <div className={styles.photoPlaceholder}>Sin Foto</div>
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
                ← Anterior
              </button>
              <span className={styles.pageInfo}>
                Página {data?.page} de {totalPages} ({data?.total} favoritos)
              </span>
              <button
                type="button"
                className={styles.pageButton}
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
              >
                Siguiente →
              </button>
            </div>
          )}
        </>
      )}
    </main>
  );
}