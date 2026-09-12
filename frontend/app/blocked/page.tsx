'use client';

import { useEffect, useState } from 'react';

import { apiFetch, ApiError, BlockedResponse } from '@/lib/api';
import styles from '../discover/page.module.css';

export default function BlockedPage() {
  const [data, setData] = useState<BlockedResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [busyId, setBusyId] = useState<string | null>(null);

  function load() {
    setLoading(true);
    setError(null);
    apiFetch<BlockedResponse>('/blocks')
      .then(setData)
      .catch((err: unknown) => {
        if (err instanceof ApiError && err.status === 401) {
          setError('Inicia sesión para ver esta página.');
        } else {
          setError('No se pudo cargar la lista de bloqueados.');
        }
      })
      .finally(() => setLoading(false));
  }

  useEffect(load, []);

  async function unblock(profileId: string) {
    setBusyId(profileId);
    try {
      await apiFetch<void>(`/blocks/${profileId}`, { method: 'DELETE' });
      setData((prev) =>
        prev ? { ...prev, items: prev.items.filter((i) => i.profile_id !== profileId) } : prev,
      );
    } catch {
      setError('No se pudo desbloquear.');
    } finally {
      setBusyId(null);
    }
  }

  return (
    <main className={styles.main}>
      <h1>Perfiles bloqueados</h1>

      {loading && <p>Cargando…</p>}
      {error && <p className={styles.error}>{error}</p>}

      {data && (
        <>
          {data.items.length === 0 ? (
            <p>No has bloqueado a nadie.</p>
          ) : (
            <ul className={styles.grid}>
              {data.items.map((item) => (
                <li key={item.profile_id} className={styles.card}>
                  <div className={styles.cardLink}>
                    <div className={styles.cardName}>
                      {item.display_name}, {item.age}
                    </div>
                    <div className={styles.cardLocation}>
                      {[item.region, item.country_code].filter(Boolean).join(', ')}
                    </div>
                    <button
                      type="button"
                      onClick={() => unblock(item.profile_id)}
                      disabled={busyId === item.profile_id}
                      style={{ marginTop: '0.5rem' }}
                    >
                      Desbloquear
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </main>
  );
}
