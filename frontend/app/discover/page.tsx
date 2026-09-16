'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, SearchResponse } from '@/lib/api';
import styles from './page.module.css';

interface PhotoItem {
  id: string;
  url: string;
  position: number;
}

// Componente que carga la foto real de cada usuario
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

export default function DiscoverPage() {
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    // 24 perfiles por página
    apiFetch<SearchResponse>(`/search/profiles?page=${page}&page_size=24`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver perfiles.');
        } else {
          setError('No se pudieron cargar los resultados.');
        }
      })
      .finally(() => setLoading(false));
  }, [page]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>Descubrir</h1>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>Cargando perfiles…</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {profiles.length === 0 ? (
            <div className={styles.emptyState}>
              <p>No hay resultados disponibles.</p>
            </div>
          ) : (
            <ul className={styles.grid}>
              {profiles.map((item) => (
                <li key={item.profile_id} className={styles.card}>
                  <Link href={`/profiles/${item.profile_id}`} className={styles.cardLink}>
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
                      {item.relationship_goal && (
                        <span className={styles.cardGoal}>{item.relationship_goal}</span>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
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
                Página {data?.page} de {totalPages} ({data?.total} perfiles)
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