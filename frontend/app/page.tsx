'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiFetch, SearchResponse } from '@/lib/api';
import styles from './page.module.css';

interface ProfileItem {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  has_photo: boolean;
}

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

export default function HomePage() {
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    // Pedimos la página actual con 24 perfiles por página
    apiFetch<SearchResponse>(`/search/profiles?page=${page}&page_size=24`)
      .then(setData)
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  }, [page]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      {loading && <p style={{ textAlign: 'center', padding: '2rem' }}>Cargando perfiles...</p>}

      <div className={styles.grid}>
        {profiles.map((profile, index) => {
          const isPremium = index === 0;
          const showBannerHorizontal = index === 14;
          const showAdSquare1 = index === 5;
          const showAdSquare2 = index === 11;

          return (
            <React.Fragment key={profile.profile_id}>
              {showAdSquare1 && (
                <div className={styles.adSquare}>[ANUNCIO PATROCINADO - 2 COLUMNAS]</div>
              )}

              {showAdSquare2 && (
                <div className={styles.adSquare}>[ANUNCIO PATROCINADO - 2 COLUMNAS]</div>
              )}

              {showBannerHorizontal && (
                <div className={styles.adBanner}>[ESPACIO PUBLICITARIO - BANNER HORIZONTAL]</div>
              )}

              <Link
                href={`/profiles/${profile.profile_id}`}
                className={`${styles.card} ${isPremium ? styles.cardPremium : ''}`}
                style={{ textDecoration: 'none', color: 'inherit' }}
              >
                <div className={styles.imageContainer}>
                  {profile.has_photo ? (
                    <ProfilePhoto profileId={profile.profile_id} name={profile.display_name} />
                  ) : (
                    <div className={styles.photoPlaceholder}>Sin Foto</div>
                  )}
                </div>

                <div className={styles.cardInfo}>
                  <h3 className={styles.name}>
                    {profile.display_name}
                    <span style={{ fontWeight: 'normal', color: '#666' }}> · {profile.age}</span>
                  </h3>
                  <p className={styles.details}>
                    {profile.region ? `${profile.region}, ` : ''}
                    {profile.country_code}
                  </p>
                </div>
              </Link>
            </React.Fragment>
          );
        })}
      </div>

      {/* Barra de Navegación de Páginas */}
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '1rem', margin: '3rem 0' }}>
        <button
          onClick={() => setPage((p) => Math.max(1, p - 1))}
          disabled={page <= 1}
          style={{ padding: '0.6rem 1.2rem', cursor: page <= 1 ? 'not-allowed' : 'pointer', borderRadius: '4px', border: '1px solid #ccc', background: '#fff' }}
        >
          ← Anterior
        </button>
        <span style={{ fontWeight: 600 }}>Página {page} de {totalPages}</span>
        <button
          onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
          disabled={page >= totalPages}
          style={{ padding: '0.6rem 1.2rem', cursor: page >= totalPages ? 'not-allowed' : 'pointer', borderRadius: '4px', border: '1px solid #ccc', background: '#fff' }}
        >
          Siguiente →
        </button>
      </div>
    </main>
  );
}