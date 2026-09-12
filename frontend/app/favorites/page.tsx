'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, FavoritesResponse } from '@/lib/api';
// Reutiliza los estilos de tarjeta/paginación de /discover: misma forma
// de lista, no tiene sentido duplicar el CSS.
import styles from '../discover/page.module.css';

export default function FavoritesPage() {
  const [data, setData] = useState<FavoritesResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    apiFetch<FavoritesResponse>(`/favorites?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver tus favoritos.');
        } else {
          setError('No se pudieron cargar los favoritos.');
        }
      })
      .finally(() => setLoading(false));
  }, [page]);

  return (
    <main className={styles.main}>
      <h1>Favoritos</h1>

      {loading && <p>Cargando…</p>}
      {error && <p className={styles.error}>{error}</p>}

      {data && (
        <>
          {data.items.length === 0 ? (
            <p>Todavía no has añadido ningún favorito.</p>
          ) : (
            <ul className={styles.grid}>
              {data.items.map((item) => (
                <li key={item.profile_id} className={styles.card}>
                  <Link href={`/profiles/${item.profile_id}`} className={styles.cardLink}>
                    <div className={styles.cardName}>
                      {item.display_name}, {item.age}
                    </div>
                    <div className={styles.cardLocation}>
                      {[item.region, item.country_code].filter(Boolean).join(', ')}
                    </div>
                    {item.relationship_goal && (
                      <div className={styles.cardGoal}>{item.relationship_goal}</div>
                    )}
                  </Link>
                </li>
              ))}
            </ul>
          )}

          {data.total_pages > 1 && (
            <div className={styles.pagination}>
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => setPage((p) => Math.max(1, p - 1))}
              >
                Anterior
              </button>
              <span>
                Página {data.page} de {data.total_pages} ({data.total} favoritos)
              </span>
              <button
                type="button"
                disabled={page >= data.total_pages}
                onClick={() => setPage((p) => p + 1)}
              >
                Siguiente
              </button>
            </div>
          )}
        </>
      )}
    </main>
  );
}
