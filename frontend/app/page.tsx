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
  
  // Estado para la posición aleatoria del anuncio cuadrado
  const [randomAdIndex, setRandomAdIndex] = useState(-1);

  useEffect(() => {
    // Escogemos un número al azar entre 2 y 22 para inyectar el anuncio en la grilla
    setRandomAdIndex(Math.floor(Math.random() * 20) + 2);
    
    setLoading(true);
    apiFetch<SearchResponse>(`/search/profiles?page=${page}&page_size=24`)
      .then(setData)
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  }, [page]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      {loading && <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>Buscando perfiles cerca de ti...</p>}

      <div className={styles.grid}>
        {profiles.map((profile, index) => {
          const isPremium = index === 0; // El primero es destacado
          const showAdSquare = index === randomAdIndex; // Solo 1 anuncio aleatorio
          const showBannerHorizontal = index === 14; // Banner intermedio

          return (
            <React.Fragment key={profile.profile_id}>
              {showAdSquare && (
                <div className={styles.adSquare}>[ANUNCIO PATROCINADO - ALEATORIO]</div>
              )}

              {showBannerHorizontal && (
                <div className={styles.adBanner}>[ESPACIO PUBLICITARIO - INTERMEDIO]</div>
              )}

              <Link
                href={`/profiles/${profile.profile_id}`}
                className={`${styles.card} ${isPremium ? styles.cardPremium : ''}`}
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
                    <span style={{ fontWeight: 'normal', color: '#9ca3af' }}> · {profile.age}</span>
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

      {!loading && profiles.length > 0 && (
        <>
          {/* Banner Horizontal Extra justo encima de la paginación */}
          <div className={styles.adBanner} style={{ marginTop: '3rem' }}>
            [ESPACIO PUBLICITARIO - INFERIOR]
          </div>

          {/* Barra de Navegación de Páginas */}
          <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '1rem', margin: '2rem 0' }}>
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              style={{ padding: '0.75rem 1.5rem', cursor: page <= 1 ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
            >
              ← Anterior
            </button>
            <span style={{ fontWeight: 600, color: '#444' }}>Página {page} de {totalPages}</span>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              style={{ padding: '0.75rem 1.5rem', cursor: page >= totalPages ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
            >
              Siguiente →
            </button>
          </div>
        </>
      )}
    </main>
  );
}