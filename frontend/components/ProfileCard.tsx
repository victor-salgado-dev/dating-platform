'use client';

// Tarjeta de perfil + modal de fotos, compartidos entre Home, Discover,
// Likes, Visits, Favorites y Activity.
//
// El estado inicial de like / favorito y las insignias de "te dio like" /
// "te tiene en favoritos" vienen en la propia ficha (profile.liked,
// profile.favorited, profile.received_like, profile.received_favorite): el
// backend los calcula en la misma consulta que el listado. Las props
// initialLiked / initialFavorited / receivedLike / receivedFavorite siguen
// aceptándose (y, si se pasan, mandan) para no romper pantallas antiguas que
// todavía usan useProfileInteractions.

import React, { useEffect, useState } from 'react';
import Link from 'next/link';
import { apiFetch } from '@/lib/api';
import { useI18n } from '@/lib/i18n/context';
import { tOption } from '@/lib/i18n/options';
import styles from './ProfileCard.module.css';

export interface ProfileItem {
  profile_id: string;
  display_name: string;
  age: number;
  gender: string;
  country_code: string;
  region: string | null;
  // Opcional: cuando el listado los trae (Home, Discover), se muestra como
  // insignia traducida en vez del código crudo del backend ("casual",
  // "long_term"...).
  relationship_goal?: string | null;
  relationship_goals?: string[] | null;
  has_photo: boolean;
  photo_url?: string | null;

  // Relación con quien mira, calculada por el backend.
  liked?: boolean;
  favorited?: boolean;
  received_like?: boolean;
  received_favorite?: boolean;
}

interface PhotoItem {
  id: string;
  url: string;
  position: number;
}

// -----------------------------------------------------------------------------
// Componente de la Galería Modal
// -----------------------------------------------------------------------------
export function PhotoGalleryModal({ profileId, name, onClose }: { profileId: string; name: string; onClose: () => void }) {
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
export function ProfileCard({
  profile,
  isPremium,
  initialLiked,
  initialFavorited,
  receivedLike,
  receivedFavorite,
  onToggleLike,
  onToggleFavorite,
}: {
  profile: ProfileItem;
  isPremium: boolean;
  // Opcionales: si no se pasan, se usan los flags de `profile`.
  initialLiked?: boolean;
  initialFavorited?: boolean;
  receivedLike?: boolean;
  receivedFavorite?: boolean;
  onToggleLike?: (id: string, liked: boolean) => void;
  onToggleFavorite?: (id: string, favorited: boolean) => void;
}) {
  const { dictionary } = useI18n();
  const [isGalleryOpen, setIsGalleryOpen] = useState(false);

  const likedFromServer = initialLiked ?? profile.liked ?? false;
  const favoritedFromServer = initialFavorited ?? profile.favorited ?? false;
  const [liked, setLiked] = useState(likedFromServer);
  const [favorited, setFavorited] = useState(favoritedFromServer);

  useEffect(() => {
    setLiked(likedFromServer);
  }, [likedFromServer]);

  useEffect(() => {
    setFavorited(favoritedFromServer);
  }, [favoritedFromServer]);

  const gotLike = receivedLike ?? profile.received_like ?? false;
  const gotFavorite = receivedFavorite ?? profile.received_favorite ?? false;

  const handleLike = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const nextState = !liked;
    setLiked(nextState);
    onToggleLike?.(profile.profile_id, nextState);

    try {
      await apiFetch(`/likes/${profile.profile_id}`, { method: nextState ? 'POST' : 'DELETE' });
    } catch (err) {
      setLiked(!nextState);
      onToggleLike?.(profile.profile_id, !nextState);
      console.error('Error al procesar like', err);
    }
  };

  const handleFavorite = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    const nextState = !favorited;
    setFavorited(nextState);
    onToggleFavorite?.(profile.profile_id, nextState);

    try {
      await apiFetch(`/favorites/${profile.profile_id}`, { method: nextState ? 'POST' : 'DELETE' });
    } catch (err) {
      setFavorited(!nextState);
      onToggleFavorite?.(profile.profile_id, !nextState);
      console.error('Error al procesar favorito', err);
    }
  };

  const receivedClass = gotLike && gotFavorite
    ? styles.receivedBoth
    : gotLike
    ? styles.receivedLike
    : gotFavorite
    ? styles.receivedFav
    : '';

  // El backend manda el código crudo (p.ej. "long_term"); se traduce con el
  // diccionario de opciones. Si llega un código que no está mapeado, se
  // muestra tal cual antes que ocultarlo por completo.
  const goalCode =
    profile.relationship_goal ||
    (profile.relationship_goals && profile.relationship_goals.length > 0 ? profile.relationship_goals[0] : null);
  const displayGoal = goalCode ? tOption(dictionary, 'relationshipGoal', goalCode) ?? goalCode : null;

  return (
    <>
      <Link
        href={`/profiles/${profile.profile_id}`}
        className={`${styles.card} ${isPremium ? styles.cardPremium : ''} ${receivedClass}`}
      >
        <div className={styles.imageContainer}>
          {profile.photo_url ? (
            <img
              src={profile.photo_url}
              alt={profile.display_name}
              className={styles.photoImg}
              loading="lazy" // Descarga escalonada: solo descarga la imagen si entra en la vista
            />
          ) : (
            <div className={styles.photoPlaceholder}>{dictionary.common.noPhoto}</div>
          )}
          {(gotLike || gotFavorite) && (
            <div className={styles.receivedBadges}>
              {gotLike && <span className={styles.badgeLike}>♥</span>}
              {gotFavorite && <span className={styles.badgeFav}>★</span>}
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
