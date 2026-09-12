'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';

import { apiFetch, ApiError, SearchResponse } from '@/lib/api';
import styles from './page.module.css';

export default function DiscoverPage() {
  const [data, setData] = useState<SearchResponse | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    setError(null);

    apiFetch<SearchResponse>(`/search/profiles?page=${page}`)
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver perfiles.');
        } else {
          setError('No se pudieron cargar los resultados.');
        }
      })
      .finally(() => setLoading(false));
  }, [page]);

  return (
    <main className={styles.main}>
      <h1>Descubrir</h1>

      {loading && <p>Cargando…</p>}
      {error && <p className={styles.error}>{error}</p>}

      {data && (
        <>
          {data.items.length === 0 ? (
            <p>No hay resultados con estos criterios.</p>
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

          <div className={styles.pagination}>
            <button
              type="button"
              disabled={page <= 1}
              onClick={() => setPage((p) => Math.max(1, p - 1))}
            >
              Anterior
            </button>
            <span>
              Página {data.page} de {Math.max(data.total_pages, 1)} ({data.total} perfiles)
            </span>
            <button
              type="button"
              disabled={page >= data.total_pages}
              onClick={() => setPage((p) => p + 1)}
            >
              Siguiente
            </button>
          </div>
        </>
      )}
    </main>
  );
}
