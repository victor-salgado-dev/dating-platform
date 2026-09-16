'use client';

import React, { useEffect, useState, Suspense } from 'react';
import Link from 'next/link';
import { useSearchParams } from 'next/navigation';

import { apiFetch, ApiError, SearchResponse } from '@/lib/api';
// Si tuvieras los estilos en la raíz, cámbialo a '../page.module.css'. 
// Asumimos que discover tiene su propio page.module.css o usa el mismo.
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

// Componente interno que maneja la lógica de los parámetros de búsqueda
function DiscoverContent() {
  const searchParams = useSearchParams();
  const searchString = searchParams.toString();
  
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [lastSearch, setLastSearch] = useState(searchString);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  if (searchString !== lastSearch) {
    setLastSearch(searchString);
    setPage(1);
  }

  useEffect(() => {
    let isMounted = true;
    setLoading(true);
    setError(null);

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
      {loading && <p style={{ textAlign: 'center', padding: '2rem', fontSize: '1.2rem', color: '#666' }}>Cargando perfiles…</p>}

      {error && <div className={styles.errorBanner} style={{ padding: '1rem', background: '#fee2e2', color: '#991b1b', borderRadius: '8px', textAlign: 'center', margin: '1rem 0' }}>{error}</div>}

      {!loading && !error && (
        <>
          {profiles.length === 0 ? (
            <div className={styles.emptyState} style={{ textAlign: 'center', padding: '4rem 1rem', color: '#666' }}>
              <p style={{ fontSize: '1.2rem', marginBottom: '1rem' }}>No hay resultados que coincidan con tu búsqueda.</p>
              <Link href="/search" style={{ color: 'var(--brand-primary)', textDecoration: 'none', fontWeight: 'bold', padding: '0.5rem 1rem', border: '1px solid var(--brand-primary)', borderRadius: '4px' }}>
                Cambiar filtros
              </Link>
            </div>
          ) : (
            <ul className={styles.grid} style={{ padding: 0, listStyle: 'none' }}>
              {profiles.map((item) => (
                <li key={item.profile_id} className={styles.card}>
                  <Link href={`/profiles/${item.profile_id}`} style={{ textDecoration: 'none', color: 'inherit', display: 'flex', flexDirection: 'column', height: '100%' }}>
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
                        <span style={{ fontWeight: 'normal', color: '#9ca3af' }}> · {item.age}</span>
                      </h3>
                      <p className={styles.details}>
                        {[item.region, item.country_code].filter(Boolean).join(', ')}
                      </p>
                      {item.relationship_goal && (
                        <span style={{ display: 'inline-block', marginTop: '8px', fontSize: '0.8rem', background: '#f3f4f6', padding: '2px 8px', borderRadius: '12px', color: '#4b5563' }}>
                          {item.relationship_goal}
                        </span>
                      )}
                    </div>
                  </Link>
                </li>
              ))}
            </ul>
          )}

          {totalPages > 1 && (
            <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', gap: '1rem', margin: '3rem 0' }}>
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                style={{ padding: '0.75rem 1.5rem', cursor: page <= 1 ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
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
                style={{ padding: '0.75rem 1.5rem', cursor: page >= totalPages ? 'not-allowed' : 'pointer', borderRadius: '8px', border: '1px solid #ccc', background: '#fff', fontWeight: 'bold' }}
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

// ESTE ES EL EXPORT DEFAULT QUE NEXT.JS NECESITA
export default function DiscoverPage() {
  return (
    <main className={styles.main}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h1 style={{ fontSize: '1.8rem', margin: 0, color: '#111827' }}>Descubrir</h1>
        <Link 
          href="/search" 
          style={{ textDecoration: 'none', background: 'var(--brand-primary)', color: 'white', padding: '0.6rem 1.2rem', borderRadius: '8px', fontWeight: 'bold', boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}
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