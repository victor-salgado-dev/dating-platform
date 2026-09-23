'use client';

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiFetch, SearchResponse, LikesResponse, FavoritesResponse } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
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

// -----------------------------------------------------------------------------
// Componente de la Galería Modal
// -----------------------------------------------------------------------------
function PhotoGalleryModal({ profileId, name, onClose }: { profileId: string, name: string, onClose: () => void }) {
  const { dictionary } = useI18n();
  const [photos, setPhotos] = useState<PhotoItem[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch<PhotoItem[]>(`/profiles/${profileId}/photos`)
      .then((data) => setPhotos(data || []))
      .catch(() => setPhotos([]))
      .finally(() => setLoading(false));
  }, [profileId]);

  const currentPhotoUrl = photos.length > 0
    ? (photos[currentIndex].url ?? `/api/v1/profiles/${profileId}/photos/${photos[currentIndex].id}/file`)
    : null;

  return (
    <div
      onClick={onClose}
      style={{ position: 'fixed', top: 0, left: 0, right: 0, bottom: 0, backgroundColor: 'rgba(0,0,0,0.9)', zIndex: 9999, display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center' }}
    >
      <button
        onClick={onClose}
        style={{ position: 'absolute', top: '20px', right: '30px', background: 'none', border: 'none', color: 'white', fontSize: '2.5rem', cursor: 'pointer', zIndex: 10, padding: '10px' }}
      >
        ✕
      </button>

      <div onClick={(e) => e.stopPropagation()} style={{ position: 'relative', display: 'flex', flexDirection: 'column', alignItems: 'center', width: '100%', maxWidth: '800px' }}>
        {loading ? (
          <p style={{ color: 'white', fontSize: '1.2rem' }}>{dictionary.common.photoGalleryLoading}</p>
        ) : photos.length === 0 ? (
          <p style={{ color: 'white', fontSize: '1.2rem' }}>{dictionary.common.photoGalleryEmpty}</p>
        ) : (
          <div style={{ position: 'relative', display: 'flex', alignItems: 'center', justifyContent: 'center', width: '100%' }}>
            {photos.length > 1 && (
              <button
                onClick={() => setCurrentIndex((prev) => (prev > 0 ? prev - 1 : photos.length - 1))}
                style={{ position: 'absolute', left: '10px', background: 'rgba(255,255,255,0.15)', border: '2px solid rgba(255,255,255,0.5)', color: 'white', fontSize: '2.5rem', cursor: 'pointer', width: '50px', height: '50px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 10 }}
              >
                ‹
              </button>
            )}

            <img
              src={currentPhotoUrl!}
              alt={dictionary.common.photoGalleryAlt.replace('{name}', name)}
              style={{ maxHeight: '80vh', maxWidth: '90vw', objectFit: 'contain', borderRadius: '8px', boxShadow: '0 10px 30px rgba(0,0,0,0.5)' }}
            />

            {photos.length > 1 && (
              <button
                onClick={() => setCurrentIndex((prev) => (prev < photos.length - 1 ? prev + 1 : 0))}
                style={{ position: 'absolute', right: '10px', background: 'rgba(255,255,255,0.15)', border: '2px solid rgba(255,255,255,0.5)', color: 'white', fontSize: '2.5rem', cursor: 'pointer', width: '50px', height: '50px', borderRadius: '50%', display: 'flex', alignItems: 'center', justifyContent: 'center', zIndex: 10 }}
              >
                ›
              </button>
            )}
          </div>
        )}

        {!loading && photos.length > 0 && (
          <p style={{ color: 'rgba(255,255,255,0.8)', marginTop: '1.5rem', fontSize: '1.2rem', fontWeight: '500' }}>
            {dictionary.common.photoGalleryCounter
              .replace('{name}', name)
              .replace('{current}', String(currentIndex + 1))
              .replace('{total}', String(photos.length))}
          </p>
        )}
      </div>
    </div>
  );
}

// -----------------------------------------------------------------------------
// Componente Principal de la Tarjeta
// -----------------------------------------------------------------------------
function ProfileCard({
  profile,
  isPremium,
  initialLiked,
  initialFavorited,
  receivedLike,
  receivedFavorite,
}: {
  profile: ProfileItem;
  isPremium: boolean;
  initialLiked: boolean;
  initialFavorited: boolean;
  receivedLike: boolean;
  receivedFavorite: boolean;
}) {
  const { dictionary } = useI18n();
  const [photoUrl, setPhotoUrl] = useState<string | null>(null);
  const [isGalleryOpen, setIsGalleryOpen] = useState(false);

  // El estado arranca con el valor REAL que viene de la base de datos
  const [liked, setLiked] = useState(initialLiked);
  const [favorited, setFavorited] = useState(initialFavorited);

  // Si cambia la información de la base de datos, actualizamos el botón
  useEffect(() => {
    setLiked(initialLiked);
  }, [initialLiked]);

  useEffect(() => {
    setFavorited(initialFavorited);
  }, [initialFavorited]);

  useEffect(() => {
    if (profile.has_photo) {
      apiFetch<PhotoItem[]>(`/profiles/${profile.profile_id}/photos`)
        .then((photos) => {
          if (photos && photos.length > 0) {
            setPhotoUrl(photos[0].url ?? `/api/v1/profiles/${profile.profile_id}/photos/${photos[0].id}/file`);
          }
        })
        .catch(() => setPhotoUrl(null));
    }
  }, [profile.profile_id, profile.has_photo]);

  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      if (liked) {
        await apiFetch(`/likes/${profile.profile_id}`, { method: 'DELETE' });
        setLiked(false);
      } else {
        await apiFetch(`/likes/${profile.profile_id}`, { method: 'POST' });
        setLiked(true);
      }
    } catch (err) {
      console.error('Error al procesar like', err);
    }
  };

  const handleFavorite = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    try {
      if (favorited) {
        await apiFetch(`/favorites/${profile.profile_id}`, { method: 'DELETE' });
        setFavorited(false);
      } else {
        await apiFetch(`/favorites/${profile.profile_id}`, { method: 'POST' });
        setFavorited(true);
      }
    } catch (err) {
      console.error('Error al procesar favorito', err);
    }
  };

  const receivedClass = receivedLike && receivedFavorite
    ? styles.receivedBoth
    : receivedLike
    ? styles.receivedLike
    : receivedFavorite
    ? styles.receivedFav
    : '';

  return (
    <>
      <Link
        href={`/profiles/${profile.profile_id}`}
        className={`${styles.card} ${isPremium ? styles.cardPremium : ''} ${receivedClass}`}
      >
        <div className={styles.imageContainer}>
          {photoUrl ? (
            <img src={photoUrl} alt={profile.display_name} className={styles.photoImg} />
          ) : (
            <div className={styles.photoPlaceholder}>{dictionary.common.noPhoto}</div>
          )}
          {(receivedLike || receivedFavorite) && (
            <div className={styles.receivedBadges}>
              {receivedLike && <span className={styles.badgeLike}>♥</span>}
              {receivedFavorite && <span className={styles.badgeFav}>★</span>}
            </div>
          )}
        </div>

        <div className={styles.cardInfo}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
            <div>
              <h3 className={styles.name}>
                {profile.display_name}
                <span style={{ fontWeight: 'normal', color: '#9ca3af' }}> · {profile.age}</span>
              </h3>
              <p className={styles.details}>
                {profile.region ? `${profile.region}, ${profile.country_code}` : profile.country_code}
              </p>
            </div>

            <div style={{ display: 'flex', gap: '8px' }}>
              {profile.has_photo && (
                <button
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setIsGalleryOpen(true);
                  }}
                  style={{
                    background: '#f3f4f6', border: '1px solid #e5e7eb', borderRadius: '50%', width: '34px', height: '34px', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', fontSize: '1rem', transition: 'all 0.2s'
                  }}
                  title={dictionary.common.viewPhotos}
                >
                  📸
                </button>
              )}

              {/* Botón Like con Datos Reales */}
              <button
                onClick={handleLike}
                style={{
                  background: liked ? '#22c55e' : '#f3f4f6',
                  border: liked ? 'none' : '1px solid #e5e7eb',
                  boxShadow: liked ? '0 0 10px rgba(34, 197, 94, 0.5)' : 'none',
                  borderRadius: '50%', width: '34px', height: '34px', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', fontSize: '1rem', transition: 'all 0.2s'
                }}
                title={liked ? dictionary.common.unlike : dictionary.common.like}
              >
                {liked ? '❤️' : '🤍'}
              </button>

              {/* Botón Favorito con Datos Reales */}
              <button
                onClick={handleFavorite}
                style={{
                  background: favorited ? '#eab308' : '#f3f4f6',
                  border: favorited ? 'none' : '1px solid #e5e7eb',
                  boxShadow: favorited ? '0 0 10px rgba(234, 179, 8, 0.5)' : 'none',
                  borderRadius: '50%', width: '34px', height: '34px', display: 'flex', alignItems: 'center', justifyContent: 'center', cursor: 'pointer', fontSize: '1.1rem', transition: 'all 0.2s'
                }}
                title={favorited ? dictionary.common.unfavorite : dictionary.common.favorite}
              >
                <span style={{ filter: favorited ? 'none' : 'grayscale(100%) opacity(0.6)' }}>⭐</span>
              </button>
            </div>
          </div>
        </div>
      </Link>

      {isGalleryOpen && (
        <PhotoGalleryModal
          profileId={profile.profile_id}
          name={profile.display_name}
          onClose={() => setIsGalleryOpen(false)}
        />
      )}
    </>
  );
}

// -----------------------------------------------------------------------------
// Página Principal
// -----------------------------------------------------------------------------
export default function HomePage() {
  const { dictionary } = useI18n();
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);

  // Listas de perfiles con like/favorito traídas del servidor
  const [likedIds, setLikedIds] = useState<Set<string>>(new Set());
  const [favoritedIds, setFavoritedIds] = useState<Set<string>>(new Set());

  // Quién te ha dado like / favorito a TI (para el borde + insignia en la tarjeta)
  const [receivedLikeIds, setReceivedLikeIds] = useState<Set<string>>(new Set());
  const [receivedFavIds, setReceivedFavIds] = useState<Set<string>>(new Set());

  const [randomAdIndex, setRandomAdIndex] = useState(-1);

  // 1. Cargar perfiles
  useEffect(() => {
    setRandomAdIndex(Math.floor(Math.random() * 20) + 2);
    setLoading(true);

    apiFetch<SearchResponse>(`/search/profiles?page=${page}&page_size=24`)
      .then(setData)
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  }, [page]);

  // 2. Cargar likes y favoritos reales del usuario autenticado (Solo 2 peticiones)
  useEffect(() => {
    // Pedimos los likes enviados
    apiFetch<LikesResponse>('/likes/sent?page_size=100')
      .then((res) => {
        const ids = new Set(res.items.map((item) => item.profile_id));
        setLikedIds(ids);
      })
      .catch(() => {}); // Si no está logueado, silenciosamente no hace nada

    // Pedimos los favoritos
    apiFetch<FavoritesResponse>('/favorites?page_size=100')
      .then((res) => {
        const ids = new Set(res.items.map((item) => item.profile_id));
        setFavoritedIds(ids);
      })
      .catch(() => {});

    // Quién te ha dado like / favorito a ti. OJO: estos dos endpoints
    // (/likes/received y /favorites/received) son una suposición siguiendo
    // el mismo patrón que /likes/sent — confirma o ajusta las rutas reales
    // de tu API antes de dar esto por definitivo.
    apiFetch<LikesResponse>('/likes/received?page_size=100')
      .then((res) => {
        const ids = new Set(res.items.map((item) => item.profile_id));
        setReceivedLikeIds(ids);
      })
      .catch(() => {});

    apiFetch<FavoritesResponse>('/favorites/received?page_size=100')
      .then((res) => {
        const ids = new Set(res.items.map((item) => item.profile_id));
        setReceivedFavIds(ids);
      })
      .catch(() => {});
  }, []);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <main className={styles.main}>
      {loading && <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>{dictionary.home.loading}</p>}

      <div className={styles.grid}>
        {profiles.map((profile, index) => {
          const isPremium = index === 0;
          const showAdSquare = index === randomAdIndex;
          const showBannerHorizontal = index === 14;

          // Comprobamos contra la base de datos real
          const isLiked = likedIds.has(profile.profile_id);
          const isFavorited = favoritedIds.has(profile.profile_id);
          const gotLike = receivedLikeIds.has(profile.profile_id);
          const gotFavorite = receivedFavIds.has(profile.profile_id);

          return (
            <React.Fragment key={profile.profile_id}>
              {showAdSquare && <div className={styles.adSquare}>{dictionary.discover.sponsoredRandom}</div>}
              {showBannerHorizontal && <div className={styles.adBanner}>{dictionary.discover.bannerMid}</div>}

              <ProfileCard
                profile={profile}
                isPremium={isPremium}
                initialLiked={isLiked}
                initialFavorited={isFavorited}
                receivedLike={gotLike}
                receivedFavorite={gotFavorite}
              />
            </React.Fragment>
          );
        })}
      </div>

      {!loading && profiles.length > 0 && (
        <>
          <div className={styles.adBanner} style={{ marginTop: '3rem' }}>
            {dictionary.home.bannerBottom}
          </div>

          {totalPages > 1 && (
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
        </>
      )}
    </main>
  );
}
