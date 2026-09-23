'use client';

import React, { useEffect, useState, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';

import {
  apiFetch,
  ApiError,
  SearchResponse,
  LikesResponse,
  FavoritesResponse,
} from '@/lib/api';
import styles from './page.module.css';

interface ProfileItem {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  relationship_goal?: string | null;
  relationship_goals?: string[] | null;
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
function PhotoGalleryModal({
  profileId,
  name,
  onClose,
}: {
  profileId: string;
  name: string;
  onClose: () => void;
}) {
  const [photos, setPhotos] = useState<PhotoItem[]>([]);
  const [currentIndex, setCurrentIndex] = useState(0);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    apiFetch<PhotoItem[]>(`/profiles/${profileId}/photos`)
      .then((data) => setPhotos(data || []))
      .catch(() => setPhotos([]))
      .finally(() => setLoading(false));
  }, [profileId]);

  const currentPhotoUrl =
    photos.length > 0
      ? (photos[currentIndex].url ?? `/api/v1/profiles/${profileId}/photos/${photos[currentIndex].id}/file`)
      : null;

  return (
    <div
      onClick={onClose}
      style={{
        position: 'fixed',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        backgroundColor: 'rgba(0,0,0,0.9)',
        zIndex: 9999,
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
        justifyContent: 'center',
      }}
    >
      <button
        onClick={onClose}
        style={{
          position: 'absolute',
          top: '20px',
          right: '30px',
          background: 'none',
          border: 'none',
          color: 'white',
          fontSize: '2.5rem',
          cursor: 'pointer',
          zIndex: 10,
          padding: '10px',
        }}
      >
        ✕
      </button>

      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          position: 'relative',
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          width: '100%',
          maxWidth: '800px',
        }}
      >
        {loading ? (
          <p style={{ color: 'white', fontSize: '1.2rem' }}>Cargando fotos...</p>
        ) : photos.length === 0 ? (
          <p style={{ color: 'white', fontSize: '1.2rem' }}>Este perfil no tiene fotos.</p>
        ) : (
          <div
            style={{
              position: 'relative',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              width: '100%',
            }}
          >
            {photos.length > 1 && (
              <button
                onClick={() => setCurrentIndex((prev) => (prev > 0 ? prev - 1 : photos.length - 1))}
                style={{
                  position: 'absolute',
                  left: '10px',
                  background: 'rgba(255,255,255,0.15)',
                  border: '2px solid rgba(255,255,255,0.5)',
                  color: 'white',
                  fontSize: '2.5rem',
                  cursor: 'pointer',
                  width: '50px',
                  height: '50px',
                  borderRadius: '50%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  zIndex: 10,
                }}
              >
                ‹
              </button>
            )}

            <img
              src={currentPhotoUrl!}
              alt={`Foto de ${name}`}
              style={{
                maxHeight: '80vh',
                maxWidth: '90vw',
                objectFit: 'contain',
                borderRadius: '8px',
                boxShadow: '0 10px 30px rgba(0,0,0,0.5)',
              }}
            />

            {photos.length > 1 && (
              <button
                onClick={() => setCurrentIndex((prev) => (prev < photos.length - 1 ? prev + 1 : 0))}
                style={{
                  position: 'absolute',
                  right: '10px',
                  background: 'rgba(255,255,255,0.15)',
                  border: '2px solid rgba(255,255,255,0.5)',
                  color: 'white',
                  fontSize: '2.5rem',
                  cursor: 'pointer',
                  width: '50px',
                  height: '50px',
                  borderRadius: '50%',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  zIndex: 10,
                }}
              >
                ›
              </button>
            )}
          </div>
        )}

        {!loading && photos.length > 0 && (
          <p
            style={{
              color: 'rgba(255,255,255,0.8)',
              marginTop: '1.5rem',
              fontSize: '1.2rem',
              fontWeight: '500',
            }}
          >
            {name} — {currentIndex + 1} de {photos.length}
          </p>
        )}
      </div>
    </div>
  );
}

// -----------------------------------------------------------------------------
// Componente de Tarjeta de Perfil
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
  const [photoUrl, setPhotoUrl] = useState<string | null>(null);
  const [isGalleryOpen, setIsGalleryOpen] = useState(false);
  const [liked, setLiked] = useState(initialLiked);
  const [favorited, setFavorited] = useState(initialFavorited);

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
            setPhotoUrl(
              photos[0].url ?? `/api/v1/profiles/${profile.profile_id}/photos/${photos[0].id}/file`
            );
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

  const receivedClass =
    receivedLike && receivedFavorite
      ? styles.receivedBoth
      : receivedLike
      ? styles.receivedLike
      : receivedFavorite
      ? styles.receivedFav
      : '';

  const displayGoal =
    profile.relationship_goal ||
    (profile.relationship_goals && profile.relationship_goals.length > 0
      ? profile.relationship_goals[0]
      : null);

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
            <div className={styles.photoPlaceholder}>Sin Foto</div>
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
                {[profile.region, profile.country_code].filter(Boolean).join(', ')}
              </p>
              {displayGoal && (
                <span
                  style={{
                    display: 'inline-block',
                    marginTop: '8px',
                    fontSize: '0.8rem',
                    background: '#f3f4f6',
                    padding: '2px 8px',
                    borderRadius: '12px',
                    color: '#4b5563',
                  }}
                >
                  {displayGoal}
                </span>
              )}
            </div>

            <div style={{ display: 'flex', gap: '8px' }}>
              {profile.has_photo && (
                <button
                  type="button"
                  onClick={(e) => {
                    e.preventDefault();
                    e.stopPropagation();
                    setIsGalleryOpen(true);
                  }}
                  style={{
                    background: '#f3f4f6',
                    border: '1px solid #e5e7eb',
                    borderRadius: '50%',
                    width: '34px',
                    height: '34px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    cursor: 'pointer',
                    fontSize: '1rem',
                    transition: 'all 0.2s',
                  }}
                  title="Ver fotos"
                >
                  📸
                </button>
              )}

              <button
                type="button"
                onClick={handleLike}
                style={{
                  background: liked ? '#22c55e' : '#f3f4f6',
                  border: liked ? 'none' : '1px solid #e5e7eb',
                  boxShadow: liked ? '0 0 10px rgba(34, 197, 94, 0.5)' : 'none',
                  borderRadius: '50%',
                  width: '34px',
                  height: '34px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  fontSize: '1rem',
                  transition: 'all 0.2s',
                }}
                title={liked ? 'Quitar Like' : 'Dar Like'}
              >
                {liked ? '❤️' : '🤍'}
              </button>

              <button
                type="button"
                onClick={handleFavorite}
                style={{
                  background: favorited ? '#eab308' : '#f3f4f6',
                  border: favorited ? 'none' : '1px solid #e5e7eb',
                  boxShadow: favorited ? '0 0 10px rgba(234, 179, 8, 0.5)' : 'none',
                  borderRadius: '50%',
                  width: '34px',
                  height: '34px',
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  cursor: 'pointer',
                  fontSize: '1.1rem',
                  transition: 'all 0.2s',
                }}
                title={favorited ? 'Quitar Favorito' : 'Añadir a Favoritos'}
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
// Componente interno con lógica de filtros y peticiones
// -----------------------------------------------------------------------------
function DiscoverContent() {
  const searchParams = useSearchParams();
  const searchString = searchParams.toString();

  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [lastSearch, setLastSearch] = useState(searchString);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Likes y Favoritos enviados y recibidos
  const [likedIds, setLikedIds] = useState<Set<string>>(new Set());
  const [favoritedIds, setFavoritedIds] = useState<Set<string>>(new Set());
  const [receivedLikeIds, setReceivedLikeIds] = useState<Set<string>>(new Set());
  const [receivedFavIds, setReceivedFavIds] = useState<Set<string>>(new Set());

  const [randomAdIndex, setRandomAdIndex] = useState(-1);

  if (searchString !== lastSearch) {
    setLastSearch(searchString);
    setPage(1);
  }

  // 1. Cargar interacciones sociales del usuario (Likes y Favoritos)
  useEffect(() => {
    apiFetch<LikesResponse>('/likes/sent?page_size=100')
      .then((res) => setLikedIds(new Set(res.items.map((i) => i.profile_id))))
      .catch(() => {});

    apiFetch<FavoritesResponse>('/favorites?page_size=100')
      .then((res) => setFavoritedIds(new Set(res.items.map((i) => i.profile_id))))
      .catch(() => {});

    apiFetch<LikesResponse>('/likes/received?page_size=100')
      .then((res) => setReceivedLikeIds(new Set(res.items.map((i) => i.profile_id))))
      .catch(() => {});

    apiFetch<FavoritesResponse>('/favorites/received?page_size=100')
      .then((res) => setReceivedFavIds(new Set(res.items.map((i) => i.profile_id))))
      .catch(() => {});
  }, []);

  // 2. Cargar perfiles según los filtros y página
  useEffect(() => {
    let isMounted = true;
    setLoading(true);
    setError(null);
    setRandomAdIndex(Math.floor(Math.random() * 20) + 2);

    const params = new URLSearchParams(searchString);
    params.set('page', page.toString());
    params.set('page_size', '24');

    apiFetch<SearchResponse>(`/search/profiles?${params.toString()}`)
      .then((res) => {
        if (isMounted) setData(res);
      })
      .catch((err: unknown) => {
        if (!isMounted) return;
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver perfiles.');
        } else if (err instanceof ApiError && err.status === 429) {
          setError('Demasiadas peticiones. Espera un minuto e inténtalo de nuevo.');
        } else {
          setError('No se pudieron cargar los resultados.');
        }
      })
      .finally(() => {
        if (isMounted) setLoading(false);
      });

    return () => {
      isMounted = false;
    };
  }, [page, searchString]);

  const profiles = data?.items ?? [];
  const totalPages = data?.total_pages ?? 1;

  return (
    <>
      {loading && (
        <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>
          Cargando perfiles…
        </p>
      )}

      {error && <div className={styles.errorBanner}>{error}</div>}

      {!loading && !error && (
        <>
          {profiles.length === 0 ? (
            <div className={styles.emptyState}>
              <p style={{ fontSize: '1.2rem', marginBottom: '1rem' }}>
                No hay resultados que coincidan con tu búsqueda.
              </p>
              <Link
                href="/search"
                style={{
                  color: 'var(--brand-primary)',
                  textDecoration: 'none',
                  fontWeight: 'bold',
                  padding: '0.5rem 1rem',
                  border: '1px solid var(--brand-primary)',
                  borderRadius: '4px',
                }}
              >
                Cambiar filtros
              </Link>
            </div>
          ) : (
            <div className={styles.grid}>
              {profiles.map((profile, index) => {
                const isPremium = index === 0;
                const showAdSquare = index === randomAdIndex;
                const showBannerHorizontal = index === 14;

                const isLiked = likedIds.has(profile.profile_id);
                const isFavorited = favoritedIds.has(profile.profile_id);
                const gotLike = receivedLikeIds.has(profile.profile_id);
                const gotFavorite = receivedFavIds.has(profile.profile_id);

                return (
                  <React.Fragment key={profile.profile_id}>
                    {showAdSquare && (
                      <div className={styles.adSquare}>[ANUNCIO PATROCINADO - ALEATORIO]</div>
                    )}
                    {showBannerHorizontal && (
                      <div className={styles.adBanner}>[ESPACIO PUBLICITARIO - INTERMEDIO]</div>
                    )}

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
          )}

          {totalPages > 1 && (
            <div
              style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                gap: '1rem',
                margin: '3rem 0',
              }}
            >
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                style={{
                  padding: '0.75rem 1.5rem',
                  cursor: page <= 1 ? 'not-allowed' : 'pointer',
                  borderRadius: '8px',
                  border: '1px solid #ccc',
                  background: '#fff',
                  fontWeight: 'bold',
                }}
              >
                ← Anterior
              </button>
              <span style={{ fontWeight: 600, color: '#444' }}>
                Página {data?.page} de {totalPages}
              </span>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => setPage((p) => p + 1)}
                style={{
                  padding: '0.75rem 1.5rem',
                  cursor: page >= totalPages ? 'not-allowed' : 'pointer',
                  borderRadius: '8px',
                  border: '1px solid #ccc',
                  background: '#fff',
                  fontWeight: 'bold',
                }}
              >
                Siguiente →
              </button>
            </div>
          )}
        </>
      )}
    </>
  );
}

// -----------------------------------------------------------------------------
// Export principal de la página Discover
// -----------------------------------------------------------------------------
export default function DiscoverPage() {
  return (
    <main className={styles.main}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '2rem',
        }}
      >
        <h1 style={{ fontSize: '1.8rem', margin: 0, color: '#111827' }}>Descubrir</h1>
        <Link
          href="/search"
          style={{
            textDecoration: 'none',
            background: 'var(--brand-primary)',
            color: 'white',
            padding: '0.6rem 1.2rem',
            borderRadius: '8px',
            fontWeight: 'bold',
            boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          }}
        >
          🔍 Filtros de Búsqueda
        </Link>
      </div>

      <Suspense fallback={<p style={{ textAlign: 'center' }}>Cargando...</p>}>
        <DiscoverContent />
      </Suspense>
    </main>
  );
}
