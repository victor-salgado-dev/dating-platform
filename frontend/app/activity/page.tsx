'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { ActivityItem, ActivityResponse, apiFetch, ApiError } from '@/lib/api';
import styles from './page.module.css';

function ProfilePhoto({ profileId, name }: { profileId: string; name: string }) {
  const [photoUrl, setPhotoUrl] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<{ id: string; url: string }[]>(`/profiles/${profileId}/photos`)
      .then((photos) => {
        if (photos[0]) setPhotoUrl(photos[0].url ?? `/api/v1/profiles/${profileId}/photos/${photos[0].id}/file`);
      })
      .catch(() => setPhotoUrl(null));
  }, [profileId]);

  return photoUrl ? <img src={photoUrl} alt={name} className={styles.photo} /> : <div className={styles.photoPlaceholder}>Sin foto</div>;
}

function eventLabel(type: ActivityItem['event_type']): string {
  if (type === 'like_received') return 'Te dio like';
  if (type === 'match_created') return 'Nuevo match';
  return 'Te agregó a favoritos';
}

function relativeDate(value: string): string {
  const seconds = Math.max(0, Math.floor((Date.now() - new Date(value).getTime()) / 1000));
  if (seconds < 60) return 'Ahora';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `Hace ${minutes} min`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `Hace ${hours} h`;
  const days = Math.floor(hours / 24);
  if (days < 7) return `Hace ${days} ${days === 1 ? 'día' : 'días'}`;
  return new Date(value).toLocaleDateString('es-ES');
}

export default function ActivityPage() {
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
        setError(err instanceof ApiError && err.status === 401 ? 'Inicia sesión para ver tu actividad.' : 'No se pudo cargar la actividad.');
      })
      .finally(() => setLoading(false));
  }, [page]);

  const items = data?.items ?? [];
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>Actividad</h1>
      {loading && <p>Cargando actividad...</p>}
      {error && <div className={styles.error}>{error}</div>}
      {!loading && !error && (items.length === 0 ? (
        <div className={styles.empty}>Todavía no tienes actividad.</div>
      ) : (
        <div className={styles.feed}>
          {items.map((item) => (
            <Link key={`${item.event_type}-${item.profile_id}-${item.created_at}`} href={`/profiles/${item.profile_id}`} className={styles.item}>
              <div className={styles.photoFrame}>
                {item.has_photo ? <ProfilePhoto profileId={item.profile_id} name={item.display_name} /> : <div className={styles.photoPlaceholder}>Sin foto</div>}
              </div>
              <div className={styles.content}>
                <strong>{eventLabel(item.event_type)}</strong>
                <span className={styles.name}>{item.display_name}, {item.age}</span>
                <span className={styles.meta}>{[item.region, item.country_code].filter(Boolean).join(', ')} · {relativeDate(item.created_at)}</span>
              </div>
              <span className={styles.chevron} aria-hidden="true">›</span>
            </Link>
          ))}
        </div>
      ))}
      {(data?.total_pages ?? 0) > 1 && <div className={styles.pagination}>
        <button type="button" className={styles.pageButton} disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>Anterior</button>
        <span className={styles.pageInfo}>Página {data?.page} de {data?.total_pages} ({data?.total} eventos)</span>
        <button type="button" className={styles.pageButton} disabled={page >= (data?.total_pages ?? 1)} onClick={() => setPage((value) => value + 1)}>Siguiente</button>
      </div>}
    </main>
  );
}
