'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, VisitsResponse } from '@/lib/api';
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

export default function VisitsPage() {
  const [tab, setTab] = useState<'received' | 'sent'>('received');
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
          setError('Inicia sesión para ver las visitas a tu perfil.');
        } else {
          setError('No se pudieron cargar las visitas.');
        }
      })
      .finally(() => setLoading(false));
  }, [tab, page]);

  function changeTab(nextTab: 'received' | 'sent') {
    setTab(nextTab);
    setPage(1);
  }

  const visits = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      <h1 className={styles.title}>Visitas</h1>

      <div className={styles.tabs} role="tablist">
        <button
          type="button"
          className={tab === 'received' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('received')}
        >
          Quién me ha visitado
        </button>
        <button
          type="button"
          className={tab === 'sent' ? styles.activeTab : styles.tab}
          onClick={() => changeTab('sent')}
        >
          Perfiles que visité
        </button>
      </div>

      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>Cargando visitas…</p>}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {visits.length === 0 ? (
            <div className={styles.emptyState}>
              <p className={styles.emptyStateTitle}>
                {tab === 'received'
                  ? 'Todavía nadie ha visitado tu perfil.'
                  : 'Todavía no has visitado ningún perfil.'}
              </p>
              <p className={styles.emptyStateSubtitle}>
                {tab === 'received'
                  ? 'Mantén tu perfil activo para recibir más visitas.'
                  : 'Explora perfiles en el inicio para ver a otros usuarios.'}
              </p>
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
                Página {data?.page} de {totalPages} ({data?.total} visitas)
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
