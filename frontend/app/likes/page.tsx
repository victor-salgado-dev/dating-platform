'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, LikesResponse } from '@/lib/api';
import styles from './page.module.css';

function ProfilePhoto({ profileId, name }: { profileId: string; name: string }) {
  const [photoUrl, setPhotoUrl] = useState<string | null>(null);
  useEffect(() => {
    apiFetch<{ id: string; url: string }[]>(`/profiles/${profileId}/photos`)
      .then((photos) => photos[0] && setPhotoUrl(photos[0].url ?? `/api/v1/profiles/${profileId}/photos/${photos[0].id}/file`))
      .catch(() => setPhotoUrl(null));
  }, [profileId]);
  return photoUrl ? <img src={photoUrl} alt={name} className={styles.photoImg} /> : <div className={styles.photoPlaceholder}>Sin Foto</div>;
}

export default function LikesPage() {
  const [tab, setTab] = useState<'sent' | 'received'>('received');
  const [data, setData] = useState<LikesResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);
    apiFetch<LikesResponse>(`/likes/${tab}?page=${page}`)
      .then(setData)
      .catch((err: unknown) => setError(err instanceof ApiError && err.status === 401 ? 'Inicia sesión para ver tus likes.' : 'No se pudieron cargar los likes.'))
      .finally(() => setLoading(false));
  }, [tab, page]);

  function changeTab(nextTab: 'sent' | 'received') { setTab(nextTab); setPage(1); }
  const likes = data?.items ?? [];
  return (
    <main className={styles.main}>
      <h1 className={styles.title}>Likes</h1>
      <div className={styles.tabs} role="tablist">
        <button type="button" className={tab === 'received' ? styles.activeTab : styles.tab} onClick={() => changeTab('received')}>Recibidos</button>
        <button type="button" className={tab === 'sent' ? styles.activeTab : styles.tab} onClick={() => changeTab('sent')}>Enviados</button>
      </div>
      {loading && <p>Cargando...</p>}
      {error && <div className={styles.errorBanner}>{error}</div>}
      {!loading && !error && (likes.length === 0 ? <div className={styles.emptyState}>No hay likes {tab === 'received' ? 'recibidos' : 'enviados'} todavía.</div> : <div className={styles.grid}>
        {likes.map((item) => <Link key={item.profile_id} href={`/profiles/${item.profile_id}`} className={styles.card}>
          <div className={styles.imageContainer}>{item.has_photo ? <ProfilePhoto profileId={item.profile_id} name={item.display_name} /> : <div className={styles.photoPlaceholder}>Sin Foto</div>}</div>
          <div className={styles.cardInfo}><h2 className={styles.name}>{item.display_name}<span className={styles.age}> · {item.age}</span></h2><p className={styles.details}>{[item.region, item.country_code].filter(Boolean).join(', ')}</p></div>
        </Link>)}
      </div>)}
      {(data?.total_pages ?? 0) > 1 && <div className={styles.pagination}>
        <button type="button" className={styles.pageButton} disabled={page <= 1} onClick={() => setPage((value) => value - 1)}>Anterior</button>
        <span className={styles.pageInfo}>Página {data?.page} de {data?.total_pages} ({data?.total} likes)</span>
        <button type="button" className={styles.pageButton} disabled={page >= (data?.total_pages ?? 1)} onClick={() => setPage((value) => value + 1)}>Siguiente</button>
      </div>}
    </main>
  );
}
