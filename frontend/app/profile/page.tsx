'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, ProfilePhoto, PublicProfile } from '@/lib/api';
import styles from '../profiles/[id]/page.module.css';

export default function MyProfilePage() {
  const [profile, setProfile] = useState<PublicProfile | null>(null);
  const [photos, setPhotos] = useState<ProfilePhoto[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    apiFetch<PublicProfile>('/profiles/me')
      .then(async (data) => {
        setProfile(data);
        try {
          setPhotos(await apiFetch<ProfilePhoto[]>('/profiles/me/photos'));
        } catch {
          setPhotos([]);
        }
      })
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 404) {
          setError('Aún no has creado tu perfil.');
        } else {
          setError('No se pudo cargar tu perfil.');
        }
      })
      .finally(() => setLoading(false));
  }, []);

  return (
    <main className={styles.main}>
      <Link href="/discover" className={styles.back}>
        &larr; Volver a resultados
      </Link>

      {loading && <p>Cargando…</p>}
      {error && (
        <>
          <p className={styles.error}>{error}</p>
          <Link href="/profile/edit">Modificar perfil</Link>
        </>
      )}

      {profile && (
        <article>
          <h1>
            {profile.display_name}, {profile.age}
          </h1>

          <p>
            <Link href="/profile/edit" className={styles.editProfileButton}>
              Modificar perfil
            </Link>
          </p>

          <p className={styles.location}>
            {[profile.region, profile.country_code].filter(Boolean).join(', ')}
          </p>

          <dl className={styles.details}>
            <dt>Género</dt>
            <dd>{profile.gender}</dd>
            <dt>¿Tiene hijos?</dt>
            <dd>{profile.has_children === null ? 'No indicado' : profile.has_children ? 'Sí' : 'No'}</dd>
            <dt>¿Quiere tener hijos?</dt>
            <dd>{profile.wants_children === null ? 'No indicado' : profile.wants_children ? 'Sí' : 'No'}</dd>
          </dl>

          {photos.length > 0 && (
            <div className={styles.photos}>
              {photos.map((photo) => (
                <img key={photo.id} src={photo.url} alt="" className={styles.photo} />
              ))}
            </div>
          )}

          {profile.bio && <p className={styles.bio}>{profile.bio}</p>}

          <dl className={styles.details}>
            {profile.relationship_goal && (
              <>
                <dt>Busca</dt>
                <dd>{profile.relationship_goal}</dd>
              </>
            )}
            {profile.languages && profile.languages.length > 0 && (
              <>
                <dt>Idiomas</dt>
                <dd>{profile.languages.join(', ')}</dd>
              </>
            )}
            {profile.interests && profile.interests.length > 0 && (
              <>
                <dt>Intereses</dt>
                <dd>{profile.interests.join(', ')}</dd>
              </>
            )}
          </dl>
        </article>
      )}
    </main>
  );
}